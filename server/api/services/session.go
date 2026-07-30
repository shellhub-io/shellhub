package services

import (
	"context"
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
	// Virtual bool-backed fields intercepted by ParseFilterProperty before any
	// SQL column binding — safe to accept bool-convertible values with eq/ne.
	// "active" is declared virtual here so that IsVirtualBoolField accurately
	// reflects the store-layer intercept, even though "active" currently only
	// allows "bool" (not eq/ne) and the distinction is not yet load-bearing.
	"active",
)

type SessionService interface {
	ListSessions(ctx context.Context, sc scope.Scope, req *requests.ListSessions) ([]models.Session, int, error)

	// GetSession fetches a session within the given namespace scope. The scope is an explicit
	// parameter rather than something recovered from the request context, so a caller cannot
	// receive a cross-namespace read by omission.
	GetSession(ctx context.Context, sc scope.Scope, uid models.UID) (*models.Session, error)
	CreateSession(ctx context.Context, session requests.SessionCreate) (*models.Session, error)
	DeactivateSession(ctx context.Context, uid models.UID) error
	KeepAliveSession(ctx context.Context, uid models.UID) error
	UpdateSession(ctx context.Context, uid models.UID, model models.SessionUpdate) error
	EventSession(ctx context.Context, events []models.SessionEvent) error
}

func (s *service) ListSessions(ctx context.Context, sc scope.Scope, req *requests.ListSessions) ([]models.Session, int, error) {
	opts := make([]store.QueryOption, 0)
	opts = append(opts, s.store.Options().Match(&req.Filters))
	opts = append(opts, s.store.Options().Sort(&query.Sorter{By: "started_at", Order: query.OrderDesc, Tiebreak: "id"}))
	opts = append(opts, s.store.Options().Paginate(&req.Paginator))

	return s.store.SessionList(ctx, sc, opts...)
}

func (s *service) GetSession(ctx context.Context, sc scope.Scope, uid models.UID) (*models.Session, error) {
	session, err := s.store.SessionResolve(ctx, sc, store.SessionUIDResolver, string(uid))
	if err != nil {
		return nil, NewErrSessionNotFound(uid, err)
	}

	return session, nil
}

func (s *service) CreateSession(ctx context.Context, session requests.SessionCreate) (*models.Session, error) {
	position, _ := s.locator.GetPosition(net.ParseIP(session.IPAddress))

	uid, err := s.store.SessionCreate(ctx, models.Session{
		UID:       session.UID,
		DeviceUID: models.UID(session.DeviceUID),
		Username:  session.Username,
		UserID:    session.UserID,
		IPAddress: session.IPAddress,
		Type:      session.Type,
		Term:      session.Term,
		Web:       session.Web,
		Position: models.SessionPosition{
			Longitude: position.Longitude,
			Latitude:  position.Latitude,
		},
	})
	if err != nil {
		return nil, err
	}

	// Reading back the row just written under a UID this call generated; there is no namespace to
	// bound by until the store has resolved it from the device.
	return s.store.SessionResolve(ctx, scope.NewUnbounded("reading back the session this call just created, by its generated UID"), store.SessionUIDResolver, uid)
}

func (s *service) DeactivateSession(ctx context.Context, uid models.UID) error {
	sess, err := s.store.SessionResolve(ctx, scope.NewUnbounded(reasonInternalSessionMutation), store.SessionUIDResolver, string(uid))
	if err != nil {
		return NewErrSessionNotFound(uid, err)
	}

	return s.store.ActiveSessionDelete(ctx, models.UID(sess.UID))
}

func (s *service) KeepAliveSession(ctx context.Context, uid models.UID) error {
	session, err := s.store.SessionResolve(ctx, scope.NewUnbounded(reasonInternalSessionMutation), store.SessionUIDResolver, string(uid))
	if err != nil {
		return NewErrSessionNotFound(uid, err)
	}

	session.LastSeen = clock.Now()

	return s.store.SessionUpdate(ctx, session)
}

func (s *service) UpdateSession(ctx context.Context, uid models.UID, model models.SessionUpdate) error {
	session, err := s.store.SessionResolve(ctx, scope.NewUnbounded(reasonInternalSessionMutation), store.SessionUIDResolver, string(uid))
	if err != nil {
		return NewErrSessionNotFound(uid, err)
	}

	if model.Authenticated != nil {
		session.Authenticated = *model.Authenticated
	}

	if model.Type != nil {
		session.Type = *model.Type
	}

	if model.Recorded != nil {
		session.Recorded = *model.Recorded
	}

	// We need to create an active session when authenticated to maintain compatibility with the old store implementation.
	// In the future, we may refactor the store to remove the active_session pattern.
	if session.Authenticated {
		if err := s.store.ActiveSessionCreate(ctx, session); err != nil {
			log.WithError(err).WithField("session_id", session.UID).Warn("failed to activate the session")
		}
	}

	return s.store.SessionUpdate(ctx, session)
}

// EventSession records session events.
//
// It does not check that the session exists first. session_events.session_id
// references sessions(id), so an event for a session that is gone is refused by the
// database anyway — and the check is not free: resolving a session aggregates the types
// and seats of every event it already has, which made recording an event cost more the
// longer the session ran.
func (s *service) EventSession(ctx context.Context, events []models.SessionEvent) error {
	return s.store.SessionEventsCreateMany(ctx, events)
}
