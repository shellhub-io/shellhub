package migrations

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func init() {
	migrations.MustRegister(migration018Up, migration018Down)
}

func migration018Up(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `ALTER TABLE users ALTER COLUMN password_digest TYPE VARCHAR USING TRIM(password_digest);`)
	if err != nil {
		log.WithError(err).Error("failed to apply migration 018")

		return err
	}

	return nil
}

func migration018Down(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `ALTER TABLE users ALTER COLUMN password_digest TYPE CHAR(72);`)
	if err != nil {
		log.WithError(err).Error("failed to revert migration 018")

		return err
	}

	return nil
}
