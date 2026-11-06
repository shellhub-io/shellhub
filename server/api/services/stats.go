package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// StatsService reports the counts shown on the dashboard.
type StatsService interface {
	// GetStats returns the device and session counts within sc, or the store's error. An unbounded
	// sc aggregates every namespace, so only a caller entitled to the whole instance passes one.
	GetStats(ctx context.Context, sc scope.Scope) (*models.Stats, error)
}

func (s *service) GetStats(ctx context.Context, sc scope.Scope) (*models.Stats, error) {
	return s.store.GetStats(ctx, sc)
}
