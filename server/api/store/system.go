package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

// SystemStore persists the instance-wide state: whether setup has run, and which login
// methods are offered.
type SystemStore interface {
	SystemGet(ctx context.Context) (*models.System, error)
	SystemSet(ctx context.Context, system *models.System) error

	// SystemLock blocks until no other transaction holds the instance-wide system lock, then
	// holds it until the transaction in ctx ends. It returns ErrLockOutsideTransaction when ctx
	// carries no transaction, ctx's error when ctx ends while waiting, and an error wrapping
	// ErrInternal when the lock query fails otherwise.
	SystemLock(ctx context.Context) error
}
