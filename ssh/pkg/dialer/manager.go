package dialer

import (
	"context"
	"errors"
	"net"
	"os"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/shellhub-io/shellhub/pkg/revdial"
	"github.com/shellhub-io/shellhub/pkg/wsconnadapter"
	log "github.com/sirupsen/logrus"
)

var ErrNoConnection = errors.New("no connection")

type Manager struct {
	Connections             *SyncSliceMap
	DialerDoneCallback      func(string)
	DialerKeepAliveCallback func(string)
}

func NewManager() *Manager {
	return &Manager{
		Connections:             &SyncSliceMap{},
		DialerDoneCallback:      func(string) {},
		DialerKeepAliveCallback: func(string) {},
	}
}

func (m *Manager) Set(key string, conn *wsconnadapter.Adapter, connPath string) {
	dialer := revdial.NewDialer(conn.Logger, conn, connPath)

	m.Connections.Store(key, dialer)

	if size := m.Connections.Size(key); size > 1 {
		log.WithFields(log.Fields{
			"key":  key,
			"size": size,
		}).Warning("Multiple connections stored for the same identifier.")
	}

	m.DialerKeepAliveCallback(key)

	// Start the ping loop and get the channel for pong responses
	pong := conn.Ping()

	go func() {
		for {
			select {
			case <-pong:
				m.DialerKeepAliveCallback(key)

				continue
			case <-dialer.Done():
				m.Connections.Delete(key, dialer)
				m.DialerDoneCallback(key)

				return
			}
		}
	}()
}

// BindPingInterval is the interval between pings sent to the yamux session
// to keep it alive. It should be less than the NAT timeout to avoid
// disconnections.
// It should be the same value as used by the revdial.Dialer ping interval.
const BindPingInterval = 35 * time.Second

// Bind binds a WebSocket connection to a yamux session and stores it in the connection manager.
// All new agents should use this handler to register their reverse connection.
func (m *Manager) Bind(tenant string, uid string, conn *wsconnadapter.Adapter) error {
	key := NewKey(tenant, uid)

	session, err := yamux.Client(conn, &yamux.Config{
		AcceptBacklog: 256,
		// NOTE: As we need to keep the registered connection alive, we use our own ping/pong mechanism.
		EnableKeepAlive: false,
		// NOTE: Although we disable the built-in keepalive, we still need to set the interval to a non-zero value to
		// avoid yamux error when verifying the configuration. We've created a Pull Request to improve this behavior.
		// TODO: Remove this workaround when yamux supports disabling keepalive completely.
		KeepAliveInterval:      BindPingInterval,
		ConnectionWriteTimeout: 15 * time.Second,
		MaxStreamWindowSize:    256 * 1024,
		StreamCloseTimeout:     5 * time.Minute,
		StreamOpenTimeout:      75 * time.Second,
		LogOutput:              os.Stderr,
	})
	if err != nil {
		log.WithError(err).Error("failed to create yamux client session")

		return err
	}

	m.Connections.Store(key, session)

	if size := m.Connections.Size(key); size > 1 {
		log.WithFields(log.Fields{
			"key":  key,
			"size": size,
		}).Warning("Multiple connections stored for the same identifier.")
	}

	m.DialerKeepAliveCallback(key)

	go func() {
		for {
			select {
			// NOTE: Ping is also important to keep the underlying WebSocket connection alive and avoid NAT timeouts.
			case <-time.After(BindPingInterval):
				if _, err := session.Ping(); err != nil {
					log.WithFields(log.Fields{
						"key": key,
					}).WithError(err).Error("failed to ping yamux session")

					m.Connections.Delete(key, session)
					m.DialerDoneCallback(key)

					return
				}

				m.DialerKeepAliveCallback(key)

				continue
			case <-session.CloseChan():
				m.Connections.Delete(key, session)
				m.DialerDoneCallback(key)

				return
			}
		}
	}()

	return nil
}

// ConnectionVersion protocol version identifiers used when dialing a device.
type ConnectionVersion byte

const (
	// ConnectionVersionUnknown is used when the transport version could not be determined.
	ConnectionVersionUnknown ConnectionVersion = 0
	// ConnectionVersion1 is the legacy transport using revdial over HTTP.
	ConnectionVersion1 ConnectionVersion = 1
	// ConnectionVersion2 is the current transport using yamux multiplexing.
	ConnectionVersion2 ConnectionVersion = 2
)

// Dial tries to find a connection by its key and dials it.
//
// It returns the connection, its version ([ConnectionVersion1] or [ConnectionVersion2]) and an error,
func (m *Manager) Dial(ctx context.Context, key string) (net.Conn, ConnectionVersion, error) {
	loaded, ok := m.Connections.Load(key)
	if !ok {
		return nil, ConnectionVersionUnknown, ErrNoConnection
	}

	if size := m.Connections.Size(key); size > 1 {
		log.WithFields(log.Fields{
			"key":  key,
			"size": size,
		}).Warning("Multiple connections found for the same identifier during reverse tunnel dialing.")
	}

	if dialer, ok := loaded.(*revdial.Dialer); ok {
		log.WithFields(log.Fields{
			"key":     key,
			"version": "v1",
		}).Debug("using v1 dialer for reverse tunnel dialing")

		conn, err := dialer.Dial(ctx)
		if err != nil {
			log.WithFields(log.Fields{
				"key":     key,
				"version": "v1",
			}).WithError(err).Error("failed to dial reverse connection")

			return nil, ConnectionVersionUnknown, err
		}

		return conn, ConnectionVersion1, nil
	}

	if session, ok := loaded.(*yamux.Session); ok {
		log.WithFields(log.Fields{
			"key":     key,
			"version": "v2",
		}).Debug("using v2 connection for reverse tunnel dialing")

		conn, err := session.Open()
		if err != nil {
			log.WithFields(log.Fields{
				"key":     key,
				"version": "v2",
			}).WithError(err).Error("failed to open yamux stream for reverse connection")
		}

		return conn, ConnectionVersion2, nil
	}

	return nil, ConnectionVersionUnknown, ErrNoConnection
}
