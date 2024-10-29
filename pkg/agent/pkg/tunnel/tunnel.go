package tunnel

import (
	"context"
	"net"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/pkg/revdial"
)

type Tunnel struct {
	router       *echo.Echo
	srv          *http.Server
	ProxyHandler func(e echo.Context) error
	ConnHandler  func(e echo.Context) error
	CloseHandler func(e echo.Context) error
}

type Builder struct {
	tunnel *Tunnel
}

func NewBuilder() *Builder {
	return &Builder{
		tunnel: NewTunnel(),
	}
}

func (t *Builder) WithProxyHandler(handler func(e echo.Context) error) *Builder {
	t.tunnel.ProxyHandler = handler

	return t
}

func (t *Builder) WithConnHandler(handler func(e echo.Context) error) *Builder {
	t.tunnel.ConnHandler = handler

	return t
}

func (t *Builder) WithCloseHandler(handler func(e echo.Context) error) *Builder {
	t.tunnel.CloseHandler = handler

	return t
}

func (t *Builder) Build() *Tunnel {
	return t.tunnel
}

func NewTunnel() *Tunnel {
	e := echo.New()

	t := &Tunnel{
		router: e,
		srv: &http.Server{
			Handler: e,
			ConnContext: func(ctx context.Context, c net.Conn) context.Context {
				return context.WithValue(ctx, "http-conn", c) //nolint:revive
			},
		},
		ConnHandler: func(e echo.Context) error {
			panic("ConnHandler can not be nil")
		},
		CloseHandler: func(e echo.Context) error {
			panic("CloseHandler can not be nil")
		},
		ProxyHandler: func(e echo.Context) error {
			panic("ProxyHandler can not be nil")
		},
	}
	e.GET("/ssh/:id", func(e echo.Context) error {
		return t.ConnHandler(e)
	})
	e.GET("/ssh/close/:id", func(e echo.Context) error {
		return t.CloseHandler(e)
	})
	e.CONNECT("/ssh/proxy/:addr", func(e echo.Context) error {
		// NOTE: The CONNECT HTTP method requests that a proxy establish a HTTP tunnel to this server, and if
		// successful, blindly forward data in both directions until the tunnel is closed.
		//
		// https://en.wikipedia.org/wiki/HTTP_tunnel
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/CONNECT
		return t.ProxyHandler(e)
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
