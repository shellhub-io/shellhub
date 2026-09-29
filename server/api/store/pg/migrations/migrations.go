package migrations

import (
	"embed"

	"github.com/shellhub-io/shellhub/server/api/store/pg/migrator"
	"github.com/uptrace/bun/migrate"
)

// Migrations is bun's registry of the same migrations as [All], which bun's migrator reads their
// status from and rolls them back with. A Go migration has no down in it.
var Migrations = migrate.NewMigrations()

// All is every core migration in name order, for [migrator.Apply].
var All []migrator.Migration

// Tables are the tables the core migrations are recorded in.
var Tables = migrator.Tables{Migrations: "bun_migrations", Locks: "bun_migration_locks"}

//go:embed *.sql
var sqlMigrations embed.FS

var goMigrations = []migrator.Migration{
	migrator.Go("022", "vacuum_full_devices", vacuumFullDevices),
}

func init() {
	if err := Migrations.Discover(sqlMigrations); err != nil {
		panic(err)
	}

	for _, m := range goMigrations {
		Migrations.Add(migrate.Migration{Name: m.Name, Comment: m.Comment})
	}

	var err error
	if All, err = migrator.Load(sqlMigrations, goMigrations...); err != nil {
		panic(err)
	}
}

// FetchMigrations returns the registered migrations, in the order their filenames impose.
func FetchMigrations() *migrate.Migrations {
	return Migrations
}
