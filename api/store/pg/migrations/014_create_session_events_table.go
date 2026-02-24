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
	_, err := db.ExecContext(ctx, `
		DROP TYPE IF EXISTS session_event_type;
		CREATE TYPE session_event_type AS ENUM (
			'pty-output', 'pty-req', 'window-change', 'exit-code',
			'exit-status', 'exit-signal', 'env', 'shell', 'exec',
			'subsystem', 'signal', 'tcpip-forward', 'auth-agent-req'
		);
	`)
	if err != nil {
		return err
	}

	table := &struct {
		bun.BaseModel `bun:"table:session_events"`
		ID            string    `bun:"id,type:uuid,pk"`
		SessionID     string    `bun:"session_id,type:char(64),notnull"`
		Type          string    `bun:"type,type:session_event_type,notnull"`
		Seat          int       `bun:"seat,type:integer,notnull"`
		Data          string    `bun:"data,type:jsonb,nullzero"`
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

	_, err = db.NewCreateIndex().
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

	_, err = db.ExecContext(ctx, `
		CREATE INDEX session_events_data_gin_idx ON session_events USING GIN (data);
	`)
	if err != nil {
		log.WithError(err).Error("failed to create data GIN index in migration 014")

		return err
	}

	return nil
}

func migration014Down(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		DROP TABLE IF EXISTS session_events;
		DROP TYPE IF EXISTS session_event_type;
	`)
	if err != nil {
		log.WithError(err).Error("failed to revert migration 014")

		return err
	}

	return nil
}
