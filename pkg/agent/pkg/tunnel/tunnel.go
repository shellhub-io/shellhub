package tunnel

import (
	"context"
	"io"
	"net"

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

		return handler(NewContext(context.TODO(), rwc), rwc)
	})
}

func (t *Tunnel) Listen(conn net.Conn) error {
	// TODO: configure the mux server.
	session, err := yamux.Server(conn, nil)
	if err != nil {
		log.WithError(err).Error("failed to create muxed session")

		return err
	}

	for {
		stream, err := session.Accept()
		if err != nil {
			log.WithError(err).Error("failed to accept stream")

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
