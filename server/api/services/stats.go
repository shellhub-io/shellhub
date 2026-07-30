package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type StatsService interface {
	GetStats(ctx context.Context, req *requests.GetStats) (*models.Stats, error)
}

func (s *service) GetStats(ctx context.Context, req *requests.GetStats) (*models.Stats, error) {
	return s.store.GetStats(ctx, req.TenantID)
}
