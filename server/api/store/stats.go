package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type StatsStore interface {
	// GetStats retrieves device and session statistics within the given namespace scope. An
	// unbounded scope returns instance-wide statistics.
	GetStats(ctx context.Context, sc scope.Scope) (*models.Stats, error)
}
