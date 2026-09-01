package migrations

import (
	"embed"

	"github.com/uptrace/bun/migrate"
)

// Migrations is the registry the embedded .sql files register themselves into.
var Migrations = migrate.NewMigrations()

//go:embed *.sql
var sqlMigrations embed.FS

func init() {
	if err := Migrations.Discover(sqlMigrations); err != nil {
		panic(err)
	}
}

// FetchMigrations returns the registered migrations, in the order their filenames impose.
func FetchMigrations() *migrate.Migrations {
	return Migrations
}
