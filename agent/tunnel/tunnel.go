package tunnel

import (
	"context"
	"net"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/pkg/revdial"
)

type Tunnel struct {
	router *echo.Echo
	server *http.Server

	SSHHandler      func(e echo.Context) error
	SSHCloseHandler func(e echo.Context) error

	SSHContainerHandler func(e echo.Context) error
}

// TunnelContextKey is the type used for context keys in the tunnel package.
type TunnelContextKey string

// TunnelContextKeyHTTPConn is the context key used to store the net.Conn in the HTTP request context.
const TunnelContextKeyHTTPConn TunnelContextKey = "http-conn"

func NewTunnel() *Tunnel {
	e := echo.New()

	t := &Tunnel{
		router: e,
		server: &http.Server{ //nolint:gosec
			Handler: e,
			ConnContext: func(ctx context.Context, conn net.Conn) context.Context {
				return context.WithValue(ctx, TunnelContextKeyHTTPConn, conn)
			},
		},
		SSHHandler: func(_ echo.Context) error {
			panic("ConnHandler can not be nil")
		},
		SSHCloseHandler: func(_ echo.Context) error {
			panic("CloseHandler can not be nil")
		},
		SSHContainerHandler: func(_ echo.Context) error {
			panic("SSHContainerHandler can not be nil")
		},
	}

	e.GET("/ssh/:id", func(e echo.Context) error {
		return t.SSHHandler(e)
	})
	e.GET("/ssh/close/:id", func(e echo.Context) error {
		return t.SSHCloseHandler(e)
	})

	WebEndpointsHandler(e)

	e.GET("/ssh/container/:session/:container", func(e echo.Context) error {
		return t.SSHContainerHandler(e)
	})

	ContainersHandler(e)

	return t
}

// Listen to reverse listener.
func (t *Tunnel) Listen(l *revdial.Listener) error {
	return t.server.Serve(l)
}

// Close closes the tunnel.
func (t *Tunnel) Close() error {
	if err := t.router.Close(); err != nil {
		return err
	}

	return t.server.Close()
}
