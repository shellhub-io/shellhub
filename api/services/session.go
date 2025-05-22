package services

import (
	"context"
	"net"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type SessionService interface {
	ListSessions(ctx context.Context, paginator query.Paginator) ([]models.Session, int, error)
	GetSession(ctx context.Context, uid models.UID) (*models.Session, error)
	CreateSession(ctx context.Context, session requests.SessionCreate) (*models.Session, error)
	DeactivateSession(ctx context.Context, uid models.UID) error
	KeepAliveSession(ctx context.Context, uid models.UID) error
	UpdateSession(ctx context.Context, uid models.UID, model models.SessionUpdate) error
	EventSession(ctx context.Context, uid models.UID, event *models.SessionEvent) error
}

func (s *service) ListSessions(ctx context.Context, paginator query.Paginator) ([]models.Session, int, error) {
	return s.store.SessionList(ctx, paginator)
}

func (s *service) GetSession(ctx context.Context, uid models.UID) (*models.Session, error) {
	session, err := s.store.SessionGet(ctx, uid)
	if err != nil {
		return nil, NewErrSessionNotFound(uid, err)
	}

	return session, nil
}

func (s *service) CreateSession(ctx context.Context, session requests.SessionCreate) (*models.Session, error) {
	position, _ := s.locator.GetPosition(net.ParseIP(session.IPAddress))

	return s.store.SessionCreate(ctx, models.Session{
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
}

func (s *service) DeactivateSession(ctx context.Context, uid models.UID) error {
	err := s.store.SessionDeleteActives(ctx, uid)
	if err == store.ErrNoDocuments {
		return NewErrSessionNotFound(uid, err)
	}

	return err
}

func (s *service) KeepAliveSession(ctx context.Context, uid models.UID) error {
	return s.store.SessionSetLastSeen(ctx, uid)
}

func (s *service) UpdateSession(ctx context.Context, uid models.UID, model models.SessionUpdate) error {
	sess, err := s.store.SessionGet(ctx, uid)
	if err != nil {
		return NewErrSessionNotFound(uid, err)
	}

	// FIXME: The store function that retrieves a session sets the `Active` field based on the existence of a document
	// in another collection. When that document exists, `Active` is set to `true`. However since we reuse the same
	// [models.Session] struct and modify it in the update method, it is persisted in the database wrongly. Something
	// similar happens with the field `Device`, but it is ignored by the BSON, it isn't persistented on the database.
	sess.Active = false

	var insertActiveSession bool

	if model.Authenticated != nil {
		if !sess.Authenticated {
			sess.Authenticated = *model.Authenticated

			insertActiveSession = true
		}
	}

	if model.Type != nil {
		sess.Type = *model.Type
	}

	if model.Recorded != nil {
		sess.Recorded = *model.Recorded
	}

	if err := s.store.SessionUpdate(ctx, uid, sess); err != nil {
		return err
	}

	if insertActiveSession {
		return s.store.SessionActiveCreate(ctx, uid, sess)
	}

	return nil
}

func (s *service) EventSession(ctx context.Context, uid models.UID, event *models.SessionEvent) error {
	sess, err := s.store.SessionGet(ctx, uid)
	if err != nil {
		return NewErrSessionNotFound(uid, err)
	}

	return s.store.SessionEvent(ctx, models.UID(sess.UID), event)
}
