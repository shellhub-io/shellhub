package migrations

import (
	"github.com/uptrace/bun/migrate"
)

var migrations = migrate.NewMigrations()

func FetchMigrations() (*migrate.Migrations, error) {
	if err := migrations.DiscoverCaller(); err != nil {
		return nil, err
	}

	return migrations, nil
}
