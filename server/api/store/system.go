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
}
