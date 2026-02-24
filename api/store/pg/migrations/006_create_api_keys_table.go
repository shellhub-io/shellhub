package migrations

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func init() {
	migrations.MustRegister(migration006Up, migration006Down)
}

func migration006Up(ctx context.Context, db *bun.DB) error {
	table := &struct {
		bun.BaseModel `bun:"table:api_keys"`
		KeyDigest     string    `bun:"key_digest,type:char(64),notnull,pk"`
		NamespaceID   string    `bun:"namespace_id,type:uuid,notnull,pk"`
		CreatedAt     time.Time `bun:"created_at,type:timestamptz,notnull"`
		UpdatedAt     time.Time `bun:"updated_at,type:timestamptz,notnull"`
		ExpiresIn     int64     `bun:"expires_in,type:bigint,nullzero"`
		Name          string    `bun:"name,type:varchar,notnull,unique"`
		Role          string    `bun:"role,type:membership_role,notnull"`
		UserID        string    `bun:"user_id,type:uuid,notnull"`
	}{}

	if _, err := db.NewCreateTable().
		Model(table).
		IfNotExists().
		ForeignKey(`("namespace_id") REFERENCES namespaces("id") ON DELETE CASCADE`).
		Exec(ctx); err != nil {
		log.WithError(err).Error("failed to apply migration 006")

		return err
	}

	return nil
}

func migration006Down(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		DROP TABLE IF EXISTS api_keys;
	`)
	if err != nil {
		log.WithError(err).Error("failed to revert migration 006")

		return err
	}

	return nil
}
