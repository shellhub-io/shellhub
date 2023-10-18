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

const (
	DefaultConnectionURL = "/connection"
	DefaultRevdialURL    = "/revdial"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	Subprotocols:    []string{"binary"},
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Tunneler interface {
	// ConnectionPath returns the connection path of the tunnel.
	ConnectionPath() string

	// DialerPath returns the dialer path of the tunnel.
	DialerPath() string

	// SetConnectionHandler sets the connection handler of the tunnel.
	SetConnectionHandler(callback func(*http.Request) (string, error))

	// SetCloseHandler sets the close handler of the tunnel.
	SetCloseHandler(callback func(string))

	// SetKeepAliveHandler sets the keep alive handler of the tunnel.
	SetKeepAliveHandler(callback func(string))

	Router() http.Handler

	Dial(ctx context.Context, id string) (net.Conn, error)

	SendRequest(ctx context.Context, id string, req *http.Request) (*http.Response, error)

	ForwardResponse(resp *http.Response, w http.ResponseWriter)
}

type tunnel struct {
	connectionPath    string
	dialerPath        string
	connectionHandler func(*http.Request) (string, error) // usado fora
	closeHandler      func(string)                        // usado fora
	keepAliveHandler  func(string)                        // usado fora
	connman           *connman.ConnectionManager
	id                chan string
	online            chan bool
}

// Ensures that tunnel implements Tunneler
var _ Tunneler = (*tunnel)(nil)

func NewTunnel(connectionPath, dialerPath string) Tunneler {
	tunnel := &tunnel{
		connectionPath: connectionPath,
		dialerPath:     dialerPath,
		connectionHandler: func(r *http.Request) (string, error) {
			panic("ConnectionHandler not implemented")
		},
		closeHandler: func(string) {
		},
		keepAliveHandler: func(string) {
		},
		connman: connman.New(),
		id:      make(chan string),
		online:  make(chan bool),
	}

	tunnel.connman.DialerDoneCallback = func(id string, _ *revdial.Dialer) {
		tunnel.closeHandler(id)
	}

	tunnel.connman.DialerKeepAliveCallback = func(id string, _ *revdial.Dialer) {
		tunnel.keepAliveHandler(id)
	}

	return tunnel
}

func (t *tunnel) ConnectionPath() string { return t.connectionPath }
func (t *tunnel) DialerPath() string     { return t.dialerPath }

func (t *tunnel) SetConnectionHandler(callback func(*http.Request) (string, error)) {
	t.connectionHandler = callback
}

func (t *tunnel) SetCloseHandler(callback func(string)) {
	t.closeHandler = callback
}

func (t *tunnel) SetKeepAliveHandler(callback func(string)) {
	t.keepAliveHandler = callback
}

func (t *tunnel) Router() http.Handler {
	e := echo.New()

	e.GET(t.connectionPath, func(c echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		id, err := t.connectionHandler(c.Request())
		if err != nil {
			conn.Close()

			return c.String(http.StatusBadRequest, err.Error())
		}

		t.connman.Set(id, wsconnadapter.New(conn))

		return nil
	})

	e.GET(t.dialerPath, echo.WrapHandler(revdial.ConnHandler(upgrader)))

	return e
}

func (t *tunnel) Dial(ctx context.Context, id string) (net.Conn, error) {
	return t.connman.Dial(ctx, id)
}

func (t *tunnel) SendRequest(ctx context.Context, id string, req *http.Request) (*http.Response, error) {
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

func (t *tunnel) ForwardResponse(resp *http.Response, w http.ResponseWriter) {
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body) // nolint:errcheck
	resp.Body.Close()
}
