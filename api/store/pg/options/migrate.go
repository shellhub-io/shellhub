package options

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store/pg/migrations"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
)

func Migrate() Option {
	return func(ctx context.Context, db *bun.DB) error {
		log.Info("starting database migration")

		migrator := migrate.NewMigrator(db, migrations.FetchMigrations())
		if err := migrator.Init(context.Background()); err != nil {
			log.WithError(err).Error("failed to start migrations tables")

			return err
		}

		if err := migrator.Lock(ctx); err != nil {
			log.WithError(err).Error("failed to acquire migration lock")

			return err
		}

		defer func() {
			if err := migrator.Unlock(ctx); err != nil {
				log.WithError(err).Error("failed to release migration lock")
			} else {
				log.Debug("migration lock released successfully")
			}
		}()

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
	}
}
