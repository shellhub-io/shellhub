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
	dialers            map[string]*revdial.Dialer
	lock               sync.RWMutex
	DialerDoneCallback func(string, *revdial.Dialer)
}

func New() *ConnectionManager {
	return &ConnectionManager{
		dialers: make(map[string]*revdial.Dialer),
		DialerDoneCallback: func(string, *revdial.Dialer) {
		},
	}
}

func (m *ConnectionManager) Set(key string, conn net.Conn) {
	m.lock.Lock()
	dialer := revdial.NewDialer(conn, "/ssh/revdial")
	m.dialers[key] = dialer
	m.lock.Unlock()

	go func() {
		<-dialer.Done()
		m.DialerDoneCallback(key, dialer)
	}()
}

func (m *ConnectionManager) Dial(ctx context.Context, key string) (net.Conn, error) {
	m.lock.RLock()
	dialer, ok := m.dialers[key]
	if !ok {
		m.lock.RUnlock()

		return nil, ErrNoConnection
	}
	m.lock.RUnlock()

	return dialer.Dial(ctx)
}
