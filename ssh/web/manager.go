package web

import (
	"sync"
	"time"
)

// manager is used to store credentials for a time period.
type manager struct {
	// ttl is the time that each credial live in the map.
	ttl         time.Duration
	credentials *sync.Map
}

// newManager creates a new [Manager] to store the credentials for a time period.
func newManager(ttl time.Duration) *manager {
	return &manager{
		ttl:         ttl,
		credentials: new(sync.Map),
	}
}

// save credentials for a time period. After this, the credentials are deleted.
func (m *manager) save(id string, data *Credentials) {
	m.credentials.Store(id, data)

	go time.AfterFunc(m.ttl, func() {
		m.credentials.Delete(id)
	})
}

// get consumes the credentials for id, if the TTL has not elapsed. The token is
// single-use: it is deleted on first read so a leaked token (it travels as a
// query param) can't be replayed within the TTL window.
func (m *manager) get(id string) (*Credentials, bool) {
	l, ok := m.credentials.LoadAndDelete(id)
	if !ok {
		return nil, false
	}

	v, ok := l.(*Credentials)

	return v, ok
}
