package services

import (
	"context"

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
	return nil, 0, nil
}

func (s *service) GetSession(ctx context.Context, uid models.UID) (*models.Session, error) {
	return nil, nil
}

func (s *service) CreateSession(ctx context.Context, session requests.SessionCreate) (*models.Session, error) {
	return nil, nil
}

func (s *service) DeactivateSession(ctx context.Context, uid models.UID) error {
	return nil
}

func (s *service) KeepAliveSession(ctx context.Context, uid models.UID) error {
	return nil
}

func (s *service) UpdateSession(ctx context.Context, uid models.UID, model models.SessionUpdate) error {
	return nil
}

func (s *service) EventSession(ctx context.Context, uid models.UID, event *models.SessionEvent) error {
	return nil
}
