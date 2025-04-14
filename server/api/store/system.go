package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type SystemStore interface {
	SystemGet(ctx context.Context) (*models.System, error)
	SystemSet(ctx context.Context, key string, value any) error
}
