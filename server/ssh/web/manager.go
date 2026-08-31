package web

import (
	"sync"
	"time"
)

type manager struct {
	ttl         time.Duration
	credentials *sync.Map
}

func newManager(ttl time.Duration) *manager {
	return &manager{
		ttl:         ttl,
		credentials: new(sync.Map),
	}
}

func (m *manager) save(id string, data *Credentials) {
	m.credentials.Store(id, data)

	go time.AfterFunc(m.ttl, func() {
		m.credentials.Delete(id)
	})
}

func (m *manager) get(id string) (*Credentials, bool) {
	l, ok := m.credentials.Load(id)
	if !ok {
		return nil, false
	}

	v, ok := l.(*Credentials)

	return v, ok
}
