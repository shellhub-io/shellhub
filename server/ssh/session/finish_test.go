package session

import (
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	servicemocks "github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer/dialertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newConnectedSession(t *testing.T, tunnel dialer.TunnelDialer) (*Session, *servicemocks.MockService) {
	t.Helper()

	Configure(Config{ConnectTimeout: 0}) //nolint:exhaustruct

	service := servicemocks.NewMockService(t)
	service.On("DeactivateSession", mock.Anything, models.UID("test-uid")).Return(nil).Maybe()

	sess := newTestSession(service, tunnel)
	sess.Events = NewEvents(sess.UID, service)
	sess.registered = true

	require.NoError(t, sess.connect(newStubContext(), noAuth))

	return sess, service
}

// TestFinishLeavesTheSSHTransportIntact is what the seam was missing. Finalisation used to
// write a raw HTTP request onto the connection the SSH client owns, which the agent cannot
// account for as an SSH packet and which breaks the transport out from under any seat still
// on it. The close request belongs on its own connection, the way the close handler sends it.
func TestFinishLeavesTheSSHTransportIntact(t *testing.T) {
	agent := dialertest.NewAgent(t)
	sess, _ := newConnectedSession(t, agent)

	require.True(t, agent.Serving(100*time.Millisecond), "the agent must be reachable before finalisation")
	require.NoError(t, sess.Finish())

	assert.True(t, agent.Serving(2*time.Second),
		"finalisation must not write onto the SSH transport the client owns")
}

// TestFinishAsksTheTunnelToCloseTheSession pins where the close request goes instead: a
// connection of its own, dialled for closing this session.
func TestFinishAsksTheTunnelToCloseTheSession(t *testing.T) {
	agent := dialertest.NewAgent(t)
	sess, _ := newConnectedSession(t, agent)

	require.NoError(t, sess.Finish())

	assert.Eventually(t, func() bool {
		return len(agent.Dials()) == 2
	}, 5*time.Second, 10*time.Millisecond, "the close is dialled off the teardown path")

	assert.Equal(t, []dialertest.Dial{
		{Tenant: "tenant-id", UID: "device-uid", Target: dialer.SSHOpenTarget{SessionID: "test-uid"}},
		{Tenant: "tenant-id", UID: "device-uid", Target: dialer.SSHCloseTarget{SessionID: "test-uid"}},
	}, agent.Dials())
}

// TestFinishDeactivatesTheSessionWhenTheCloseCannotBeDelivered covers the device that dropped
// its tunnel before the session ended: the close request has nowhere to go, and the API must
// still be told the session is over.
func TestFinishDeactivatesTheSessionWhenTheCloseCannotBeDelivered(t *testing.T) {
	sess, service := newConnectedSession(t, dialertest.NewAgent(t))
	sess.dialer = &dialertest.Stub{Err: dialer.ErrNoConnection} //nolint:exhaustruct

	require.NoError(t, sess.Finish())

	service.AssertNumberOfCalls(t, "DeactivateSession", 1)
	service.AssertCalled(t, "DeactivateSession", mock.Anything, models.UID("test-uid"))
}

// TestFinishSkipsTheCloseWhenTheDeviceWasNeverReached keeps a failed dial from producing a
// second one: there is no session on the device to close.
func TestFinishSkipsTheCloseWhenTheDeviceWasNeverReached(t *testing.T) {
	service := servicemocks.NewMockService(t)
	service.On("DeactivateSession", mock.Anything, models.UID("test-uid")).Return(nil).Once()

	stub := &dialertest.Stub{Err: dialer.ErrNoConnection} //nolint:exhaustruct

	sess := newTestSession(service, stub)
	sess.Events = NewEvents(sess.UID, service)
	sess.registered = true

	require.NoError(t, sess.Finish())

	assert.Empty(t, stub.Dials(), "a session that never reached the device has nothing to close")
}

// TestFinishSkipsTheDeactivationWhenTheLoginNeverRegistered covers the login abandoned at the
// approval prompt. It holds no session on the API, so asking for one to be deactivated reports
// a missing session as a failure on a path that did nothing wrong.
func TestFinishSkipsTheDeactivationWhenTheLoginNeverRegistered(t *testing.T) {
	service := servicemocks.NewMockService(t)

	sess := newTestSession(service, dialertest.NewAgent(t))
	sess.Events = NewEvents(sess.UID, service)

	require.NoError(t, sess.Finish())

	service.AssertNotCalled(t, "DeactivateSession", mock.Anything, mock.Anything)
}

// TestFinishRunsOnce keeps finalisation idempotent: the close request is sent, and the session
// deactivated, exactly once however many times the teardown paths call it.
func TestFinishRunsOnce(t *testing.T) {
	agent := dialertest.NewAgent(t)
	sess, service := newConnectedSession(t, agent)

	require.NoError(t, sess.Finish())
	require.NoError(t, sess.Finish())

	assert.Eventually(t, func() bool {
		return len(agent.Dials()) == 2
	}, 5*time.Second, 10*time.Millisecond, "the open and the close, and no second close")

	assert.Never(t, func() bool {
		return len(agent.Dials()) > 2
	}, 200*time.Millisecond, 20*time.Millisecond, "the second Finish must not dial again")

	service.AssertNumberOfCalls(t, "DeactivateSession", 1)
}
