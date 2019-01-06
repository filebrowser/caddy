package caddy

import (
	"net/http"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

func init() {
	caddy.RegisterPlugin("filebrowser", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

type handler struct {
	baseURL string
	http.Handler
}

type plugin struct {
	next     httpserver.Handler
	handlers []*handler
}

func (p plugin) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	for _, h := range p.handlers {
		if !httpserver.Path(r.URL.Path).Matches(h.baseURL) {
			continue
		}

		h.ServeHTTP(w, r)
		return 0, nil
	}

	return p.next.ServeHTTP(w, r)
}

func setup(c *caddy.Controller) error {
	handlers := []*handler{}

	for c.Next() {
		handler, err := parse(c)
		if err != nil {
			return err
		}
		handlers = append(handlers, handler)
	}

	httpserver.GetConfig(c).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		return plugin{
			handlers: handlers,
			next:     next,
		}
	})

	return nil
}
