package session

import (
	"net"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recordingConn reports whether the connection-state module closed it.
type recordingConn struct {
	net.Conn
	closed atomic.Bool
}

func (c *recordingConn) Close() error {
	c.closed.Store(true)

	return nil
}

func TestStatePredicates(t *testing.T) {
	cases := []struct {
		state       State
		evaluated   bool
		established bool
	}{
		{StateNil, false, false},
		{StateCreated, false, false},
		{StateDialed, false, false},
		{StateEvaluated, true, false},
		{StateRegistered, true, false},
		{StateFinished, true, true},
	}

	for _, tt := range cases {
		assert.Equal(t, tt.evaluated, tt.state.Evaluated(), "Evaluated() for state %d", tt.state)
		assert.Equal(t, tt.established, tt.state.Established(), "Established() for state %d", tt.state)
	}
}

func TestObtainSessionOnAFreshContext(t *testing.T) {
	sess, state := ObtainSession(newStubContext())

	assert.Nil(t, sess)
	assert.Equal(t, StateNil, state)
}

// The snapshot must not be reachable through a string key: a typed key is what
// stops an unrelated ctx.SetValue from being read back as session state.
func TestObtainSessionIgnoresAStringKey(t *testing.T) {
	ctx := newStubContext()
	ctx.SetValue("snap", &snapshot{session: &Session{UID: "planted"}, state: StateFinished})

	sess, state := ObtainSession(ctx)

	assert.Nil(t, sess)
	assert.Equal(t, StateNil, state)
}

func TestAuthenticableSessionRefusesAndClosesBeforeEvaluation(t *testing.T) {
	for _, state := range []State{StateNil, StateCreated, StateDialed} {
		ctx := newStubContext()
		conn := new(recordingConn)
		StoreConn(ctx, conn)
		getSnapshot(ctx).save(&Session{UID: "uid"}, state)

		sess, ok := AuthenticableSessionOrDrop(ctx)

		assert.False(t, ok, "state %d must not authenticate", state)
		assert.Nil(t, sess)
		assert.True(t, conn.closed.Load(), "state %d must drop the connection", state)
	}
}

func TestAuthenticableSessionReturnsTheSessionOnceEvaluated(t *testing.T) {
	for _, state := range []State{StateEvaluated, StateRegistered, StateFinished} {
		ctx := newStubContext()
		conn := new(recordingConn)
		StoreConn(ctx, conn)
		want := &Session{UID: "uid"}
		getSnapshot(ctx).save(want, state)

		sess, ok := AuthenticableSessionOrDrop(ctx)

		require.True(t, ok, "state %d must authenticate", state)
		assert.Same(t, want, sess)
		assert.False(t, conn.closed.Load(), "state %d must keep the connection", state)
	}
}

// A connection that was never stored must not panic the refusal path.
func TestAuthenticableSessionWithoutAStoredConn(t *testing.T) {
	ctx := newStubContext()

	assert.NotPanics(t, func() {
		sess, ok := AuthenticableSessionOrDrop(ctx)

		assert.False(t, ok)
		assert.Nil(t, sess)
	})
}
