package tunnel

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/multiformats/go-multistream"
	"github.com/shellhub-io/shellhub/pkg/api/client"
	log "github.com/sirupsen/logrus"
)

type HandlerConstraint interface {
	echo.HandlerFunc | HandlerFunc
}

type Tunnel[H HandlerConstraint] interface {
	Handle(protocol string, handler H)
	Listen(ctx context.Context, listener net.Listener) error
	Close() error
}

type TunnelV2 struct {
	mux      *multistream.MultistreamMuxer[string]
	cli      client.Client
	listener net.Listener
}

func NewTunnelV2(cli client.Client) Tunnel[HandlerFunc] {
	return &TunnelV2{
		mux: multistream.NewMultistreamMuxer[string](),
		cli: cli,
	}
}

func (t *TunnelV2) Handle(protocol string, handler HandlerFunc) {
	t.mux.AddHandler(protocol, func(protocol string, rwc io.ReadWriteCloser) error {
		log.WithField("protocol", protocol).Debug("handling connection")
		defer log.WithField("protocol", protocol).Debug("handling connection closed")

		// TODO: Should we receive a context from outside?
		return handler(NewContext(context.TODO(), rwc), rwc)
	})
}

func (t *TunnelV2) Listen(ctx context.Context, listener net.Listener) error {
	t.listener = listener

	for {
		stream, err := listener.Accept()
		if err != nil {
			log.WithError(err).Trace("failed to accept stream")

			switch {
			case websocket.IsCloseError(err, websocket.CloseAbnormalClosure):
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

type TunnelV1 struct {
	router *echo.Echo
	srv    *http.Server
}

func NewTunnelV1() *TunnelV1 {
	e := echo.New()

	t := &TunnelV1{
		router: e,
		srv: &http.Server{ //nolint:gosec
			Handler: e,
			ConnContext: func(ctx context.Context, c net.Conn) context.Context {
				// TODO: Create a constant for the key.
				return context.WithValue(ctx, "http-conn", c) //nolint:revive
			},
		},
	}

	return t
}

func (t *TunnelV1) Handle(protocol string, handler echo.HandlerFunc) {
	parts := strings.SplitN(protocol, "://", 2)

	method := parts[0]
	path := parts[1]

	t.router.Add(method, path, func(c echo.Context) error {
		log.WithField("protocol", protocol).Debug("handling connection")
		defer log.WithField("protocol", protocol).Debug("handling connection closed")

		return handler(c)
	})
}

func (t *TunnelV1) Listen(ctx context.Context, listener net.Listener) error {
	return t.srv.Serve(listener)
}

func (t *TunnelV1) Close() error {
	if err := t.router.Close(); err != nil {
		return err
	}

	return t.srv.Close()
}
