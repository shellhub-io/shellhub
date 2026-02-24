package migrations

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func init() {
	migrations.MustRegister(migration008Up, migration008Down)
}

func migration008Up(ctx context.Context, db *bun.DB) error {
	table := &struct {
		bun.BaseModel `bun:"table:tags"`
		ID            string    `bun:"id,type:uuid,pk"`
		NamespaceID   string    `bun:"namespace_id,type:uuid,notnull"`
		Name          string    `bun:"name,type:varchar,notnull"`
		CreatedAt     time.Time `bun:"created_at,type:timestamptz,notnull"`
		UpdatedAt     time.Time `bun:"updated_at,type:timestamptz,notnull"`
	}{}

	if _, err := db.NewCreateTable().
		Model(table).
		IfNotExists().
		ForeignKey(`("namespace_id") REFERENCES namespaces("id") ON DELETE CASCADE`).
		Exec(ctx); err != nil {
		log.WithError(err).Error("failed to apply migration 008")

		return err
	}

	_, err := db.NewCreateIndex().
		Model((*struct {
			bun.BaseModel `bun:"table:tags"`
		})(nil)).
		Index("tags_namespace_id_name_unique").
		Column("namespace_id", "name").
		Unique().
		Exec(ctx)
	if err != nil {
		log.WithError(err).Error("failed to apply migration 008")

		return err
	}

	return nil
}

func migration008Down(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, "DROP TABLE IF EXISTS tags")
	if err != nil {
		log.WithError(err).Error("failed to revert migration 008")

		return err
	}

	return nil
}
