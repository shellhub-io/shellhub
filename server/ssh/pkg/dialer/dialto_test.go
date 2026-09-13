package dialer

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newDialerWith(t *testing.T, key string, tunnel any) *Dialer {
	t.Helper()

	d := &Dialer{Manager: NewManager()}

	if tunnel != nil {
		d.Manager.Connections.Store(key, tunnel)
	}

	return d
}

// TestDialToRejectsAnUnnameableDevice covers the mode that is a programming error rather than
// a device state: without both halves there is no tunnel key to look up.
func TestDialToRejectsAnUnnameableDevice(t *testing.T) {
	cases := []struct {
		description string
		tenant      string
		uid         string
	}{
		{description: "no tenant", tenant: "", uid: "device"},
		{description: "no device", tenant: "tenant", uid: ""},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			d := newDialerWith(t, "", nil)

			conn, err := d.DialTo(context.Background(), tc.tenant, tc.uid, SSHOpenTarget{SessionID: "session"})

			require.ErrorIs(t, err, ErrInvalidArgument)
			assert.Nil(t, conn)
		})
	}
}

// TestDialToReportsADeviceHoldingNoTunnel pins how an offline device is observed: not as an
// unreachable one, so that a caller can tell a device that will come back from one to look at.
func TestDialToReportsADeviceHoldingNoTunnel(t *testing.T) {
	d := newDialerWith(t, "", nil)

	conn, err := d.DialTo(context.Background(), "tenant", "device", SSHOpenTarget{SessionID: "session"})

	require.ErrorIs(t, err, ErrNoConnection)
	require.NotErrorIs(t, err, ErrUnreachable, "a device holding no tunnel is offline, not unreachable")
	assert.Nil(t, conn)
}

// TestDialToReportsATunnelThatCannotCarryAStream covers the device that is present but whose
// tunnel is gone underneath it.
func TestDialToReportsATunnelThatCannotCarryAStream(t *testing.T) {
	server, agent := net.Pipe()

	t.Cleanup(func() {
		_ = server.Close()
		_ = agent.Close()
	})

	session, err := yamux.Client(server, nil)
	require.NoError(t, err)
	require.NoError(t, session.Close())

	d := newDialerWith(t, NewKey("tenant", "device"), session)

	conn, err := d.DialTo(context.Background(), "tenant", "device", SSHOpenTarget{SessionID: "session"})

	require.ErrorIs(t, err, ErrUnreachable)
	require.NotErrorIs(t, err, ErrNoConnection, "the device is present; only the exchange failed")
	assert.Nil(t, conn)
}

// TestDialToReportsAnAgentThatDoesNotAnswer covers the other half of unreachable: the stream
// opens and the agent never completes the bootstrap on it.
func TestDialToReportsAnAgentThatDoesNotAnswer(t *testing.T) {
	server, agent := net.Pipe()

	t.Cleanup(func() {
		_ = server.Close()
		_ = agent.Close()
	})

	session, err := yamux.Client(server, nil)
	require.NoError(t, err)

	t.Cleanup(func() { _ = session.Close() })

	peer, err := yamux.Server(agent, nil)
	require.NoError(t, err)

	t.Cleanup(func() { _ = peer.Close() })

	go func() {
		if stream, err := peer.Accept(); err == nil {
			<-t.Context().Done()

			_ = stream.Close()
		}
	}()

	d := newDialerWith(t, NewKey("tenant", "device"), session)

	ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
	defer cancel()

	conn, err := d.DialTo(ctx, "tenant", "device", SSHOpenTarget{SessionID: "session"})

	require.ErrorIs(t, err, ErrUnreachable)
	assert.Nil(t, conn)
}

// TestDescribeNamesEachFailure keeps the three modes distinguishable to an operator reading the
// log, which is the only place the difference surfaces today.
func TestDescribeNamesEachFailure(t *testing.T) {
	cases := []struct {
		description string
		err         error
		expected    string
	}{
		{
			description: "a device holding no tunnel is offline",
			err:         ErrNoConnection,
			expected:    "the device holds no tunnel",
		},
		{
			description: "a device holding a wedged tunnel is present",
			err:         errors.Join(ErrUnreachable, io.EOF),
			expected:    "the device holds a tunnel but did not answer",
		},
		{
			description: "an empty tenant or device is a programming error",
			err:         ErrInvalidArgument,
			expected:    "the device cannot be named",
		},
		{
			description: "anything else is not classified",
			err:         io.EOF,
			expected:    "the tunnel could not be dialled",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expected, Describe(tc.err))
		})
	}
}
