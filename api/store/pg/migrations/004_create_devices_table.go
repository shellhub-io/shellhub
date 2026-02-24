package migrations

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func init() {
	migrations.MustRegister(migration004Up, migration004Down)
}

func migration004Up(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		DROP TYPE IF EXISTS device_status;
		CREATE TYPE device_status AS ENUM ('accepted', 'pending', 'rejected', 'removed', 'unused');
	`)
	if err != nil {
		return err
	}

	deviceTable := &struct {
		bun.BaseModel  `bun:"table:devices"`
		ID             string     `bun:"id,type:varchar,pk"`
		NamespaceID    string     `bun:"namespace_id,type:uuid,notnull"`
		CreatedAt      time.Time  `bun:"created_at,type:timestamptz,notnull"`
		UpdatedAt      time.Time  `bun:"updated_at,type:timestamptz,notnull"`
		RemovedAt      *time.Time `bun:"removed_at,type:timestamptz"`
		LastSeen       time.Time  `bun:"last_seen,type:timestamptz,notnull"`
		DisconnectedAt time.Time  `bun:"disconnected_at,type:timestamptz,nullzero"`
		Status         string     `bun:"status,type:device_status,notnull"`
		Name           string     `bun:"name,type:varchar(64),notnull"`
		Mac            string     `bun:"mac,type:varchar(17),notnull"`
		PublicKey      string     `bun:"public_key,type:text,notnull"`
		Identifier     string     `bun:"identifier,type:varchar,nullzero"`
		PrettyName     string     `bun:"pretty_name,type:varchar(64),nullzero"`
		Version        string     `bun:"version,type:varchar(32),nullzero"`
		Arch           string     `bun:"arch,type:varchar(16),nullzero"`
		Platform       string     `bun:"platform,type:varchar(32),nullzero"`
		Latitude       float64    `bun:"latitude,type:numeric,nullzero"`
		Longitude      float64    `bun:"longitude,type:numeric,nullzero"`
	}{}

	_, err = db.NewCreateTable().
		Model(deviceTable).
		IfNotExists().
		ForeignKey(`("namespace_id") REFERENCES namespaces("id") ON DELETE CASCADE`).
		Exec(ctx)
	if err != nil {
		log.WithError(err).Error("failed to apply migration 004")

		return err
	}

	_, err = db.NewCreateIndex().
		Model((*struct {
			bun.BaseModel `bun:"table:devices"`
		})(nil)).
		Index("devices_namespace_id").
		Column("namespace_id").
		Exec(ctx)
	if err != nil {
		log.WithError(err).Error("failed to apply migration 004")

		return err
	}

	_, err = db.NewCreateIndex().
		Model((*struct {
			bun.BaseModel `bun:"table:devices"`
		})(nil)).
		Index("devices_last_seen").
		Column("last_seen").
		Exec(ctx)
	if err != nil {
		log.WithError(err).Error("failed to apply migration 004")

		return err
	}

	_, err = db.NewCreateIndex().
		Model((*struct {
			bun.BaseModel `bun:"table:devices"`
		})(nil)).
		Index("devices_disconnected_at").
		Column("disconnected_at").
		Exec(ctx)
	if err != nil {
		log.WithError(err).Error("failed to apply migration 004")

		return err
	}

	return nil
}

func migration004Down(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		DROP TABLE IF EXISTS devices;
		DROP TYPE IF EXISTS device_status;
	`)
	if err != nil {
		log.WithError(err).Error("failed to revert migration 004")

		return err
	}

	return nil
}
