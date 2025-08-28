package httptunnel

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/pkg/connman"
	"github.com/shellhub-io/shellhub/pkg/revdial"
	"github.com/shellhub-io/shellhub/pkg/wsconnadapter"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	Subprotocols:    []string{"binary"},
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

const (
	DefaultConnectionURL = "/connection"
	DefaultRevdialURL    = "/revdial"
)

type Tunnel struct {
	ConnectionPath    string
	DialerPath        string
	ConnectionHandler func(*http.Request) (string, error)
	CloseHandler      func(string)
	KeepAliveHandler  func(string)
	connman           *connman.ConnectionManager
	id                chan string
	online            chan bool
}

func NewTunnel(connectionPath, dialerPath string) *Tunnel {
	tunnel := &Tunnel{
		ConnectionPath: connectionPath,
		DialerPath:     dialerPath,
		ConnectionHandler: func(_ *http.Request) (string, error) {
			panic("ConnectionHandler not implemented")
		},
		CloseHandler: func(string) {
		},
		KeepAliveHandler: func(string) {
		},
		connman: connman.New(),
		id:      make(chan string),
		online:  make(chan bool),
	}

	tunnel.connman.DialerDoneCallback = func(id string, _ *revdial.Dialer) {
		tunnel.CloseHandler(id)
	}

	tunnel.connman.DialerKeepAliveCallback = func(id string, _ *revdial.Dialer) {
		tunnel.KeepAliveHandler(id)
	}

	return tunnel
}

func (t *Tunnel) Router() http.Handler {
	e := echo.New()

	e.GET(t.ConnectionPath, func(c echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		key, err := t.ConnectionHandler(c.Request())
		if err != nil {
			conn.Close()

			return c.String(http.StatusBadRequest, err.Error())
		}

		requestID := c.Request().Header.Get("X-Request-ID")
		parts := strings.Split(key, ":")
		tenant := parts[0]
		device := parts[1]

		t.connman.Set(
			key,
			wsconnadapter.
				New(conn).
				WithID(requestID).
				WithDevice(tenant, device),
			t.DialerPath,
		)

		return nil
	})

	e.GET(t.DialerPath, echo.WrapHandler(revdial.ConnHandler(upgrader)))

	return e
}

func (t *Tunnel) Dial(ctx context.Context, id string) (net.Conn, error) {
	return t.connman.Dial(ctx, id)
}
