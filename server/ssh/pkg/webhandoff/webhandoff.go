// Package webhandoff carries the context the web terminal bridge cannot express
// over SSH.
//
// The bridge is an SSH client living inside the SSH server: it dials the local
// listener so a browser session goes through the very same handshake, access
// policies and recording as a native client. The SSH protocol has nowhere to
// put the browser's address or the logged-in account, so the bridge parks them
// here under the synthetic username it dials with, and the session picks them
// up when it recognises a loopback peer.
//
// This used to travel through Redis. Both sides are the same process, so the
// round trip bought nothing: a full write, read and delete against the cache to
// move sixty bytes between two goroutines, plus a one minute expiry that made a
// slow handshake fail outright.
package webhandoff

import (
	"sync"
	"time"
)

const entryTTL = 5 * time.Minute

// Data is what the bridge knows and the SSH handshake cannot carry.
type Data struct {
	// Device is the UID the browser asked for. The synthetic username the bridge
	// dials with cannot hold it, because the SSHID grammar is reserved for
	// namespace.device.
	Device string
	// IP is the browser's address, not the loopback address of the dial. It ends
	// up on the session record and feeds the firewall and geoip decisions.
	IP string
	// UserID is the logged-in account, set only in identity mode. Legacy web
	// sessions authenticate with a device credential and leave it empty.
	UserID string
}

// Store holds the pending handoffs. It is safe for concurrent use.
type Store struct {
	entries *sync.Map
}

// NewStore returns an empty [Store].
func NewStore() *Store {
	return &Store{entries: new(sync.Map)}
}

// Put parks the data under id, which is the username the bridge is about to
// dial with.
func (s *Store) Put(id string, data Data) {
	s.entries.Store(id, data)

	time.AfterFunc(entryTTL, func() {
		s.entries.Delete(id)
	})
}

// Take returns the data for id and removes it, so a handoff is single use: a
// second connection claiming the same username finds nothing.
func (s *Store) Take(id string) (Data, bool) {
	value, ok := s.entries.LoadAndDelete(id)
	if !ok {
		return Data{}, false
	}

	data, ok := value.(Data)

	return data, ok
}
