// Package dialertest provides substitutes for the reverse tunnel a device holds open to the
// server, so that the SSH path can be driven from a unit test without a live agent.
//
// [Agent] stands in for a device that answers, [Stub] for a dial that fails. Both satisfy
// [dialer.TunnelDialer], which is the whole of what the SSH path asks of a tunnel. Neither
// speaks the V1 or V2 bootstrap: that belongs behind the seam and keeps its own tests there.
package dialertest

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

// Dial is one call the code under test made on a tunnel.
type Dial struct {
	Tenant string
	UID    string
	Target dialer.Target
}

// Agent is a device that holds a tunnel open. Every dial is recorded so a caller can assert
// what the tunnel was asked for, and a dial that opens a session lands on an in-process SSH
// server reached over an in-memory pipe.
//
// The zero value is not usable; build one with [NewAgent] or [NewSilentAgent].
type Agent struct {
	t *testing.T

	signer gossh.Signer
	speaks bool

	broken chan struct{}
	once   sync.Once

	mu     sync.Mutex
	dials  []Dial
	served bool
}

// NewAgent returns an [Agent] that completes the SSH handshake and then accepts session
// channels, which is what a device does once the server reaches it.
func NewAgent(t *testing.T) *Agent {
	t.Helper()

	return newAgent(t, true)
}

// NewSilentAgent returns an [Agent] that accepts the dial and then never speaks, which is how
// a device that holds a tunnel it can no longer serve presents itself. A handshake against it
// only ends on a deadline.
func NewSilentAgent(t *testing.T) *Agent {
	t.Helper()

	return newAgent(t, false)
}

func newAgent(t *testing.T, speaks bool) *Agent {
	t.Helper()

	_, key, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	signer, err := gossh.NewSignerFromKey(key)
	require.NoError(t, err)

	return &Agent{ //nolint:exhaustruct // the recording fields start empty and are appended to under the mutex
		t:      t,
		signer: signer,
		speaks: speaks,
		broken: make(chan struct{}),
	}
}

// DialTo records the dial and returns the server's end of a pipe whose far end is the agent.
// The target is recorded rather than prepared: the bootstrap it names lives behind the seam.
func (a *Agent) DialTo(ctx context.Context, tenant, uid string, target dialer.Target) (net.Conn, error) { //nolint:ireturn // net.Conn is the interface TunnelDialer returns
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	a.mu.Lock()
	a.dials = append(a.dials, Dial{Tenant: tenant, UID: uid, Target: target})
	a.mu.Unlock()

	server, device := memPipe(a.t)

	switch {
	case !a.speaks:
	case isMessage(target):
		go io.Copy(io.Discard, device) //nolint:errcheck // the agent reads the message and the stream ends; there is no caller to report to
	default:
		go a.serve(device)
	}

	return server, nil
}

func isMessage(target dialer.Target) bool {
	_, closing := target.(dialer.SSHCloseTarget)

	return closing
}

func (a *Agent) serve(device net.Conn) {
	config := &gossh.ServerConfig{NoClientAuth: true} //nolint:exhaustruct // a fake agent authenticates nobody; the host key is added below
	config.AddHostKey(a.signer)

	conn, chans, reqs, err := gossh.NewServerConn(device, config)
	if err != nil {
		a.breakDown()

		return
	}

	a.mu.Lock()
	a.served = true
	a.mu.Unlock()

	go gossh.DiscardRequests(reqs)
	go func() {
		_ = conn.Wait()

		a.breakDown()
	}()

	for newChannel := range chans {
		channel, requests, err := newChannel.Accept()
		if err != nil {
			return
		}

		go gossh.DiscardRequests(requests)
		go io.Copy(io.Discard, channel) //nolint:errcheck // the fake agent consumes whatever a seat writes and answers nothing
	}
}

func (a *Agent) breakDown() {
	a.once.Do(func() { close(a.broken) })
}

// Dials returns the dials the agent was asked for, in order.
func (a *Agent) Dials() []Dial {
	a.mu.Lock()
	defer a.mu.Unlock()

	return append([]Dial(nil), a.dials...)
}

// Serving reports whether the agent has a session transport it can still read from, which
// stops being true the moment anything writes bytes the SSH transport cannot account for.
//
// It waits up to grace for the transport to fail, so that a caller asserting the transport
// survived does not race the write it is asserting about.
func (a *Agent) Serving(grace time.Duration) bool {
	select {
	case <-a.broken:
		return false
	case <-time.After(grace):
		a.mu.Lock()
		defer a.mu.Unlock()

		return a.served
	}
}

// Stub is a tunnel that answers every dial with Err and no connection, for driving a caller's
// failure paths. It records what it was asked for, so a caller that must not dial at all is
// assertable too.
type Stub struct {
	Err error

	mu    sync.Mutex
	dials []Dial
}

// DialTo records the dial and returns the stub's Err.
func (s *Stub) DialTo(_ context.Context, tenant, uid string, target dialer.Target) (net.Conn, error) { //nolint:ireturn // net.Conn is the interface TunnelDialer returns
	s.mu.Lock()
	s.dials = append(s.dials, Dial{Tenant: tenant, UID: uid, Target: target})
	s.mu.Unlock()

	return nil, s.Err
}

// Dials returns the dials the stub was asked for, in order.
func (s *Stub) Dials() []Dial {
	s.mu.Lock()
	defer s.mu.Unlock()

	return append([]Dial(nil), s.dials...)
}
