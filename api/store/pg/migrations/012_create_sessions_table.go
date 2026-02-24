package migrations

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func init() {
	migrations.MustRegister(migration012Up, migration012Down)
}

func migration012Up(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		DROP TYPE IF EXISTS session_type;
		CREATE TYPE session_type AS ENUM ('shell', 'exec', 'none');
	`)
	if err != nil {
		return err
	}

	table := &struct {
		bun.BaseModel `bun:"table:sessions"`
		ID            string    `bun:"id,type:char(64),pk"`
		DeviceID      string    `bun:"device_id,type:varchar,notnull"`
		Username      string    `bun:"username,type:varchar(64),notnull"`
		IPAddress     string    `bun:"ip_address,type:inet,notnull"`
		StartedAt     time.Time `bun:"started_at,type:timestamptz,notnull"`
		SeenAt        time.Time `bun:"seen_at,type:timestamptz,notnull"`
		Closed        bool      `bun:"closed,notnull,default:false"`
		Authenticated bool      `bun:"authenticated,notnull,default:false"`
		Recorded      bool      `bun:"recorded,notnull,default:false"`
		Type          string    `bun:"type,type:session_type,nullzero"`
		Term          string    `bun:"term,type:varchar(32),nullzero"`
		Longitude     float64   `bun:"longitude,type:numeric(10,7),nullzero"`
		Latitude      float64   `bun:"latitude,type:numeric(10,7),nullzero"`
		CreatedAt     time.Time `bun:"created_at,type:timestamptz,notnull"`
		UpdatedAt     time.Time `bun:"updated_at,type:timestamptz,notnull"`
	}{}

	if _, err := db.NewCreateTable().
		Model(table).
		IfNotExists().
		ForeignKey(`("device_id") REFERENCES devices("id") ON DELETE CASCADE`).
		Exec(ctx); err != nil {
		log.WithError(err).Error("failed to apply migration 012")

		return err
	}

	_, err = db.NewCreateIndex().
		Model((*struct {
			bun.BaseModel `bun:"table:sessions"`
		})(nil)).
		Index("sessions_device_id_idx").
		Column("device_id").
		Exec(ctx)
	if err != nil {
		log.WithError(err).Error("failed to create device_id index in migration 012")

		return err
	}

	_, err = db.NewCreateIndex().
		Model((*struct {
			bun.BaseModel `bun:"table:sessions"`
		})(nil)).
		Index("sessions_started_at_idx").
		Column("started_at").
		Exec(ctx)
	if err != nil {
		log.WithError(err).Error("failed to create started_at index in migration 012")

		return err
	}

	_, err = db.NewCreateIndex().
		Model((*struct {
			bun.BaseModel `bun:"table:sessions"`
		})(nil)).
		Index("sessions_username_idx").
		Column("username").
		Exec(ctx)
	if err != nil {
		log.WithError(err).Error("failed to create username index in migration 012")

		return err
	}

	_, err = db.NewCreateIndex().
		Model((*struct {
			bun.BaseModel `bun:"table:sessions"`
		})(nil)).
		Index("sessions_type_idx").
		Column("type").
		Exec(ctx)
	if err != nil {
		log.WithError(err).Error("failed to create type index in migration 012")

		return err
	}

	_, err = db.NewCreateIndex().
		Model((*struct {
			bun.BaseModel `bun:"table:sessions"`
		})(nil)).
		Index("sessions_closed_started_idx").
		Column("closed", "started_at").
		Exec(ctx)
	if err != nil {
		log.WithError(err).Error("failed to create closed_started_at index in migration 012")

		return err
	}

	return nil
}

func migration012Down(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		DROP TABLE IF EXISTS sessions;
		DROP TYPE IF EXISTS session_type;
	`)
	if err != nil {
		log.WithError(err).Error("failed to revert migration 012")

		return err
	}

	return nil
}
