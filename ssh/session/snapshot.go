package session

import gliderssh "github.com/gliderlabs/ssh"

type snapshot struct {
	session *Session
	state   State
}

// State type defines the current State of an associated session. It is used as "savepoints" for snapshots.
type State int

const (
	StateNil        = iota // stateNil represents an empty session.
	StateCreated           // stateCreated represents a session that has been created but not yet registered with the API.
	StateRegistered        // stateRegistered represents a session that has been registered with the API but not yet connected to an agent.
	StateFinished          // stateFinished represents a session that has been completed.
)

// getSnapshot is responsible for managing the state of a session associated with
// the provided context. It creates a new getSnapshot if one is not already associated.
// A getSnapshot can be used to retrieve and modify the current state of a session, enabling
// composition between steps and avoiding redundant operations.
//
// Utilize [save] to store the state of a session and [retrieve] to retrieve it.
func getSnapshot(ctx gliderssh.Context) *snapshot {
	if snap, ok := ctx.Value("snap").(*snapshot); ok && snap != nil {
		return snap
	}

	snap := &snapshot{session: nil, state: StateNil}
	ctx.SetValue("snap", snap)

	return snap
}

// save stores the provided session and it's state.
func (s *snapshot) save(session *Session, state State) {
	s.session = session
	s.state = state
}

// retrieve retrieves the current state and the associated session.
func (s *snapshot) retrieve() (*Session, State) {
	return s.session, s.state
}

// ObtainSession obtains a session and its state from the provided context. If there's
// no session associated, it creates a new one with state [StateNil].
func ObtainSession(ctx gliderssh.Context) (*Session, State) {
	return getSnapshot(ctx).retrieve()
}
