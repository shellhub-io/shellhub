package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type StatsStore interface {
	GetStats(ctx context.Context) (*models.Stats, error)
}
