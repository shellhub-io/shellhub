package share

import (
	"sync"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
)

// entry holds a live share together with the device that owns it and its expiry.
type entry struct {
	hub       *Hub
	deviceUID string
	tenantID  string
	name      string
	command   string
	writable  bool
	createdAt time.Time
	expiresAt time.Time
}

// registry tracks live shares in memory. Each share's hub is in-memory by nature (it holds live
// websocket subscribers), so the registry itself is the source of truth and an expiry timer
// cleans up abandoned shares — mirroring the TTL approach used by the web-terminal manager.
type Registry struct {
	mu      sync.RWMutex
	entries map[string]*entry
	ttl     time.Duration
}

func NewRegistry(ttl time.Duration) *Registry {
	return &Registry{
		entries: make(map[string]*entry),
		ttl:     ttl,
	}
}

// create allocates a new share for the given device and returns its token, hub and expiry.
func (r *Registry) create(deviceUID, tenantID string, req models.ShareCreateRequest) (string, *entry) {
	token := uuid.Generate()

	now := clock.Now()

	// Resolve the requested lifetime: a negative TTL means never expire (the share only ends when
	// the producer disconnects); zero falls back to the server default; positive is a custom span.
	ttl := r.ttl
	noExpiry := req.TTLSeconds < 0
	if req.TTLSeconds > 0 {
		ttl = time.Duration(req.TTLSeconds) * time.Second
	}

	var expiresAt time.Time
	if !noExpiry {
		expiresAt = now.Add(ttl)
	}

	e := &entry{
		hub:       newHub(),
		deviceUID: deviceUID,
		tenantID:  tenantID,
		name:      req.Name,
		command:   req.Command,
		writable:  req.Writable,
		createdAt: now,
		expiresAt: expiresAt,
	}

	r.mu.Lock()
	r.entries[token] = e
	r.mu.Unlock()

	if !noExpiry {
		go time.AfterFunc(ttl, func() {
			r.remove(token)
		})
	}

	return token, e
}

// CreateLocal registers a share whose producer is an in-process terminal session (e.g. the web
// console) instead of an external agent stream. It returns the token, the hub to feed output into
// and drain guest input from, and a close function to tear the share down when the session ends.
func (r *Registry) CreateLocal(deviceUID, tenantID string, req models.ShareCreateRequest) (string, *Hub, func()) {
	token, e := r.create(deviceUID, tenantID, req)

	return token, e.hub, func() { r.remove(token) }
}

// get returns the share entry for a token if it exists and has not expired.
func (r *Registry) get(token string) (*entry, bool) {
	r.mu.RLock()
	e, ok := r.entries[token]
	r.mu.RUnlock()

	if !ok {
		return nil, false
	}

	if !e.expiresAt.IsZero() && clock.Now().After(e.expiresAt) {
		r.remove(token)

		return nil, false
	}

	return e, true
}

// list returns all live (non-expired) shares belonging to the given tenant.
func (r *Registry) list(tenantID string) map[string]*entry {
	now := clock.Now()

	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make(map[string]*entry)
	for token, e := range r.entries {
		if e.tenantID == tenantID && (e.expiresAt.IsZero() || !now.After(e.expiresAt)) {
			out[token] = e
		}
	}

	return out
}

// remove deletes a share and tears down its hub.
func (r *Registry) remove(token string) {
	r.mu.Lock()
	e, ok := r.entries[token]
	if ok {
		delete(r.entries, token)
	}
	r.mu.Unlock()

	if ok {
		e.hub.Close()
	}
}
