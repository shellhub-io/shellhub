package store

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type SessionResolver uint

const (
	SessionUIDResolver SessionResolver = iota + 1
)

// ExpiredSession is a session that has outlived the retention window, reduced to what deleting
// it needs to know. Recorded travels with the UID because it decides whether the session owns a
// recording at all: without it, an instance that records nothing still pays a storage lookup per
// session it deletes.
type ExpiredSession struct {
	UID      string
	Recorded bool
}

type SessionStore interface {
	// SessionList retrieves a list of sessions based on the provided filters and pagination settings.
	// It returns the list of sessions, the total count of matching documents, and an error if any.
	SessionList(ctx context.Context, sc scope.Scope, opts ...QueryOption) ([]models.Session, int, error)
	// SessionResolve fetches a session using a specific resolver within the given namespace scope.
	// It returns the resolved session if found and an error, if any.
	SessionResolve(ctx context.Context, sc scope.Scope, resolver SessionResolver, value string, opts ...QueryOption) (*models.Session, error)
	// SessionCreate creates a new session. It returns the inserted UID and an error if any.
	SessionCreate(ctx context.Context, session models.Session) (string, error)
	// SessionUpdate updates a session. It returns an error if any.
	SessionUpdate(ctx context.Context, session *models.Session) error

	// ActiveSessionCreate creates an active session entry. It returns an error if any.
	ActiveSessionCreate(ctx context.Context, session *models.Session) error
	// ActiveSessionResolve fetches an active session using a specific resolver. It returns the active session if found and an error, if any.
	ActiveSessionResolve(ctx context.Context, resolver SessionResolver, value string) (*models.ActiveSession, error)
	// ActiveSessionUpdate updates an active session. It returns an error if any.
	ActiveSessionUpdate(ctx context.Context, activeSession *models.ActiveSession) error

	// ActiveSessionDelete removes active session entries. It returns an error if any.
	ActiveSessionDelete(ctx context.Context, uid models.UID) error

	// SessionEventsCreate creates a session event. It returns an error if any.
	SessionEventsCreate(ctx context.Context, event *models.SessionEvent) error
	// SessionEventsCreateMany creates session events in a single statement. Events
	// carry their own timestamp, so a batch does not reorder anything a reader
	// sees. It returns an error if any; an empty slice is a no-op.
	SessionEventsCreateMany(ctx context.Context, events []models.SessionEvent) error
	// SessionEventsList retrieves session events based on filters. It returns the list of events, total count, and an error if any.
	SessionEventsList(ctx context.Context, uid models.UID, seat int, event models.SessionEventType, opts ...QueryOption) ([]models.SessionEvent, int, error)
	// SessionEventsDelete removes session events based on filters. It returns an error if any.
	SessionEventsDelete(ctx context.Context, uid models.UID, seat int, event models.SessionEventType) error

	// SessionUpdateDeviceUID updates device UID references across sessions. It returns an error if any.
	SessionUpdateDeviceUID(ctx context.Context, oldUID models.UID, newUID models.UID) error

	// SessionListExpired returns up to limit sessions started before the given time, oldest
	// first. A session that is still active is never returned, however old it is, and a limit
	// that is not positive returns nothing.
	//
	// Listing is separate from deleting so a caller can act on what a session owns outside the
	// database while the row that names it still exists.
	SessionListExpired(ctx context.Context, before time.Time, limit int) ([]ExpiredSession, error)

	// SessionDeleteMany deletes the given sessions, cascading into their events. It returns the
	// number deleted, which may be lower than the number asked for if a session went away in
	// between. An empty slice is a no-op.
	SessionDeleteMany(ctx context.Context, uids []string) (int64, error)
}
