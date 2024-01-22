package web

import (
	"sync"
	"time"
)

// Manager is used to store credentials for a time period.
type Manager struct {
	// ttl is the time that each credial live in the map.
	ttl         time.Duration
	credentials *sync.Map
}

// NewManager creates a new [Manager] to store the credentials for a time period.
func NewManager(ttl time.Duration) *Manager {
	return &Manager{
		ttl:         ttl,
		credentials: new(sync.Map),
	}
}

// Save credentials for a time period. After this, the credentials are deleted.
func (m *Manager) Save(id string, data *Credentials) {
	m.credentials.Store(id, data)

	go time.AfterFunc(m.ttl, func() {
		m.credentials.Delete(id)
	})
}

// Get gets the credentials if it time period have not ended.
func (m *Manager) Get(id string) (*Credentials, bool) {
	l, ok := m.credentials.Load(id)
	if !ok {
		return nil, false
	}

	v, ok := l.(*Credentials)

	return v, ok
}
