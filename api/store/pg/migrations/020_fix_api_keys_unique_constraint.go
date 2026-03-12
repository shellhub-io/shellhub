package migrations

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func init() {
	migrations.MustRegister(migration020Up, migration020Down)
}

func migration020Up(ctx context.Context, db *bun.DB) error {
	if _, err := db.ExecContext(ctx, `ALTER TABLE api_keys DROP CONSTRAINT IF EXISTS api_keys_name_key`); err != nil {
		log.WithError(err).Error("failed to drop api_keys_name_key constraint in migration 020")

		return err
	}

	if _, err := db.NewCreateIndex().
		Model((*struct {
			bun.BaseModel `bun:"table:api_keys"`
		})(nil)).
		Index("api_keys_namespace_id_name_unique").
		Column("namespace_id", "name").
		Unique().
		IfNotExists().
		Exec(ctx); err != nil {
		log.WithError(err).Error("failed to create api_keys_namespace_id_name_unique index in migration 020")

		return err
	}

	return nil
}

// migration020Down drops the composite unique index but intentionally does not
// restore the original api_keys_name_key constraint, which enforced global
// uniqueness on name — a bug, not intended behavior.
func migration020Down(ctx context.Context, db *bun.DB) error {
	if _, err := db.ExecContext(ctx, `DROP INDEX IF EXISTS api_keys_namespace_id_name_unique`); err != nil {
		log.WithError(err).Error("failed to revert migration 020")

		return err
	}

	return nil
}
