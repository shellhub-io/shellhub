package options

import (
	"context"

	"github.com/uptrace/bun"
)

// Option prepares the database connection at startup. Returning an error aborts startup.
type Option func(ctx context.Context, db *bun.DB) error
