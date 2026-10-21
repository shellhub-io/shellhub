package store

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// StatsStore computes the counts shown on the dashboard and reported by telemetry.
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
	CountRegisteredDevices(ctx context.Context, sc scope.Scope) (int64, error)

	// InstanceUsage counts the namespaces, users and devices of the whole instance, and the
	// sessions started at or after since. It returns ctx's error when it is cancelled or its
	// deadline passes, and [ErrInternal] for any other failure.
	InstanceUsage(ctx context.Context, since time.Time) (*models.InstanceUsage, error)

	// NamespaceUsages counts, for every namespace, its members and devices and the sessions
	// started at or after since. The order of the result is unspecified. It returns ctx's error
	// when it is cancelled or its deadline passes, and [ErrInternal] for any other failure.
	NamespaceUsages(ctx context.Context, since time.Time) ([]models.NamespaceUsage, error)
}
