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
	SetSessionAuthenticated(ctx context.Context, uid models.UID, authenticated bool) error
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

func (s *service) SetSessionAuthenticated(ctx context.Context, uid models.UID, authenticated bool) error {
	return s.store.SessionSetAuthenticated(ctx, uid, authenticated)
}
