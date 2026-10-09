package services

import (
	"context"
	"errors"
	"net"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	log "github.com/sirupsen/logrus"
)

// SessionFilterFields maps each filter field the session list endpoint accepts
// to the set of operators valid for it.
//
// "closed" and "active" are boolean-typed; only the "bool" operator is
// permitted. Allowing "eq" on a boolean column lets a string value
// (e.g. "true") bypass validation but fail at the Postgres level with
// "operator does not exist: boolean = text", producing a 500 instead of 400.
var SessionFilterFields = query.NewFieldConstraints(map[string][]string{
	"device_uid": {"eq", "ne"},
	"closed":     {"bool"},
	"active":     {"bool"},
},
	"active",
)

// SessionService owns SSH session records and their recordings — the history of who reached
// which device, and what was seen on screen.
type SessionService interface {
	ListSessions(ctx context.Context, sc scope.Scope, req *requests.ListSessions) ([]models.Session, int64, error)

	// GetSession fetches a session within the given namespace scope. The scope is an explicit
	// parameter rather than something recovered from the request context, so a caller cannot
	// receive a cross-namespace read by omission.
	GetSession(ctx context.Context, sc scope.Scope, uid models.UID) (*models.Session, error)
	CreateSession(ctx context.Context, session requests.SessionCreate) (*models.Session, error)
	// DeactivateSession retires the session within the namespace sc is bounded to. It returns
	// ErrSessionNotFound when no session in that namespace has the UID, store.ErrInvalidScope when
	// sc is not bounded, and the store's error when retiring the session fails.
	DeactivateSession(ctx context.Context, sc scope.Scope, uid models.UID) error
	// KeepAliveSession stamps the session within the namespace sc is bounded to as still live. It
	// returns ErrSessionNotFound when no session in that namespace has the UID, store.ErrInvalidScope
	// when sc is not bounded, and the store's error when the write fails.
	KeepAliveSession(ctx context.Context, sc scope.Scope, uid models.UID) error
	// UpdateSession applies model to the session within the namespace sc is bounded to. It returns
	// ErrSessionNotFound when no session in that namespace has the UID, store.ErrInvalidScope when sc
	// is not bounded, and the store's error when the write fails. A failure to put an authenticated
	// session in the active set is logged, not returned.
	UpdateSession(ctx context.Context, sc scope.Scope, uid models.UID, model models.SessionUpdate) error
	// EventSession records events against sessions in the namespace sc is bounded to, all of
	// them or none. It returns store.ErrNoDocuments when any event's session is gone or outside
	// that namespace, store.ErrInvalidScope when sc is not bounded, and the store's error when the
	// write fails.
	EventSession(ctx context.Context, sc scope.Scope, events []models.SessionEvent) error
}

func (s *service) ListSessions(ctx context.Context, sc scope.Scope, req *requests.ListSessions) ([]models.Session, int64, error) {
	opts := make([]store.QueryOption, 0)
	opts = append(opts, s.store.Options().Match(&req.Filters))
	opts = append(opts, s.store.Options().Sort(&query.Sorter{By: "started_at", Order: query.OrderDesc, Tiebreak: "id"}))
	opts = append(opts, s.store.Options().Paginate(&req.Paginator))

	return s.store.SessionList(ctx, sc, opts...)
}

const sessionTimelineMaxEvents = 1000

func (s *service) GetSession(ctx context.Context, sc scope.Scope, uid models.UID) (*models.Session, error) {
	session, err := s.store.SessionResolve(ctx, sc, store.SessionUIDResolver, string(uid))
	if err != nil {
		return nil, NewErrSessionNotFound(uid, err)
	}

	events, err := s.store.SessionEventsTimeline(ctx, uid, sessionTimelineMaxEvents)
	if err != nil {
		log.WithError(err).
			WithField("session_uid", session.UID).
			Warn("failed to read the session timeline; returning the session without it")

		return session, nil
	}

	session.Events.Items = events

	return session, nil
}

func (s *service) CreateSession(ctx context.Context, session requests.SessionCreate) (*models.Session, error) {
	position, _ := s.locator.GetPosition(net.ParseIP(session.IPAddress))

	uid, err := s.store.SessionCreate(ctx, models.Session{
		UID:       session.UID,
		DeviceUID: models.UID(session.DeviceUID),
		Username:  session.Username,
		UserID:    session.UserID,
		APIKeyID:  session.APIKeyID,
		IPAddress: session.IPAddress,
		Web:       session.Web,
		Position: models.SessionPosition{
			Longitude: position.Longitude,
			Latitude:  position.Latitude,
		},
	})
	if err != nil {
		return nil, err
	}

	return s.store.SessionResolve(ctx, scope.NewUnbounded("reading back the session this call just created, by its generated UID"), store.SessionUIDResolver, uid)
}

func (s *service) DeactivateSession(ctx context.Context, sc scope.Scope, uid models.UID) error {
	if !sc.IsBounded() {
		return store.ErrInvalidScope
	}

	sess, err := s.store.SessionResolve(ctx, sc, store.SessionUIDResolver, string(uid))
	if err != nil {
		return NewErrSessionNotFound(uid, err)
	}

	return s.store.ActiveSessionDelete(ctx, sc, models.UID(sess.UID))
}

func (s *service) KeepAliveSession(ctx context.Context, sc scope.Scope, uid models.UID) error {
	err := s.store.SessionKeepAlive(ctx, sc, uid, clock.Now())
	switch {
	case err == nil:
		return nil
	case errors.Is(err, store.ErrNoDocuments):
		return NewErrSessionNotFound(uid, err)
	default:
		return err
	}
}

func (s *service) UpdateSession(ctx context.Context, sc scope.Scope, uid models.UID, model models.SessionUpdate) error {
	if !sc.IsBounded() {
		return store.ErrInvalidScope
	}

	session, err := s.store.SessionResolve(ctx, sc, store.SessionUIDResolver, string(uid))
	if err != nil {
		return NewErrSessionNotFound(uid, err)
	}

	if model.Authenticated != nil {
		session.Authenticated = *model.Authenticated
	}

	if model.Recorded != nil {
		session.Recorded = *model.Recorded
	}

	if session.Authenticated {
		if err := s.store.ActiveSessionCreate(ctx, session); err != nil {
			log.WithError(err).WithField("session_id", session.UID).Warn("failed to activate the session")
		}
	}

	return s.store.SessionUpdate(ctx, sc, session)
}

func (s *service) EventSession(ctx context.Context, sc scope.Scope, events []models.SessionEvent) error {
	return s.store.SessionEventsCreateMany(ctx, sc, events)
}
