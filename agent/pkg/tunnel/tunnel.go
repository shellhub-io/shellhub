package tunnel

import (
	"context"
	"errors"
	"io"
	"net"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/yamux"
	"github.com/multiformats/go-multistream"
	log "github.com/sirupsen/logrus"
)

type Tunnel struct {
	mux *multistream.MultistreamMuxer[string]
}

func NewTunnel() *Tunnel {
	return &Tunnel{
		mux: multistream.NewMultistreamMuxer[string](),
	}
}

func (t *Tunnel) Handle(protocol string, handler Handler) {
	t.mux.AddHandler(protocol, func(protocol string, rwc io.ReadWriteCloser) error {
		log.WithField("protocol", protocol).Debug("handling connection")
		defer log.WithField("protocol", protocol).Debug("handling connection closed")

		// TODO: Should we receive a context from outside?
		return handler(NewContext(context.TODO(), rwc), rwc)
	})
}

// ErrTunnelDisconnect is returned when the tunnel connection is closed.
var ErrTunnelDisconnect = errors.New("tunnel disconnected")

func (t *Tunnel) Listen(conn net.Conn) error {
	session, err := yamux.Server(conn, &yamux.Config{
		AcceptBacklog:          256,
		EnableKeepAlive:        true,
		KeepAliveInterval:      35 * time.Second,
		ConnectionWriteTimeout: 15 * time.Second,
		MaxStreamWindowSize:    256 * 1024,
		StreamCloseTimeout:     5 * time.Minute,
		StreamOpenTimeout:      75 * time.Second,
		LogOutput:              os.Stderr,
	})
	if err != nil {
		log.WithError(err).Error("failed to create muxed session")

		return err
	}

	for {
		stream, err := session.Accept()
		if err != nil {
			defer session.Close()

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
