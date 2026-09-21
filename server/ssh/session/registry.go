package session

import "sync"

// Registry is the set of sessions this gateway process owns, meaning the ones whose teardown it
// will run. A process starts with an empty one, so after a restart every session it held reads as
// unowned. It is per process: a horizontally scaled gateway needs a shared one, or a close arriving
// at the wrong replica reads a live session as unowned.
type Registry struct {
	mu  sync.Mutex
	ids map[string]struct{}
}

// NewRegistry returns an empty Registry, ready to use.
func NewRegistry() *Registry {
	return &Registry{ids: make(map[string]struct{})} //nolint:exhaustruct // the zero mutex is the usable one
}

// Add records that this process owns the session.
func (r *Registry) Add(uid string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.ids[uid] = struct{}{}
}

// Remove drops the session from the set. It is safe to call for one that was never added, and
// twice for the same one.
func (r *Registry) Remove(uid string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.ids, uid)
}

// Has reports whether this process owns the session.
func (r *Registry) Has(uid string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.ids[uid]

	return ok
}
