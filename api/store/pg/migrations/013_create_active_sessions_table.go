package migrations

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func init() {
	migrations.MustRegister(migration013Up, migration013Down)
}

func migration013Up(ctx context.Context, db *bun.DB) error {
	table := &struct {
		bun.BaseModel `bun:"table:active_sessions"`
		SessionID     string    `bun:"session_id,type:char(64),pk"`
		SeenAt        time.Time `bun:"seen_at,type:timestamptz,notnull"`
		CreatedAt     time.Time `bun:"created_at,type:timestamptz,notnull"`
	}{}

	if _, err := db.NewCreateTable().
		Model(table).
		IfNotExists().
		ForeignKey(`("session_id") REFERENCES sessions("id") ON DELETE CASCADE`).
		Exec(ctx); err != nil {
		log.WithError(err).Error("failed to apply migration 013")

		return err
	}

	return nil
}

func migration013Down(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS active_sessions;`)
	if err != nil {
		log.WithError(err).Error("failed to revert migration 013")

		return err
	}

	return nil
}
