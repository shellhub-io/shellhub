package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

// SystemStore persists the instance-wide state: whether setup has run, and which login
// methods are offered.
type SystemStore interface {
	SystemGet(ctx context.Context) (*models.System, error)
	// SystemSet writes the instance state and returns the database's error when the row cannot
	// be read or written. It never writes InstanceID or Telemetry: the instance ID is generated
	// once, and the telemetry setting has its own write path.
	SystemSet(ctx context.Context, system *models.System) error

	// SystemLock blocks until no other transaction holds the instance-wide system lock, then
	// holds it until the transaction in ctx ends. It returns ErrLockOutsideTransaction when ctx
	// carries no transaction, ctx's error when ctx ends while waiting, and an error wrapping
	// ErrInternal when the lock query fails otherwise.
	SystemLock(ctx context.Context) error
}
