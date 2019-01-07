package caddy

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

func parse(c *caddy.Controller) (*handler, error) {
	values := map[string]string{
		"baseURL":  "/",
		"scope":    ".",
		"database": "",
	}

	args := c.RemainingArgs()

	if len(args) >= 1 {
		values["baseURL"] = args[0]
	}

	if len(args) > 1 {
		values["scope"] = args[1]
	}

	for c.NextBlock() {
		switch val := c.Val(); val {
		case "database",
			"locale",
			"auth_method",
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

	err := parseDatabasePath(c, values)
	if err != nil {
		return nil, err
	}

	/*


			u.Scope = scope
			u.FileSystem = fileutils.Dir(scope)

			var db *storm.DB
			if stored, ok := databases[database]; ok {
				db = stored
			} else {
				db, err = storm.Open(database)
				databases[database] = db
			}

			if err != nil {
				return nil, err
			}

			authMethod := "default"
			if noAuth {
				authMethod = "none"
			}

			m := &l.FileBrowser{
				BaseURL:     "",
				PrefixURL:   "",
				DefaultUser: u,
				Auth: &l.Auth{
					Method: authMethod,
				},
				ReCaptcha: &l.ReCaptcha{
					Host:   reCaptchaHost,
					Key:    reCaptchaKey,
					Secret: reCaptchaSecret,
				},
				Store: &l.Store{
					Config: bolt.ConfigStore{DB: db},
					Users:  bolt.UsersStore{DB: db},
					Share:  bolt.ShareStore{DB: db},
				},
				NewFS: func(scope string) l.FileSystem {
					return fileutils.Dir(scope)
				},
			}

			err = m.Setup()
			if err != nil {
				return nil, err
			}



			if err != nil {
				return nil, err
			}

			m.SetBaseURL(baseURL)
			m.SetPrefixURL(strings.TrimSuffix(caddyConf.Addr.Path, "/"))


	} */

	return nil, nil
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
		fmt.Printf(warningTemplate, cfg.Addr.Host, cfg.Addr.Path, values["baseURL"], sha)
	}

	values["database"] = database
	return nil
}
