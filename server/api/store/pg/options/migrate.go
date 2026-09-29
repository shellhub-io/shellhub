package options

import (
	"context"

	"github.com/shellhub-io/shellhub/server/api/store/pg/migrations"
	"github.com/shellhub-io/shellhub/server/api/store/pg/migrator"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

// Migrate runs the pending migrations at startup, so a deployment does not need a separate
// migration step. It records them through [migrator.Apply], so a boot killed midway resumes from
// the first migration it did not finish. It holds [MigrationLockKey] while migrating, so a replica
// booting during another one's migration waits for it instead of failing.
func Migrate() Option {
	return func(ctx context.Context, db *bun.DB) error {
		return WithMigrationLock(ctx, db, MigrationLockKey, func(ctx context.Context) error {
			log.Info("starting database migration")

			applied, err := migrator.Apply(ctx, db, migrations.Tables, migrations.All)
			if err != nil {
				log.WithError(err).Error("migration failed")

				return err
			}

			if applied == 0 {
				log.Info("no new migrations to run (database is up to date)")

				return nil
			}

			log.Info("migration completed successfully")

			return nil
		})
	}
}
