package services

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type SessionService interface {
	ListSessions(ctx context.Context, pagination paginator.Query) ([]models.Session, int, error)
	GetSession(ctx context.Context, uid models.UID) (*models.Session, error)
	CreateSession(ctx context.Context, session models.Session) (*models.Session, error)
	DeactivateSession(ctx context.Context, uid models.UID) error
	KeepAliveSession(ctx context.Context, uid models.UID) error
	SetSessionAuthenticated(ctx context.Context, uid models.UID, authenticated bool) error
}

func (s *service) ListSessions(ctx context.Context, pagination paginator.Query) ([]models.Session, int, error) {
	return s.store.SessionList(ctx, pagination)
}

func (s *service) GetSession(ctx context.Context, uid models.UID) (*models.Session, error) {
	return s.store.SessionGet(ctx, uid)
}

func (s *service) CreateSession(ctx context.Context, session models.Session) (*models.Session, error) {
	return s.store.SessionCreate(ctx, session)
}

func (s *service) DeactivateSession(ctx context.Context, uid models.UID) error {
	if err := s.store.SessionDeleteActives(ctx, uid); err != nil && err == store.ErrNoDocuments {
		return ErrNotFound
	} else {
		return err
	}
}

func (s *service) KeepAliveSession(ctx context.Context, uid models.UID) error {
	return s.store.SessionSetLastSeen(ctx, uid)
}

func (s *service) SetSessionAuthenticated(ctx context.Context, uid models.UID, authenticated bool) error {
	return s.store.SessionSetAuthenticated(ctx, uid, authenticated)
}
