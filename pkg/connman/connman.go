package connman

import (
	"context"
	"errors"
	"net"

	"github.com/shellhub-io/shellhub/pkg/revdial"
	"github.com/sirupsen/logrus"
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

func (m *ConnectionManager) Set(key string, conn net.Conn) {
	dialer := revdial.NewDialer(conn, "/ssh/revdial")

	m.dialers.Store(key, dialer)

	if size := m.dialers.Size(key); size > 1 {
		logrus.WithFields(logrus.Fields{
			"key":  key,
			"size": size,
		}).Warning("Multiple connections stored for the same identifier.")
	}

	go func() {
		for {
			select {
			case <-dialer.KeepAlives():
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

func (m *ConnectionManager) Dial(ctx context.Context, key string) (net.Conn, error) {
	dialer, ok := m.dialers.Load(key)
	if !ok {
		return nil, ErrNoConnection
	}

	if size := m.dialers.Size(key); size > 1 {
		logrus.WithFields(logrus.Fields{
			"key":  key,
			"size": size,
		}).Warning("Multiple connections found for the same identifier during reverse tunnel dialing.")
	}

	return dialer.(*revdial.Dialer).Dial(ctx)
}
