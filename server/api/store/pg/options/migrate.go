package options

import (
	"context"

	"github.com/shellhub-io/shellhub/server/api/store/pg/migrations"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
)

// Migrate runs the pending migrations at startup, so a deployment does not need a separate
// migration step. It holds [MigrationLockKey] while migrating, so a replica booting during another
// one's migration waits for it instead of failing.
func Migrate() Option {
	return func(ctx context.Context, db *bun.DB) error {
		return WithMigrationLock(ctx, db, MigrationLockKey, func(ctx context.Context) error {
			log.Info("starting database migration")

			migrator := migrate.NewMigrator(db, migrations.FetchMigrations())
			if err := migrator.Init(ctx); err != nil {
				log.WithError(err).Error("failed to start migrations tables")

				return err
			}

			group, err := migrator.Migrate(ctx)
			if err != nil {
				log.WithError(err).Error("migration failed")

				return err
			}

			if group.IsZero() {
				log.Info("no new migrations to run (database is up to date)")

				return nil
			}

			log.Info("migration completed successfully")

			return nil
		})
	}
}
