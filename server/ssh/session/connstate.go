package session

import (
	"net"

	gliderssh "github.com/gliderlabs/ssh"
	log "github.com/sirupsen/logrus"
)

type snapshot struct {
	session *Session
	state   State
}

// State type defines the current State of an associated session. It is used as "savepoints" for snapshots.
type State int

// The states a session passes through, in order. The zero value is not a state, so that an
// uninitialised session is distinguishable from one that has only just been created.
const (
	StateNil        State = iota + 1 // StateNil represents a non initialized session.
	StateCreated                     // StateCreated represents a session that has been created but not yet registered with the API.
	StateDialed                      // StateDialed represents a session that has been connected to a device.
	StateEvaluated                   // StateEvaluated represents a evaluated session.
	StateRegistered                  // StateRegistered represents a session that has been registered with the API but not yet connected to an agent.
	StateFinished                    // StateFinished represents a session that has been completed.
)

func (s State) String() string {
	switch s {
	case StateNil:
		return "nil"
	case StateCreated:
		return "created"
	case StateDialed:
		return "dialed"
	case StateEvaluated:
		return "evaluated"
	case StateRegistered:
		return "registered"
	case StateFinished:
		return "finished"
	default:
		return "unknown"
	}
}

// Evaluated reports whether the session has passed evaluation and may now
// authenticate.
func (s State) Evaluated() bool {
	return s >= StateEvaluated
}

// Established reports whether authentication has completed, so the session owns
// a connection to the agent and can carry channels.
func (s State) Established() bool {
	return s >= StateFinished
}

// snapshotCtxKey and connCtxKey address the per-connection values the SSH path
// hangs off [gliderssh.Context]. They are types rather than strings so an
// unrelated SetValue cannot be read back as connection state, and so a typo is
// a compile error instead of a nil at runtime.
type (
	snapshotCtxKey struct{}
	connCtxKey     struct{}
)

func getSnapshot(ctx gliderssh.Context) *snapshot {
	if snap, ok := ctx.Value(snapshotCtxKey{}).(*snapshot); ok && snap != nil {
		return snap
	}

	snap := &snapshot{session: nil, state: StateNil}
	ctx.SetValue(snapshotCtxKey{}, snap)

	return snap
}

func (s *snapshot) save(session *Session, state State) {
	s.session = session
	s.state = state
}

func (s *snapshot) retrieve() (*Session, State) {
	return s.session, s.state
}

// ObtainSession obtains a session and its state from the provided context. If there's
// no session associated, it creates a new one with state [StateNil].
func ObtainSession(ctx gliderssh.Context) (*Session, State) {
	return getSnapshot(ctx).retrieve()
}

func advance(ctx gliderssh.Context, session *Session, state State) {
	getSnapshot(ctx).save(session, state)
}

// StoreConn records the connection underlying this SSH context, so a handler
// that must refuse a login can drop it. The SSH server owns the connection; the
// auth handlers only ever reach it through [AuthenticableSessionOrDrop].
func StoreConn(ctx gliderssh.Context, conn net.Conn) {
	ctx.SetValue(connCtxKey{}, conn)
}

// AuthenticableSessionOrDrop returns the session for a connection that has reached
// evaluation and may now authenticate.
//
// When the connection has not got that far there is nothing to authenticate
// against, so it drops the connection and reports false: the caller must refuse
// the attempt without reporting anything further to the client.
func AuthenticableSessionOrDrop(ctx gliderssh.Context) (*Session, bool) {
	sess, state := ObtainSession(ctx)
	if sess == nil || !state.Evaluated() {
		log.WithFields(log.Fields{"uid": ctx.SessionID(), "sshid": ctx.User(), "state": state}).
			Trace("dropping a connection that tried to authenticate before evaluation")

		if conn, ok := ctx.Value(connCtxKey{}).(net.Conn); ok {
			conn.Close() //nolint:errcheck
		}

		return nil, false
	}

	return sess, true
}
