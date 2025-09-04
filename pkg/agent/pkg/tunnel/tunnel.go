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
	// YamuxAcceptBacklog is used to limit how many streams may be
	// waiting an accept.
	YamuxAcceptBacklog int `json:"yamux_accept_backlog"`

	// EnableKeepalive is used to do a period keep alive
	// messages using a ping.
	YamuxEnableKeepAlive bool `json:"yamux_enable_keep_alive"`

	// YamuxKeepAliveInterval is how often to perform the keep alive
	YamuxKeepAliveInterval time.Duration `json:"yamux_keep_alive_interval"`

	// YamuxConnectionWriteTimeout is meant to be a "safety valve" timeout after
	// we which will suspect a problem with the underlying connection and
	// close it. This is only applied to writes, where's there's generally
	// an expectation that things will move along quickly.
	YamuxConnectionWriteTimeout time.Duration `json:"yamux_connection_write_timeout"`

	// YamuxMaxStreamWindowSize is used to control the maximum
	// window size that we allow for a stream.
	YamuxMaxStreamWindowSize uint32 `json:"yamux_max_stream_window_size"`

	// YamuxStreamOpenTimeout is the maximum amount of time that a stream will
	// be allowed to remain in pending state while waiting for an ack from the peer.
	// Once the timeout is reached the session will be gracefully closed.
	// A zero value disables the YamuxStreamOpenTimeout allowing unbounded
	// blocking on OpenStream calls.
	YamuxStreamOpenTimeout time.Duration `json:"yamux_stream_open_timeout"`

	// YamuxStreamCloseTimeout is the maximum time that a stream will allowed to
	// be in a half-closed state when `Close` is called before forcibly
	// closing the connection. Forcibly closed connections will empty the
	// receive buffer, drop any future packets received for that stream,
	// and send a RST to the remote side.
	YamuxStreamCloseTimeout time.Duration `json:"yamux_stream_close_timeout"`
}

// NewConfigFromMap creates a new Config from a map[string]any received from auth data from the server
// or returns the default config if the map is nil. If a key is missing, the default value is used.
func NewConfigFromMap(m map[string]any) *Config {
	cfg := DefaultConfig

	if v, ok := m["yamux_accept_backlog"].(int); ok {
		cfg.YamuxAcceptBacklog = v
	}

	if v, ok := m["yamux_enable_keep_alive"].(bool); ok {
		cfg.YamuxEnableKeepAlive = v
	}

	if v, ok := m["yamux_keep_alive_interval"].(time.Duration); ok {
		cfg.YamuxKeepAliveInterval = v
	}

	if v, ok := m["yamux_connection_write_timeout"].(time.Duration); ok {
		cfg.YamuxConnectionWriteTimeout = v
	}

	if v, ok := m["yamux_max_stream_window_size"].(uint32); ok {
		cfg.YamuxMaxStreamWindowSize = v
	}

	if v, ok := m["yamux_stream_open_timeout"].(time.Duration); ok {
		cfg.YamuxStreamOpenTimeout = v
	}

	if v, ok := m["yamux_stream_close_timeout"].(time.Duration); ok {
		cfg.YamuxStreamCloseTimeout = v
	}

	return &cfg
}

func YamuxConfigFromConfig(cfg *Config) *yamux.Config {
	if cfg == nil {
		cfg = &DefaultConfig
	}

	return &yamux.Config{
		AcceptBacklog:          cfg.YamuxAcceptBacklog,
		EnableKeepAlive:        cfg.YamuxEnableKeepAlive,
		KeepAliveInterval:      cfg.YamuxKeepAliveInterval,
		ConnectionWriteTimeout: cfg.YamuxConnectionWriteTimeout,
		MaxStreamWindowSize:    cfg.YamuxMaxStreamWindowSize,
		StreamCloseTimeout:     cfg.YamuxStreamCloseTimeout,
		StreamOpenTimeout:      cfg.YamuxStreamOpenTimeout,
		LogOutput:              os.Stderr,
	}
}

var DefaultConfig = Config{
	YamuxAcceptBacklog:          256,
	YamuxEnableKeepAlive:        true,
	YamuxKeepAliveInterval:      35 * time.Second,
	YamuxConnectionWriteTimeout: 15 * time.Second,
	YamuxMaxStreamWindowSize:    256 * 1024,
	YamuxStreamCloseTimeout:     5 * time.Minute,
	YamuxStreamOpenTimeout:      75 * time.Second,
}

func (t *Tunnel) Listen(conn net.Conn, cfg *Config) error {
	if cfg == nil {
		cfg = &DefaultConfig
	}

	var session *yamux.Session
	var err error

	session, err = yamux.Server(conn, YamuxConfigFromConfig(cfg))
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"yamux_accept_backlog":           cfg.YamuxAcceptBacklog,
			"yamux_enable_keep_alive":        cfg.YamuxEnableKeepAlive,
			"yamux_keep_alive_interval":      cfg.YamuxKeepAliveInterval,
			"yamux_connection_write_timeout": cfg.YamuxConnectionWriteTimeout,
			"yamux_max_stream_window_size":   cfg.YamuxMaxStreamWindowSize,
			"yamux_stream_close_timeout":     cfg.YamuxStreamCloseTimeout,
			"yamux_stream_open_timeout":      cfg.YamuxStreamOpenTimeout,
		}).Error("failed to create muxed session")

		// NOTE: If we fail to create the session, we should try again with the [DefaultConfig] as the client
		// could be using different settings.
		log.WithError(err).Warning("trying to create muxed session with default config")
		session, err = yamux.Server(conn, YamuxConfigFromConfig(&DefaultConfig))
		if err != nil {
			log.WithError(err).Error("failed to create muxed session with default config")

			return err
		}

		log.WithError(err).Warning("muxed session created with default config due to error with custom config")
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
