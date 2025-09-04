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

type Config struct {
	// AcceptBacklog is used to limit how many streams may be
	// waiting an accept.
	AcceptBacklog int `json:"accept_backlog"`

	// EnableKeepalive is used to do a period keep alive
	// messages using a ping.
	EnableKeepAlive bool `json:"enable_keep_alive"`

	// KeepAliveInterval is how often to perform the keep alive
	KeepAliveInterval time.Duration `json:"keep_alive_interval"`

	// ConnectionWriteTimeout is meant to be a "safety valve" timeout after
	// we which will suspect a problem with the underlying connection and
	// close it. This is only applied to writes, where's there's generally
	// an expectation that things will move along quickly.
	ConnectionWriteTimeout time.Duration `json:"connection_write_timeout"`

	// MaxStreamWindowSize is used to control the maximum
	// window size that we allow for a stream.
	MaxStreamWindowSize uint32 `json:"max_stream_window_size"`

	// StreamOpenTimeout is the maximum amount of time that a stream will
	// be allowed to remain in pending state while waiting for an ack from the peer.
	// Once the timeout is reached the session will be gracefully closed.
	// A zero value disables the StreamOpenTimeout allowing unbounded
	// blocking on OpenStream calls.
	StreamOpenTimeout time.Duration `json:"stream_open_timeout"`

	// StreamCloseTimeout is the maximum time that a stream will allowed to
	// be in a half-closed state when `Close` is called before forcibly
	// closing the connection. Forcibly closed connections will empty the
	// receive buffer, drop any future packets received for that stream,
	// and send a RST to the remote side.
	StreamCloseTimeout time.Duration `json:"stream_close_timeout"`
}

var DefaultConfig = Config{
	AcceptBacklog:          256,
	EnableKeepAlive:        true,
	KeepAliveInterval:      35 * time.Second,
	ConnectionWriteTimeout: 15 * time.Second,
	MaxStreamWindowSize:    256 * 1024,
	StreamCloseTimeout:     5 * time.Minute,
	StreamOpenTimeout:      75 * time.Second,
}

func (t *Tunnel) Listen(conn net.Conn, cfg *Config) error {
	if cfg == nil {
		cfg = &DefaultConfig
	}

	session, err := yamux.Server(conn, &yamux.Config{
		AcceptBacklog:          cfg.AcceptBacklog,
		EnableKeepAlive:        cfg.EnableKeepAlive,
		KeepAliveInterval:      cfg.KeepAliveInterval,
		ConnectionWriteTimeout: cfg.ConnectionWriteTimeout,
		MaxStreamWindowSize:    cfg.MaxStreamWindowSize,
		StreamCloseTimeout:     cfg.StreamCloseTimeout,
		StreamOpenTimeout:      cfg.StreamOpenTimeout,
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
