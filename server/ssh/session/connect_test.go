package session

import (
	"net"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

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
	listener, err := new(net.ListenConfig).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)

	defer listener.Close() //nolint:errcheck

	accepted := make(chan net.Conn, 1)

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}

		accepted <- conn
	}()

	peer, err := new(net.Dialer).DialContext(t.Context(), "tcp", listener.Addr().String())
	require.NoError(t, err)

	defer peer.Close() //nolint:errcheck

	Configure(Config{ //nolint:exhaustruct
		ConnectTimeout: 2 * time.Second,
	})

	defer Configure(Config{ConnectTimeout: 0}) //nolint:exhaustruct

	sess := newTestSession(nil)
	sess.agent.conn = peer

	noAuth := func(*Session, *gossh.ClientConfig) error { return nil }

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

	if conn := <-accepted; conn != nil {
		_ = conn.Close()
	}
}

// TestConnectWithoutTimeoutNeverGivesUp pins why the deadline is the mechanism:
// with it unset the same silent agent holds the handshake open indefinitely,
// which is what the code did before ConnectTimeout was wired up.
func TestConnectWithoutTimeoutNeverGivesUp(t *testing.T) {
	listener, err := new(net.ListenConfig).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)

	defer listener.Close() //nolint:errcheck

	go func() {
		if conn, err := listener.Accept(); err == nil {
			defer conn.Close() //nolint:errcheck

			time.Sleep(10 * time.Second)
		}
	}()

	peer, err := new(net.Dialer).DialContext(t.Context(), "tcp", listener.Addr().String())
	require.NoError(t, err)

	defer peer.Close() //nolint:errcheck

	Configure(Config{ConnectTimeout: 0}) //nolint:exhaustruct

	sess := newTestSession(nil)
	sess.agent.conn = peer

	noAuth := func(*Session, *gossh.ClientConfig) error { return nil }

	done := make(chan error, 1)
	go func() { done <- sess.connect(newStubContext(), noAuth) }()

	select {
	case <-done:
		t.Fatal("it returned without a deadline; the timeout is not what bounds this")
	case <-time.After(3 * time.Second):
	}
}
