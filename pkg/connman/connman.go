package connman

import (
	"context"
	"errors"
	"net"
	"sync"

	"github.com/shellhub-io/shellhub/pkg/revdial"
)

var ErrNoConnection = errors.New("no connection")

type ConnectionManager struct {
	dialers                 sync.Map
	DialerDoneCallback      func(string, *revdial.Dialer)
	DialerKeepAliveCallback func(string, *revdial.Dialer)
}

func New() *ConnectionManager {
	return &ConnectionManager{
		DialerDoneCallback: func(string, *revdial.Dialer) {
		},
	}
}

func (m *ConnectionManager) Set(key string, conn net.Conn) {
	dialer := revdial.NewDialer(conn, "/ssh/revdial")

	m.dialers.Store(key, dialer)

	go func() {
		for {
			select {
			case <-dialer.KeepAlives():
				m.DialerKeepAliveCallback(key, dialer)

				continue
			case <-dialer.Done():
				m.dialers.Delete(key)
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

	return dialer.(*revdial.Dialer).Dial(ctx)
}
