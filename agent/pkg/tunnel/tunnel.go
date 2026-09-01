package tunnel

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v5"
	"github.com/multiformats/go-multistream"
	"github.com/shellhub-io/shellhub/pkg/api/client"
	log "github.com/sirupsen/logrus"
)

// HandlerConstraint is the handler shape each transport version accepts: V1 routes HTTP
// requests through echo, V2 hands the raw stream to the handler.
type HandlerConstraint interface {
	echo.HandlerFunc | HandlerFunc
}

// Tunnel accepts the connections a ShellHub server opens back to the device and dispatches
// each to the handler registered for its protocol.
type Tunnel[H HandlerConstraint] interface {
	Handle(protocol string, handler H)
	Listen(ctx context.Context, listener net.Listener) error
	Close() error
}

// TunnelV2 multiplexes streams over the reverse connection, selecting a handler by the
// protocol name the peer negotiates.
type TunnelV2 struct {
	mux      *multistream.MultistreamMuxer[string]
	cli      client.Client
	listener net.Listener
}

// NewTunnelV2 returns a tunnel that uses cli to re-open the reverse connection when it drops.
func NewTunnelV2(cli client.Client) Tunnel[HandlerFunc] {
	return &TunnelV2{
		mux: multistream.NewMultistreamMuxer[string](),
		cli: cli,
	}
}

// Handle registers handler as the one serving streams that negotiate protocol.
func (t *TunnelV2) Handle(protocol string, handler HandlerFunc) {
	t.mux.AddHandler(protocol, func(protocol string, rwc io.ReadWriteCloser) error {
		log.WithField("protocol", protocol).Debug("handling connection")
		defer log.WithField("protocol", protocol).Debug("handling connection closed")

		return handler(NewContext(context.TODO(), rwc), rwc)
	})
}

// Listen accepts streams from listener until ctx is cancelled, serving each in its own
// goroutine. It reports [ErrTunnelDisconnect] when the reverse connection is lost.
func (t *TunnelV2) Listen(ctx context.Context, listener net.Listener) error {
	t.listener = listener

	for {
		stream, err := listener.Accept()
		if err != nil {
			log.WithError(err).Trace("failed to accept stream")

			if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
				return errors.Join(ErrTunnelDisconnect, err)
			}

			return err
		}

		log.Trace("new stream accepted")

		go func() {
			log.Trace("handling stream")

			if err := t.mux.Handle(stream); err != nil {
				log.WithError(err).Trace("failed to handle stream")

				_ = stream.Close()
			}

			log.Trace("stream handled")
		}()
	}
}

// Close implements Tunnel.
func (t *TunnelV2) Close() error {
	return t.listener.Close()
}

// ErrTunnelDisconnect is returned when the tunnel connection is closed.
var ErrTunnelDisconnect = errors.New("tunnel disconnected")

// TunnelV1 serves the reverse connection as HTTP, routing by method and path.
type TunnelV1 struct {
	router *echo.Echo
	srv    *http.Server
}

// NewTunnelV1 returns a tunnel serving the legacy HTTP transport.
func NewTunnelV1() *TunnelV1 {
	e := echo.New()

	t := &TunnelV1{
		router: e,
		srv: &http.Server{ //nolint:gosec
			Handler: e,
			ConnContext: func(ctx context.Context, c net.Conn) context.Context {
				return context.WithValue(ctx, "http-conn", c)
			},
		},
	}

	return t
}

// Handle registers handler for protocol, written as "METHOD://path".
func (t *TunnelV1) Handle(protocol string, handler echo.HandlerFunc) {
	parts := strings.SplitN(protocol, "://", 2)

	method := parts[0]
	path := parts[1]

	t.router.Add(method, path, func(c *echo.Context) error {
		log.WithField("protocol", protocol).Debug("handling connection")
		defer log.WithField("protocol", protocol).Debug("handling connection closed")

		return handler(c)
	})
}

// Listen serves HTTP over listener until the listener is closed.
func (t *TunnelV1) Listen(ctx context.Context, listener net.Listener) error {
	return t.srv.Serve(listener)
}

// Close stops serving and drops the connections still open.
func (t *TunnelV1) Close() error {
	return t.srv.Close()
}
