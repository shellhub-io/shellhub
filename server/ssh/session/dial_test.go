package session

import (
	"testing"

	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer/dialertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDialAsksTheTunnelToOpenTheSession pins what the session states at the call site: the
// device it wants, and that the connection is for opening this session and not for closing it.
func TestDialAsksTheTunnelToOpenTheSession(t *testing.T) {
	agent := dialertest.NewAgent(t)
	sess := newTestSession(nil, agent)

	require.NoError(t, sess.Dial(newStubContext()))

	assert.Equal(t, []dialertest.Dial{{
		Tenant: "tenant-id",
		UID:    "device-uid",
		Target: dialer.SSHOpenTarget{SessionID: "test-uid"},
	}}, agent.Dials())

	assert.NotNil(t, sess.agent.conn, "a successful dial must leave the session holding the connection")
}

// TestDialReportsAFailedTunnelAsErrDial keeps the banner's classification working: every dial
// failure reaches the caller under ErrDial, with the tunnel's own error still readable under it.
func TestDialReportsAFailedTunnelAsErrDial(t *testing.T) {
	stub := &dialertest.Stub{Err: dialer.ErrNoConnection} //nolint:exhaustruct
	sess := newTestSession(nil, stub)

	err := sess.Dial(newStubContext())

	require.ErrorIs(t, err, ErrDial)
	require.ErrorIs(t, err, dialer.ErrNoConnection)
	assert.Nil(t, sess.agent.conn, "a failed dial must not leave a connection behind")
}
