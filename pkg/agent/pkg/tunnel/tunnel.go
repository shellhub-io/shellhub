package tunnel

import (
	"context"
	"net"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/pkg/revdial"
)

type Tunnel struct {
	router           *echo.Echo
	srv              *http.Server
	HTTPProxyHandler func(e echo.Context) error
	SSHHandler       func(e echo.Context) error
	SSHCloseHandler  func(e echo.Context) error
}

type Builder struct {
	tunnel *Tunnel
}

func NewBuilder() *Builder {
	return &Builder{
		tunnel: NewTunnel(),
	}
}

func (t *Builder) WithHTTPProxyHandler(handler func(e echo.Context) error) *Builder {
	t.tunnel.HTTPProxyHandler = handler

	return t
}

func (t *Builder) WithSSHHandler(handler func(e echo.Context) error) *Builder {
	t.tunnel.SSHHandler = handler

	return t
}

func (t *Builder) WithSSHCloseHandler(handler func(e echo.Context) error) *Builder {
	t.tunnel.SSHCloseHandler = handler

	return t
}

func (t *Builder) Build() *Tunnel {
	return t.tunnel
}

func NewTunnel() *Tunnel {
	e := echo.New()

	t := &Tunnel{
		router: e,
		srv: &http.Server{ //nolint:gosec
			Handler: e,
			ConnContext: func(ctx context.Context, c net.Conn) context.Context {
				return context.WithValue(ctx, "http-conn", c) //nolint:revive
			},
		},
		SSHHandler: func(_ echo.Context) error {
			panic("ConnHandler can not be nil")
		},
		SSHCloseHandler: func(_ echo.Context) error {
			panic("CloseHandler can not be nil")
		},
		HTTPProxyHandler: func(_ echo.Context) error {
			panic("ProxyHandler can not be nil")
		},
	}
	e.GET("/ssh/:id", func(e echo.Context) error {
		return t.SSHHandler(e)
	})
	e.GET("/ssh/close/:id", func(e echo.Context) error {
		return t.SSHCloseHandler(e)
	})
	e.CONNECT("/http/proxy/:addr", func(e echo.Context) error {
		// NOTE: The CONNECT HTTP method requests that a proxy establish a HTTP tunnel to this server, and if
		// successful, blindly forward data in both directions until the tunnel is closed.
		//
		// https://en.wikipedia.org/wiki/HTTP_tunnel
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/CONNECT
		return t.HTTPProxyHandler(e)
	})

	return t
}

// Listen to reverse listener.
func (t *Tunnel) Listen(l *revdial.Listener) error {
	return t.srv.Serve(l)
}

// Close closes the tunnel.
func (t *Tunnel) Close() error {
	if err := t.router.Close(); err != nil {
		return err
	}

	return t.srv.Close()
}
