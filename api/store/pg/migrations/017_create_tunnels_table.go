package migrations

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func init() {
	migrations.MustRegister(migration017Up, migration017Down)
}

func migration017Up(ctx context.Context, db *bun.DB) error {
	table := &struct {
		bun.BaseModel `bun:"table:tunnels"`
		ID            string    `bun:"id,type:uuid,pk,default:gen_random_uuid()"`
		NamespaceID   string    `bun:"namespace_id,type:uuid,notnull"`
		DeviceID      string    `bun:"device_id,type:varchar,notnull"`
		Address       string    `bun:"address,type:varchar(255),notnull,unique"`
		Host          string    `bun:"host,type:varchar(255),notnull"`
		Port          int       `bun:"port,notnull"`
		TLSEnabled    bool      `bun:"tls_enabled,notnull,default:false"`
		TLSVerify     bool      `bun:"tls_verify,notnull,default:false"`
		TLSDomain     string    `bun:"tls_domain,type:varchar(255),nullzero"`
		CreatedAt     time.Time `bun:"created_at,type:timestamptz,notnull"`
		UpdatedAt     time.Time `bun:"updated_at,type:timestamptz,notnull"`
	}{}

	if _, err := db.NewCreateTable().
		Model(table).
		IfNotExists().
		ForeignKey(`("namespace_id") REFERENCES namespaces("id") ON DELETE CASCADE`).
		ForeignKey(`("device_id") REFERENCES devices("id") ON DELETE CASCADE`).
		Exec(ctx); err != nil {
		log.WithError(err).Error("failed to apply migration 017")

		return err
	}

	_, err := db.NewCreateIndex().
		Model((*struct {
			bun.BaseModel `bun:"table:tunnels"`
		})(nil)).
		Index("tunnels_namespace_device_idx").
		Column("namespace_id", "device_id").
		Exec(ctx)
	if err != nil {
		log.WithError(err).Error("failed to create namespace_device index in migration 017")

		return err
	}

	_, err = db.NewCreateIndex().
		Model((*struct {
			bun.BaseModel `bun:"table:tunnels"`
		})(nil)).
		Index("tunnels_address_idx").
		Column("address").
		Exec(ctx)
	if err != nil {
		log.WithError(err).Error("failed to create address index in migration 017")

		return err
	}

	return nil
}

func migration017Down(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS tunnels;`)
	if err != nil {
		log.WithError(err).Error("failed to revert migration 017")

		return err
	}

	return nil
}
