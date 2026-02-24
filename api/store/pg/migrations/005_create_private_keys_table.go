package migrations

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func init() {
	migrations.MustRegister(migration005Up, migration005Down)
}

func migration005Up(ctx context.Context, db *bun.DB) error {
	table := &struct {
		bun.BaseModel `bun:"table:private_keys"`
		Fingerprint   string    `bun:"fingerprint,type:varchar,pk"`
		CreatedAt     time.Time `bun:"created_at,type:timestamptz,notnull"`
		UpdatedAt     time.Time `bun:"updated_at,type:timestamptz,notnull"`
		Data          []byte    `bun:"data,type:bytea,nullzero"`
	}{}

	if _, err := db.NewCreateTable().Model(table).IfNotExists().Exec(ctx); err != nil {
		log.WithError(err).Error("failed to apply migration 005")

		return err
	}

	return nil
}

func migration005Down(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		DROP TABLE IF EXISTS private_keys;
	`)
	if err != nil {
		log.WithError(err).Error("failed to revert migration 005")

		return err
	}

	return nil
}
