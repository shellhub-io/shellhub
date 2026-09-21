package session

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	servicemocks "github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestKeepAliveLoopMarksTheSessionAlive(t *testing.T) {
	service := servicemocks.NewMockService(t)

	var calls atomic.Int32
	service.On("KeepAliveSession", mock.Anything, models.UID("test-uid")).
		Run(func(mock.Arguments) { calls.Add(1) }).
		Return(nil)

	sess := newTestSession(service, nil)
	ctx := newStubContext()

	sess.startKeepAlive(ctx, time.Millisecond)
	t.Cleanup(sess.stopKeepAlive)

	assert.Eventually(t, func() bool { return calls.Load() >= 3 }, time.Second, 5*time.Millisecond,
		"the ticker must keep marking the session alive while it lives")
}

func TestKeepAliveLoopStopsBeforeTheSessionIsDeactivated(t *testing.T) {
	service := servicemocks.NewMockService(t)

	var (
		seq        atomic.Int64
		deactivate atomic.Int64
		lastAlive  atomic.Int64
		calls      atomic.Int32
	)

	service.On("KeepAliveSession", mock.Anything, models.UID("test-uid")).
		Run(func(mock.Arguments) {
			lastAlive.Store(seq.Add(1))
			calls.Add(1)
		}).
		Return(nil)

	service.On("DeactivateSession", mock.Anything, models.UID("test-uid")).
		Run(func(mock.Arguments) {
			deactivate.Store(seq.Add(1))
		}).
		Return(nil).Once()

	sess := newTestSession(service, nil)
	sess.registered = true
	sess.Events = NewEvents("test-uid", service)

	ctx := newStubContext()
	sess.startKeepAlive(ctx, time.Millisecond)

	require.Eventually(t, func() bool { return calls.Load() >= 2 }, time.Second, 5*time.Millisecond)

	require.NoError(t, sess.Finish())

	assert.NotZero(t, deactivate.Load(), "the session must have been deactivated")
	assert.Less(t, lastAlive.Load(), deactivate.Load(),
		"the last keep-alive must precede deactivation, or it puts the ended session back in the active set")
}

func TestKeepAliveStartsOnlyOnce(t *testing.T) {
	sess := newTestSession(servicemocks.NewMockService(t), nil)
	ctx := newStubContext()

	sess.startKeepAlive(ctx, time.Hour)
	t.Cleanup(sess.stopKeepAlive)

	running := sess.keepaliveDone

	sess.startKeepAlive(ctx, time.Millisecond)

	assert.Equal(t, running, sess.keepaliveDone, "a second start must leave the running loop in place")
}

func TestKeepAliveNeverStartsOnceStopped(t *testing.T) {
	sess := newTestSession(servicemocks.NewMockService(t), nil)

	sess.stopKeepAlive()
	sess.startKeepAlive(newStubContext(), time.Millisecond)

	assert.Nil(t, sess.keepaliveDone, "a keep-alive started after the session finished would outlive its deactivation")
}
