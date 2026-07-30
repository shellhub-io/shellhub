package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type StatsStore interface {
	// GetStats retrieves device and session statistics. If tenantID is provided,
	// statistics are filtered to that tenant. If empty, returns global statistics.
	GetStats(ctx context.Context, tenantID string) (*models.Stats, error)
}
