package migrations

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func init() {
	migrations.MustRegister(migration003Up, migration003Down)
}

func migration003Up(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		DROP TYPE IF EXISTS membership_role;
		CREATE TYPE membership_role AS ENUM ('owner', 'administrator', 'operator', 'observer');
	`)
	if err != nil {
		return err
	}

	table := &struct {
		bun.BaseModel `bun:"table:memberships"`
		UserID        string    `bun:"user_id,type:uuid,notnull,pk"`
		NamespaceID   string    `bun:"namespace_id,type:uuid,notnull,pk"`
		CreatedAt     time.Time `bun:"created_at,type:timestamptz,notnull"`
		UpdatedAt     time.Time `bun:"updated_at,type:timestamptz,notnull"`
		Role          string    `bun:"role,type:membership_role,notnull"`
	}{}

	if _, err := db.NewCreateTable().
		Model(table).
		IfNotExists().
		ForeignKey(`("user_id") REFERENCES users("id") ON DELETE CASCADE`).
		ForeignKey(`("namespace_id") REFERENCES namespaces("id") ON DELETE CASCADE`).
		Exec(ctx); err != nil {
		log.WithError(err).Error("failed to apply migration 003")

		return err
	}

	return nil
}

func migration003Down(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		DROP TABLE IF EXISTS memberships;
		DROP TYPE IF EXISTS membership_role;
	`)
	if err != nil {
		log.WithError(err).Error("failed to revert migration 003")

		return err
	}

	return nil
}
