package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

func (pg *Pg) SystemGet(ctx context.Context) (*models.System, error) {
	return nil, nil
}

func (pg *Pg) SystemSet(ctx context.Context, key string, value any) error {
	return nil
}
