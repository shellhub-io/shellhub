package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

// TODO: maybe systems config can be an file?

func (pg *pg) SystemGet(ctx context.Context) (*models.System, error) {
	return nil, nil
}

func (pg *pg) SystemSet(ctx context.Context, key string, value any) error {
	return nil
}
