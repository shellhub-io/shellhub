package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type StatsService interface {
	GetStats(ctx context.Context) (*models.Stats, error)
}

func (s *service) GetStats(ctx context.Context) (*models.Stats, error) {
	return s.store.GetStats(ctx)
}
