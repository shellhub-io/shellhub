package connman

import (
	"context"
	"errors"
	"net"

	"github.com/hashicorp/yamux"
	"github.com/shellhub-io/shellhub/pkg/revdial"
	"github.com/shellhub-io/shellhub/pkg/wsconnadapter"
	log "github.com/sirupsen/logrus"
)

var ErrNoConnection = errors.New("no connection")

type ConnectionManager struct {
	dialers                 *SyncSliceMap
	DialerDoneCallback      func(string, *revdial.Dialer)
	DialerKeepAliveCallback func(string, *revdial.Dialer)
}

func New() *ConnectionManager {
	return &ConnectionManager{
		dialers: &SyncSliceMap{},
		DialerDoneCallback: func(string, *revdial.Dialer) {
		},
	}
}

func (m *ConnectionManager) Set(key string, conn *wsconnadapter.Adapter, connPath string) {
	dialer := revdial.NewDialer(conn.Logger, conn, connPath)

	m.dialers.Store(key, dialer)

	if size := m.dialers.Size(key); size > 1 {
		log.WithFields(log.Fields{
			"key":  key,
			"size": size,
		}).Warning("Multiple connections stored for the same identifier.")
	}

	m.DialerKeepAliveCallback(key, dialer)

	// Start the ping loop and get the channel for pong responses
	pong := conn.Ping()

	go func() {
		for {
			select {
			case <-pong:
				m.DialerKeepAliveCallback(key, dialer)

				continue
			case <-dialer.Done():
				m.dialers.Delete(key, dialer)
				m.DialerDoneCallback(key, dialer)

				return
			}
		}
	}()
}

// Bind binds a WebSocket connection to a yamux session and stores it in the connection manager.
func (m *ConnectionManager) Bind(key string, conn *wsconnadapter.Adapter) {
	// TODO: configure yamux client session.
	session, err := yamux.Client(conn, nil)
	if err != nil {
		log.WithError(err).Error("failed to create yamux client session")

		return
	}

	m.dialers.Store(key, session)

	if size := m.dialers.Size(key); size > 1 {
		log.WithFields(log.Fields{
			"key":  key,
			"size": size,
		}).Warning("Multiple connections stored for the same identifier.")
	}
}

const (
	ConnectionUnknown byte = 0
	ConnectionV1      byte = 1
	ConnectionV2      byte = 2
)

// Dial tries to find a connection by its key and dials it.
//
// It returns the connection, its version (v1 or v2) and an error,
func (m *ConnectionManager) Dial(ctx context.Context, key string) (net.Conn, byte, error) {
	loaded, ok := m.dialers.Load(key)
	if !ok {
		return nil, ConnectionUnknown, ErrNoConnection
	}

	if size := m.dialers.Size(key); size > 1 {
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

			return nil, ConnectionUnknown, err
		}

		return conn, ConnectionV1, nil
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

		return conn, ConnectionV2, nil
	}

	return nil, ConnectionUnknown, nil
}
