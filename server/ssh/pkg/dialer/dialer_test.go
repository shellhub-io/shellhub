package dialer

import (
	"context"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingConn struct {
	net.Conn

	mu        sync.Mutex
	deadlines []time.Time
	closed    bool
}

func (c *recordingConn) SetDeadline(t time.Time) error {
	c.mu.Lock()
	c.deadlines = append(c.deadlines, t)
	c.mu.Unlock()

	return c.Conn.SetDeadline(t)
}

func (c *recordingConn) Close() error {
	c.mu.Lock()
	c.closed = true
	c.mu.Unlock()

	return c.Conn.Close()
}

func (c *recordingConn) isClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.closed
}

func (c *recordingConn) deadlineAt(i int) (time.Time, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if i < 0 {
		i += len(c.deadlines)
	}

	if i < 0 || i >= len(c.deadlines) {
		return time.Time{}, false
	}

	return c.deadlines[i], true
}

func unresponsivePeer(t *testing.T) *recordingConn {
	t.Helper()

	server, agent := net.Pipe()

	go io.Copy(io.Discard, agent) //nolint:errcheck

	t.Cleanup(func() {
		_ = agent.Close()
	})

	return &recordingConn{Conn: server}
}

func TestHandshakeFailsAndClosesTheStream(t *testing.T) {
	cases := []struct {
		description string
		target      Target
		version     TransportVersion
		ctx         func(t *testing.T) context.Context
	}{
		{
			description: "http proxy v2 negotiation the agent never answers",
			target:      HTTPProxyTarget{RequestID: "request", Host: "localhost", Port: 8080},
			version:     TransportVersion2,
			ctx: func(t *testing.T) context.Context {
				t.Helper()

				ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
				t.Cleanup(cancel)

				return ctx
			},
		},
		{
			description: "ssh open v2 protocol selection the agent never answers",
			target:      SSHOpenTarget{SessionID: "session"},
			version:     TransportVersion2,
			ctx: func(t *testing.T) context.Context {
				t.Helper()

				ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
				t.Cleanup(cancel)

				return ctx
			},
		},
		{
			description: "caller gives up while the handshake is in flight",
			target:      HTTPProxyTarget{RequestID: "request", Host: "localhost", Port: 8080},
			version:     TransportVersion2,
			ctx: func(t *testing.T) context.Context {
				t.Helper()

				ctx, cancel := context.WithCancel(context.Background())
				time.AfterFunc(100*time.Millisecond, cancel)
				t.Cleanup(cancel)

				return ctx
			},
		},
		{
			description: "unsupported transport version",
			target:      SSHOpenTarget{SessionID: "session"},
			version:     TransportVersionUnknown,
			ctx: func(_ *testing.T) context.Context {
				return context.Background()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			conn := unresponsivePeer(t)

			done := make(chan error, 1)
			go func() {
				_, err := handshake(tc.ctx(t), conn, tc.version, tc.target)
				done <- err
			}()

			select {
			case err := <-done:
				require.Error(t, err)
			case <-time.After(5 * time.Second):
				t.Fatal("handshake blocked instead of failing")
			}

			assert.True(t, conn.isClosed(), "a failed handshake must not leak the stream")
		})
	}
}

func TestHandshakeReportsCancellationAsSuch(t *testing.T) {
	cases := []struct {
		description string
		version     TransportVersion
	}{
		{
			description: "cancelled while the handshake is blocked",
			version:     TransportVersion2,
		},
		{
			description: "cancelled as the handshake completes",
			version:     TransportVersion1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			conn := unresponsivePeer(t)

			ctx, cancel := context.WithCancel(context.Background())
			if tc.version == TransportVersion2 {
				time.AfterFunc(100*time.Millisecond, cancel)
			} else {
				cancel()
			}

			defer cancel()

			_, err := handshake(ctx, conn, tc.version, SSHOpenTarget{SessionID: "session"})
			assert.ErrorIs(t, err, context.Canceled, "the caller must be able to tell a hang-up from a broken agent")
		})
	}
}

func TestHandshakeClearsDeadlineOnSuccess(t *testing.T) {
	conn := unresponsivePeer(t)

	prepared, err := handshake(context.Background(), conn, TransportVersion1, SSHOpenTarget{SessionID: "session"})
	require.NoError(t, err)
	assert.Equal(t, net.Conn(conn), prepared)

	deadline, ok := conn.deadlineAt(-1)
	require.True(t, ok, "the handshake must bound itself with a deadline")
	assert.True(t, deadline.IsZero(), "the deadline must be cleared before streaming")
	assert.False(t, conn.isClosed())
}

func TestHandshakeHonorsTheEarlierDeadline(t *testing.T) {
	conn := unresponsivePeer(t)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	callerDeadline, _ := ctx.Deadline()

	_, err := handshake(ctx, conn, TransportVersion2, SSHOpenTarget{SessionID: "session"})
	require.Error(t, err)

	first, ok := conn.deadlineAt(0)
	require.True(t, ok)

	assert.Equal(t, callerDeadline, first, "the caller's deadline is earlier than HandshakeTimeout and must win")
}
