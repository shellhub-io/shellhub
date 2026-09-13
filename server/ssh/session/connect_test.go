package session

import (
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer/dialertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

func noAuth(*Session, *gossh.ClientConfig) error { return nil }

// TestConnectBoundsASilentAgent is the reason ConnectTimeout exists.
//
// An agent that accepts the stream and then says nothing leaves the SSH
// handshake with nothing to wait on. The library offers no deadline of its own
// here — ClientConfig.Timeout only bounds the TCP dial in gossh.Dial, which is
// not the path this takes — so without the deadline the goroutine parks for the
// life of the process.
//
// Note this covers the handshake specifically. A device that has gone away
// entirely fails earlier, in the dialer, on its own deadlines.
func TestConnectBoundsASilentAgent(t *testing.T) {
	Configure(Config{ //nolint:exhaustruct
		ConnectTimeout: 2 * time.Second,
	})

	defer Configure(Config{ConnectTimeout: 0}) //nolint:exhaustruct

	sess := newTestSession(nil, dialertest.NewSilentAgent(t))

	done := make(chan error, 1)

	start := time.Now() //nolint:forbidigo // a deadline, an elapsed-time measurement, or the clock mock itself

	go func() { done <- sess.connect(newStubContext(), noAuth) }()

	select {
	case err := <-done:
		require.Error(t, err, "a silent agent must not authenticate")
		assert.WithinDuration(t, start.Add(2*time.Second), clock.Now(), 3*time.Second,
			"the handshake should give up at the configured timeout")
	case <-time.After(30 * time.Second):
		t.Fatal("the handshake against a silent agent never gave up")
	}
}

// TestConnectWithoutTimeoutNeverGivesUp pins why the deadline is the mechanism:
// with it unset the same silent agent holds the handshake open indefinitely,
// which is what the code did before ConnectTimeout was wired up.
func TestConnectWithoutTimeoutNeverGivesUp(t *testing.T) {
	Configure(Config{ConnectTimeout: 0}) //nolint:exhaustruct

	sess := newTestSession(nil, dialertest.NewSilentAgent(t))

	done := make(chan error, 1)
	go func() { done <- sess.connect(newStubContext(), noAuth) }()

	select {
	case <-done:
		t.Fatal("it returned without a deadline; the timeout is not what bounds this")
	case <-time.After(3 * time.Second):
	}
}

// TestConnectReachesTheAgentThroughTheTunnel covers the path the two tests above only fail
// on: a device that answers leaves the session holding an SSH client it can open channels on.
func TestConnectReachesTheAgentThroughTheTunnel(t *testing.T) {
	Configure(Config{ConnectTimeout: 0}) //nolint:exhaustruct

	agent := dialertest.NewAgent(t)
	sess := newTestSession(nil, agent)

	require.NoError(t, sess.connect(newStubContext(), noAuth))

	channel, _, err := sess.agent.client.OpenChannel("session", nil)
	require.NoError(t, err, "the session must hold a usable SSH client after connect")

	assert.NoError(t, channel.Close())
}
