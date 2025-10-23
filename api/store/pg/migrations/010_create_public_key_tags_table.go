package migrations

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func init() {
	migrations.MustRegister(migration010Up, migration010Down)
}

func migration010Up(ctx context.Context, db *bun.DB) error {
	publicKeyTagsTable := &struct {
		bun.BaseModel `bun:"table:public_key_tags"`
		PublicKeyID   string    `bun:"public_key_id,type:uuid,pk"`
		TagID         string    `bun:"tag_id,type:uuid,pk"`
		CreatedAt     time.Time `bun:"created_at,type:timestamptz,notnull"`
	}{}

	if _, err := db.NewCreateTable().
		Model(publicKeyTagsTable).
		IfNotExists().
		ForeignKey(`("public_key_id") REFERENCES public_keys("id") ON DELETE CASCADE`).
		ForeignKey(`("tag_id") REFERENCES tags("id") ON DELETE CASCADE`).
		Exec(ctx); err != nil {
		log.WithError(err).Error("failed to create public_key_tags table in migration 010")

		return err
	}

	if _, err := db.NewCreateIndex().
		Model(publicKeyTagsTable).
		Index("public_key_tags_public_key_id").
		Column("public_key_id").
		Exec(ctx); err != nil {
		log.WithError(err).Error("failed to create public_key_id index for public_key_tags in migration 010")

		return err
	}

	if _, err := db.NewCreateIndex().
		Model(publicKeyTagsTable).
		Index("public_key_tags_tag_id").
		Column("tag_id").
		Exec(ctx); err != nil {
		log.WithError(err).Error("failed to create tag_id index for public_key_tags in migration 010")

		return err
	}

	return nil
}

func migration010Down(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, "DROP TABLE IF EXISTS public_key_tags")

	return err
}
