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

type ConnectionHandler func(*http.Request) (*connman.Info, error)
type KeepAliveHandler func(context.Context, *connman.Info)
type CloseHandler func(context.Context, *connman.Info)

type Tunnel struct {
	id             chan string
	connman        *connman.ConnectionManager
	online         chan bool
	ConnectionPath string
	DialerPath     string

	// ConnectionHandler is a callback function called when an agent initiates a new connection through the ShellHub server.
	// It receives a request from the agent and should return a string containing sufficient information to identify
	// the connection in subsequent callbacks, or an error if any.
	//
	// TODO: Consider returning a struct containing the information instead of a formatted string.
	ConnectionHandler ConnectionHandler

	// CloseHandler is a callback function called when an agent requests to end a connection.
	CloseHandler CloseHandler

	// KeepAliveHandler is a callback function called to handle keep-alive pings from agents to maintain connection
	// stability. This function may perform any necessary actions to ensure the connection remains active.
	//
	// TODO: Currently, it receives the formatted string returned from [Tunnel.ConnectionHandler]. Consider receive
	// a struct instead.
	KeepAliveHandler KeepAliveHandler
}

func NewTunnel(connectionPath, dialerPath string) *Tunnel {
	tunnel := &Tunnel{
		id:             make(chan string),
		connman:        connman.New(),
		online:         make(chan bool),
		ConnectionPath: connectionPath,
		DialerPath:     dialerPath,
		ConnectionHandler: func(r *http.Request) (*connman.Info, error) {
			panic("ConnectionHandler not yet implemented.")
		},
		CloseHandler: func(_ context.Context, _ *connman.Info) {
			panic("CloseHandler not yet implemented.")
		},
		KeepAliveHandler: func(_ context.Context, _ *connman.Info) {
			panic("KeepAliveHandler not yet implemented.")
		},
	}

	tunnel.connman.DialerDoneCallback = func(ctx context.Context, info *connman.Info, _ *revdial.Dialer) {
		tunnel.CloseHandler(ctx, info)
	}

	tunnel.connman.DialerKeepAliveCallback = func(ctx context.Context, info *connman.Info, _ *revdial.Dialer) {
		tunnel.KeepAliveHandler(ctx, info)
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

		info, err := t.ConnectionHandler(c.Request())
		if err != nil {
			conn.Close()

			return c.String(http.StatusBadRequest, err.Error())
		}

		ctx := context.Background()
		t.connman.Set(ctx, info, wsconnadapter.New(conn))

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
