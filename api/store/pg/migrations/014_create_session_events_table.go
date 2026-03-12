package migrations

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func init() {
	migrations.MustRegister(migration014Up, migration014Down)
}

func migration014Up(ctx context.Context, db *bun.DB) error {
	table := &struct {
		bun.BaseModel `bun:"table:session_events"`
		ID            string    `bun:"id,type:uuid,pk"`
		SessionID     string    `bun:"session_id,type:varchar(128),notnull"`
		Type          string    `bun:"type,type:varchar(64),notnull"`
		Seat          int       `bun:"seat,type:integer,notnull"`
		Data          string    `bun:"data,type:text,nullzero"`
		CreatedAt     time.Time `bun:"created_at,type:timestamptz,notnull,default:now()"`
	}{}

	if _, err := db.NewCreateTable().
		Model(table).
		IfNotExists().
		ForeignKey(`("session_id") REFERENCES sessions("id") ON DELETE CASCADE`).
		Exec(ctx); err != nil {
		log.WithError(err).Error("failed to apply migration 014")

		return err
	}

	_, err := db.NewCreateIndex().
		Model((*struct {
			bun.BaseModel `bun:"table:session_events"`
		})(nil)).
		Index("session_events_session_id_created_at_idx").
		Column("session_id", "created_at").
		Exec(ctx)
	if err != nil {
		log.WithError(err).Error("failed to create session_id_created_at index in migration 014")

		return err
	}

	_, err = db.NewCreateIndex().
		Model((*struct {
			bun.BaseModel `bun:"table:session_events"`
		})(nil)).
		Index("session_events_type_created_at_idx").
		Column("type", "created_at").
		Exec(ctx)
	if err != nil {
		log.WithError(err).Error("failed to create type_created_at index in migration 014")

		return err
	}

	_, err = db.NewCreateIndex().
		Model((*struct {
			bun.BaseModel `bun:"table:session_events"`
		})(nil)).
		Index("session_events_seat_idx").
		Column("seat").
		Exec(ctx)
	if err != nil {
		log.WithError(err).Error("failed to create seat index in migration 014")

		return err
	}

	return nil
}

func migration014Down(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		DROP TABLE IF EXISTS session_events;
	`)
	if err != nil {
		log.WithError(err).Error("failed to revert migration 014")

		return err
	}

	return nil
}
