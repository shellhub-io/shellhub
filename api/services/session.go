package services

import (
	"context"
	"net"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
)

type SessionService interface {
	ListSessions(ctx context.Context, req *requests.ListSessions) ([]models.Session, int, error)
	GetSession(ctx context.Context, uid models.UID) (*models.Session, error)
	CreateSession(ctx context.Context, session requests.SessionCreate) (*models.Session, error)
	DeactivateSession(ctx context.Context, uid models.UID) error
	KeepAliveSession(ctx context.Context, uid models.UID) error
	UpdateSession(ctx context.Context, uid models.UID, model models.SessionUpdate) error
	EventSession(ctx context.Context, uid models.UID, event *models.SessionEvent) error
}

func (s *service) ListSessions(ctx context.Context, req *requests.ListSessions) ([]models.Session, int, error) {
	opts := make([]store.QueryOption, 0)
	if req.TenantID != "" {
		opts = append(opts, s.store.Options().InNamespace(req.TenantID))
	}

	opts = append(opts, s.store.Options().Sort(&query.Sorter{By: "started_at", Order: query.OrderDesc}))
	opts = append(opts, s.store.Options().Paginate(&req.Paginator))

	return s.store.SessionList(ctx, opts...)
}

func (s *service) GetSession(ctx context.Context, uid models.UID) (*models.Session, error) {
	session, err := s.store.SessionResolve(ctx, store.SessionUIDResolver, string(uid))
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
		IPAddress: session.IPAddress,
		Type:      session.Type,
		Term:      session.Term,
		Position: models.SessionPosition{
			Longitude: position.Longitude,
			Latitude:  position.Latitude,
		},
	})
	if err != nil {
		return nil, err
	}

	return s.store.SessionResolve(ctx, store.SessionUIDResolver, uid)
}

func (s *service) DeactivateSession(ctx context.Context, uid models.UID) error {
	sess, err := s.store.SessionResolve(ctx, store.SessionUIDResolver, string(uid))
	if err != nil {
		return NewErrSessionNotFound(uid, err)
	}

	return s.store.ActiveSessionDelete(ctx, models.UID(sess.UID))
}

func (s *service) KeepAliveSession(ctx context.Context, uid models.UID) error {
	session, err := s.store.SessionResolve(ctx, store.SessionUIDResolver, string(uid))
	if err != nil {
		return NewErrSessionNotFound(uid, err)
	}

	session.LastSeen = clock.Now()

	return s.store.SessionUpdate(ctx, session)
}

func (s *service) UpdateSession(ctx context.Context, uid models.UID, model models.SessionUpdate) error {
	session, err := s.store.SessionResolve(ctx, store.SessionUIDResolver, string(uid))
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

func (s *service) EventSession(ctx context.Context, uid models.UID, event *models.SessionEvent) error {
	if _, err := s.store.SessionResolve(ctx, store.SessionUIDResolver, string(uid)); err != nil {
		return NewErrSessionNotFound(uid, err)
	}

	return s.store.SessionEventsCreate(ctx, event)
}
