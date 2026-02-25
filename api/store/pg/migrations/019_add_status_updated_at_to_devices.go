package migrations

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func init() {
	migrations.MustRegister(migration019Up, migration019Down)
}

func migration019Up(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `ALTER TABLE devices ADD COLUMN status_updated_at TIMESTAMPTZ NOT NULL DEFAULT '0001-01-01 00:00:00+00';`)
	if err != nil {
		log.WithError(err).Error("failed to apply migration 019")

		return err
	}

	_, err = db.ExecContext(ctx, `UPDATE devices SET status_updated_at = created_at WHERE status_updated_at = '0001-01-01 00:00:00+00';`)
	if err != nil {
		log.WithError(err).Error("failed to backfill status_updated_at in migration 019")

		return err
	}

	return nil
}

func migration019Down(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `ALTER TABLE devices DROP COLUMN status_updated_at;`)
	if err != nil {
		log.WithError(err).Error("failed to revert migration 019")

		return err
	}

	return nil
}
