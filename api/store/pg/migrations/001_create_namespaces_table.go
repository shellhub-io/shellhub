package migrations

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func init() {
	migrations.MustRegister(migration001Up, migration001Down)
}

func migration001Up(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		DROP TYPE IF EXISTS namespace_scope;
		CREATE TYPE namespace_scope AS ENUM ('personal', 'team');
	`)
	if err != nil {
		return err
	}

	table := &struct {
		bun.BaseModel          `bun:"table:namespaces"`
		ID                     string    `bun:"id,type:uuid,pk"`
		CreatedAt              time.Time `bun:"created_at,type:timestamptz,notnull"`
		UpdatedAt              time.Time `bun:"updated_at,type:timestamptz,notnull"`
		Scope                  string    `bun:"scope,type:namespace_scope,notnull"`
		Name                   string    `bun:"name,type:varchar(64),notnull"`
		OwnerID                string    `bun:"owner_id,type:uuid,notnull"`
		MaxDevices             int       `bun:"max_devices,type:integer,notnull"`
		RecordSessions         bool      `bun:"record_sessions,notnull"`
		ConnectionAnnouncement string    `bun:"connection_announcement,type:text,nullzero"`
		DevicesAcceptedCount   int64     `bun:"devices_accepted_count,type:bigint,notnull,default:0"`
		DevicesPendingCount    int64     `bun:"devices_pending_count,type:bigint,notnull,default:0"`
		DevicesRejectedCount   int64     `bun:"devices_rejected_count,type:bigint,notnull,default:0"`
		DevicesRemovedCount    int64     `bun:"devices_removed_count,type:bigint,notnull,default:0"`
	}{}

	if _, err := db.NewCreateTable().Model(table).IfNotExists().Exec(ctx); err != nil {
		log.WithError(err).Error("failed to apply migration 001")

		return err
	}

	return nil
}

func migration001Down(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		DROP TABLE IF EXISTS namespaces;
		DROP TYPE IF EXISTS namespace_scope;
	`)
	if err != nil {
		log.WithError(err).Error("failed to revert migration 001")

		return err
	}

	return nil
}
