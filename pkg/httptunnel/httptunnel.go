package httptunnel

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/http"

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
	CheckOrigin: func(r *http.Request) bool {
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
		ConnectionHandler: func(r *http.Request) (string, error) {
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

		id, err := t.ConnectionHandler(c.Request())
		if err != nil {
			conn.Close()

			return c.String(http.StatusBadRequest, err.Error())
		}

		t.connman.Set(id, wsconnadapter.New(conn), t.DialerPath)

		return nil
	})

	e.GET(t.DialerPath, echo.WrapHandler(revdial.ConnHandler(upgrader)))

	return e
}

func (t *Tunnel) Dial(ctx context.Context, id string) (net.Conn, error) {
	return t.connman.Dial(ctx, id)
}

func (t *Tunnel) SendRequest(ctx context.Context, id string, req *http.Request) (*http.Response, error) {
	conn, err := t.connman.Dial(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := req.Write(conn); err != nil {
		return nil, err
	}

	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (t *Tunnel) ForwardResponse(resp *http.Response, w http.ResponseWriter) {
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body) // nolint:errcheck
	resp.Body.Close()
}
