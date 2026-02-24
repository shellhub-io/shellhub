package migrations

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func init() {
	migrations.MustRegister(migration009Up, migration009Down)
}

func migration009Up(ctx context.Context, db *bun.DB) error {
	deviceTagsTable := &struct {
		bun.BaseModel `bun:"table:device_tags"`
		DeviceID      string    `bun:"device_id,type:varchar,pk"`
		TagID         string    `bun:"tag_id,type:uuid,pk"`
		CreatedAt     time.Time `bun:"created_at,type:timestamptz,notnull"`
	}{}

	if _, err := db.NewCreateTable().
		Model(deviceTagsTable).
		IfNotExists().
		ForeignKey(`("device_id") REFERENCES devices("id") ON DELETE CASCADE`).
		ForeignKey(`("tag_id") REFERENCES tags("id") ON DELETE CASCADE`).
		Exec(ctx); err != nil {
		log.WithError(err).Error("failed to create device_tags table in migration 009")

		return err
	}

	if _, err := db.NewCreateIndex().
		Model(deviceTagsTable).
		Index("device_tags_device_id").
		Column("device_id").
		Exec(ctx); err != nil {
		log.WithError(err).Error("failed to create device_id index for device_tags in migration 009")

		return err
	}

	if _, err := db.NewCreateIndex().
		Model(deviceTagsTable).
		Index("device_tags_tag_id").
		Column("tag_id").
		Exec(ctx); err != nil {
		log.WithError(err).Error("failed to create tag_id index for device_tags in migration 009")

		return err
	}

	return nil
}

func migration009Down(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		DROP TABLE IF EXISTS public_key_tags;
		DROP TABLE IF EXISTS device_tags;
	`)
	if err != nil {
		log.WithError(err).Error("failed to revert migration 009")

		return err
	}

	return nil
}
