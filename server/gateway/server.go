package gateway

import (
	"bufio"
	"errors"
	"net"
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

// ErrNoServerHandler is what Start returns for a nil handler, and what loading a configuration
// that uses shellhub_server returns outside Start, when the routes would have nothing to serve
// them.
var ErrNoServerHandler = errors.New("the gateway has no server handler")

var serverHandler http.Handler

func init() {
	caddy.RegisterModule(new(ServerHandler))
	httpcaddyfile.RegisterHandlerDirective("shellhub_server", func(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
		handler := new(ServerHandler)

		return handler, handler.UnmarshalCaddyfile(h.Dispenser)
	})
}

// ServerHandler is the shellhub_server Caddy handler. It ends the route by serving the request
// with the handler given to Start, so the server answers in the gateway's process instead of
// behind a proxied connection.
type ServerHandler struct {
	handler http.Handler
}

// CaddyModule registers the handler as http.handlers.shellhub_server, the module the
// shellhub_server Caddyfile directive names.
func (*ServerHandler) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.shellhub_server",
		New: func() caddy.Module { return new(ServerHandler) },
	}
}

// Provision binds the handler given to the Start that is loading this configuration, and returns
// ErrNoServerHandler when the configuration is loaded some other way.
func (s *ServerHandler) Provision(caddy.Context) error {
	if serverHandler == nil {
		return ErrNoServerHandler
	}

	s.handler = serverHandler

	return nil
}

// UnmarshalCaddyfile accepts the bare directive, and returns an argument error for anything after
// it.
func (s *ServerHandler) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next()

	if d.NextArg() {
		return d.ArgErr()
	}

	return nil
}

// ServeHTTP serves the request in process and returns nil. It never calls next, so it must close
// its route. The writer it hands on implements http.Hijacker and http.Flusher.
func (s *ServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, _ caddyhttp.Handler) error {
	s.handler.ServeHTTP(hijackable{w}, r)

	return nil
}

type hijackable struct {
	http.ResponseWriter
}

func (h hijackable) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return http.NewResponseController(h.ResponseWriter).Hijack()
}

func (h hijackable) Flush() {
	http.NewResponseController(h.ResponseWriter).Flush() //nolint:errcheck // http.Flusher has no error to report it through
}

func (h hijackable) Unwrap() http.ResponseWriter {
	return h.ResponseWriter
}

var (
	_ caddy.Provisioner           = (*ServerHandler)(nil)
	_ caddyhttp.MiddlewareHandler = (*ServerHandler)(nil)
	_ caddyfile.Unmarshaler       = (*ServerHandler)(nil)
)
