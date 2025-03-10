package options

import (
	"context"
	"database/sql"
)

type Option func(ctx context.Context, db *sql.DB) error
