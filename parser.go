package caddy

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/asdine/storm"
	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/http"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/storage/bolt"
	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

var databases = map[string]*storm.DB{}

func parse(c *caddy.Controller) (*handler, error) {
	values := map[string]string{
		"baseURL":          "/",
		"root":             ".",
		"database":         "",
		"auth_method":      "json",
		"auth_header":      "",
		"recaptcha_host":   "",
		"recaptcha_key":    "",
		"recaptcha_secret": "",
	}

	args := c.RemainingArgs()

	if len(args) >= 1 {
		values["baseURL"] = args[0]
	}

	if len(args) > 1 {
		values["root"] = args[1]
	}

	for c.NextBlock() {
		switch val := c.Val(); val {
		case "database",
			"auth_method",
			"auth_header",
			"recaptcha_host",
			"recaptcha_key",
			"recaptcha_secret":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}

			values[val] = c.Val()
		default:
			return nil, c.ArgErr()
		}
	}

	cfg := httpserver.GetConfig(c)
	if cfg.Addr.Path != "" {
		values["baseURL"] = path.Join(cfg.Addr.Path, values["baseURL"])
	}

	err := parseDatabasePath(c, values)
	if err != nil {
		return nil, err
	}

	ser := &settings.Server{
		Root:    values["root"],
		BaseURL: values["baseURL"],
	}

	ser.Root, err = filepath.Abs(ser.Root)
	if err != nil {
		return nil, err
	}

	var (
		db *storm.DB
		ok bool
	)

	if db, ok = databases[values["database"]]; !ok {
		db, err = storm.Open(values["database"])
		if err != nil {
			return nil, err
		}

		databases[values["database"]] = db
	}

	sto, err := bolt.NewStorage(db)
	if err != nil {
		return nil, err
	}

	set, err := sto.Settings.Get()
	if err == errors.ErrNotExist {
		err = quickSetup(sto, values)
		if err != nil {
			return nil, err
		}

		set, err = sto.Settings.Get()
	}

	if err != nil {
		return nil, err
	}

	var auther auth.Auther

	switch settings.AuthMethod(values["auth_method"]) {
	case auth.MethodJSONAuth:
		set.AuthMethod = auth.MethodJSONAuth
		auther = &auth.JSONAuth{
			ReCaptcha: &auth.ReCaptcha{
				Host:   values["recaptcha_host"],
				Key:    values["recaptcha_key"],
				Secret: values["recaptcha_secret"],
			},
		}
	case auth.MethodNoAuth:
		set.AuthMethod = auth.MethodNoAuth
		auther = &auth.NoAuth{}
	case auth.MethodProxyAuth:
		set.AuthMethod = auth.MethodProxyAuth
		header := values["auth_header"]
		if header == "" {
			return nil, c.ArgErr()
		}
		auther = &auth.ProxyAuth{Header: header}
	default:
		return nil, c.ArgErr()
	}

	err = sto.Settings.Save(set)
	if err != nil {
		return nil, err
	}

	err = sto.Settings.SaveServer(ser)
	if err != nil {
		return nil, err
	}

	err = sto.Auth.Save(auther)
	if err != nil {
		return nil, err
	}

	httpHandler, err := http.NewHandler(sto, ser)
	if err != nil {
		return nil, err
	}

	return &handler{
		Handler: httpHandler,
		baseURL: values["baseURL"],
	}, nil
}

func quickSetup(sto *storage.Storage, values map[string]string) error {
	key, err := settings.GenerateKey()
	if err != nil {
		return err
	}

	set := &settings.Settings{
		Key:    key,
		Signup: false,
		Defaults: settings.UserDefaults{
			Scope:  ".",
			Locale: "en",
			Perm: users.Permissions{
				Admin:    false,
				Execute:  true,
				Create:   true,
				Rename:   true,
				Modify:   true,
				Delete:   true,
				Share:    true,
				Download: true,
			},
		},
	}

	err = sto.Settings.Save(set)
	if err != nil {
		return err
	}

	password, err := users.HashPwd("admin")
	if err != nil {
		return err
	}

	user := &users.User{
		Username:     "admin",
		Password:     password,
		LockPassword: false,
	}

	set.Defaults.Apply(user)
	user.Perm.Admin = true

	return sto.Users.Save(user)
}

var warningTemplate = `[WARNING] A database is going to be created for your File Browser
instance at the following configuration:

Host: 		%s
Path: 		%s
BaseURL:	%s

It is highly recommended that you set the 'database' option to "%s.db".`

func parseDatabasePath(c *caddy.Controller, values map[string]string) error {
	cfg := httpserver.GetConfig(c)
	assets := filepath.Join(caddy.AssetsPath(), "filebrowser")
	err := os.MkdirAll(assets, 0700)
	if err != nil {
		return err
	}

	database := values["database"]

	if !filepath.IsAbs(database) && database != "" {
		database = filepath.Join(assets, database)
	}

	if database == "" {
		hasher := md5.New()
		hasher.Write([]byte(cfg.Addr.Host + cfg.Addr.Path + values["baseURL"]))
		sha := hex.EncodeToString(hasher.Sum(nil))
		database = filepath.Join(assets, sha+".db")
		fmt.Printf(warningTemplate+"\n", cfg.Addr.Host, cfg.Addr.Path, values["baseURL"], sha)
	}

	values["database"] = database
	return nil
}
