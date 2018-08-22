package filemanager

import (
	"net/http"

	"github.com/filebrowser/caddy/parser"
	l "github.com/filebrowser/filebrowser/lib"
	h "github.com/filebrowser/filebrowser/lib/http"
	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

func init() {
	caddy.RegisterPlugin("filemanager", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

type plugin struct {
	Next    httpserver.Handler
	Configs []*l.FileBrowser
}

// ServeHTTP determines if the request is for this plugin, and if all prerequisites are met.
func (f plugin) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	for i := range f.Configs {
		// Checks if this Path should be handled by File Manager.
		if !httpserver.Path(r.URL.Path).Matches(f.Configs[i].BaseURL) {
			continue
		}

		h.Handler(f.Configs[i]).ServeHTTP(w, r)
		return 0, nil
	}

	return f.Next.ServeHTTP(w, r)
}

// setup configures a new FileManager middleware instance.
func setup(c *caddy.Controller) error {
	configs, err := parser.Parse(c, "")
	if err != nil {
		return err
	}

	httpserver.GetConfig(c).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		return plugin{Configs: configs, Next: next}
	})

	return nil
}
