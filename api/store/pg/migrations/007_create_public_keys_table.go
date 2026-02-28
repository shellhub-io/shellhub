package migrations

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func init() {
	migrations.MustRegister(migration007Up, migration007Down)
}

func migration007Up(ctx context.Context, db *bun.DB) error {
	table := &struct {
		bun.BaseModel  `bun:"table:public_keys"`
		Fingerprint    string    `bun:"fingerprint,type:char(47),pk"`
		NamespaceID    string    `bun:"namespace_id,type:uuid,notnull"`
		CreatedAt      time.Time `bun:"created_at,type:timestamptz,notnull"`
		UpdatedAt      time.Time `bun:"updated_at,type:timestamptz,notnull"`
		Name           string    `bun:"name,type:varchar,notnull"`
		Username       string    `bun:"username,type:varchar,notnull,default:''"`
		Data           []byte    `bun:"data,type:bytea,nullzero"`
		FilterHostname string    `bun:"filter_hostname,type:varchar,default:''"`
	}{}

	if _, err := db.NewCreateTable().
		Model(table).
		IfNotExists().
		ForeignKey(`("namespace_id") REFERENCES namespaces("id") ON DELETE CASCADE`).
		Exec(ctx); err != nil {
		log.WithError(err).Error("failed to apply migration 007")

		return err
	}

	return nil
}

func migration007Down(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		DROP TABLE IF EXISTS public_keys;
	`)
	if err != nil {
		log.WithError(err).Error("failed to revert migration 007")

		return err
	}

	return nil
}
