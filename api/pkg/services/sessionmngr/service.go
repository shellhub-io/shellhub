package sessionmngr

import (
	"context"

	"github.com/shellhub-io/shellhub/api/pkg/models"
	"github.com/shellhub-io/shellhub/api/pkg/store"
)

type Service interface {
	ListSessions(ctx context.Context) ([]models.Session, error)
	CreateSession(ctx context.Context, session models.Session) (*models.Session, error)
	DeactivateSession(ctx context.Context, uid models.UID) error
}

type service struct {
	store store.Store
}

func NewService(store store.Store) Service {
	return &service{store}
}

func (s *service) ListSessions(ctx context.Context) ([]models.Session, error) {
	return nil, nil
	//	return s.store.ListDevices(ctx)
}

func (s *service) CreateSession(ctx context.Context, session models.Session) (*models.Session, error) {
	return s.store.CreateSession(ctx, session)
}

func (s *service) DeactivateSession(ctx context.Context, uid models.UID) error {
	return s.store.DeactivateSession(ctx, uid)
}
