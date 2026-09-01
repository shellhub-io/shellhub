package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// StatsStore computes the counts shown on the dashboard.
type StatsStore interface {
	// GetStats retrieves device and session statistics within the given namespace scope. An
	// unbounded scope returns instance-wide statistics.
	GetStats(ctx context.Context, sc scope.Scope) (*models.Stats, error)

	// CountRegisteredDevices reports how many devices are accepted within the given namespace
	// scope. An unbounded scope counts across the whole instance.
	//
	// It exists alongside GetStats for callers that need this count alone: GetStats answers five
	// questions, four of them by scanning the devices table, and a caller that wants one of them
	// pays for all five.
	CountRegisteredDevices(ctx context.Context, sc scope.Scope) (int, error)
}
