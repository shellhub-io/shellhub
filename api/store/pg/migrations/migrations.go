package migrations

import (
	"github.com/uptrace/bun/migrate"
)

var migrations = migrate.NewMigrations()

func FetchMigrations() *migrate.Migrations {
	return migrations
}
