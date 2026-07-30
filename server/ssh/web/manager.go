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

// get gets the credentials if it time period have not ended.
func (m *manager) get(id string) (*Credentials, bool) {
	l, ok := m.credentials.Load(id)
	if !ok {
		return nil, false
	}

	v, ok := l.(*Credentials)

	return v, ok
}
