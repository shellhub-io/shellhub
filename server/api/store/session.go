package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type SessionResolver uint

const (
	SessionUIDResolver SessionResolver = iota + 1
)

type SessionStore interface {
	// SessionList retrieves a list of sessions based on the provided filters and pagination settings.
	// It returns the list of sessions, the total count of matching documents, and an error if any.
	SessionList(ctx context.Context, opts ...QueryOption) ([]models.Session, int, error)
	// SessionResolve fetches a session using a specific resolver.
	// It returns the resolved session if found and an error, if any.
	SessionResolve(ctx context.Context, resolver SessionResolver, value string, opts ...QueryOption) (*models.Session, error)
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
	// SessionEventsList retrieves session events based on filters. It returns the list of events, total count, and an error if any.
	SessionEventsList(ctx context.Context, uid models.UID, seat int, event models.SessionEventType, opts ...QueryOption) ([]models.SessionEvent, int, error)
	// SessionEventsDelete removes session events based on filters. It returns an error if any.
	SessionEventsDelete(ctx context.Context, uid models.UID, seat int, event models.SessionEventType) error

	// SessionUpdateDeviceUID updates device UID references across sessions. It returns an error if any.
	SessionUpdateDeviceUID(ctx context.Context, oldUID models.UID, newUID models.UID) error
}
