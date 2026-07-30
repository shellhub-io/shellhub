package options

import (
	"context"

	"github.com/uptrace/bun"
)

type Option func(ctx context.Context, db *bun.DB) error
