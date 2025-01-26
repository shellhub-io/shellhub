package tunnel

import (
	"context"
	"net"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/pkg/revdial"
)

const HTTPConnContextKey = "http-conn"

type Tunnel struct {
	router *echo.Echo
	server *http.Server
}

func NewTunnel() *Tunnel {
	router := echo.New()

	tunnel := &Tunnel{
		router: router,
		server: &http.Server{ //nolint:gosec
			Handler: router,
			ConnContext: func(ctx context.Context, connection net.Conn) context.Context {
				return context.WithValue(ctx, HTTPConnContextKey, connection) //nolint:revive
			},
		},
	}

	return tunnel
}

func (t *Tunnel) Register(handler Module) {
	group := t.router.Group(handler.Prefix())

	handler.Register(group)
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
