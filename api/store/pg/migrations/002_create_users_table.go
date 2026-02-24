package migrations

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func init() {
	migrations.MustRegister(migration002Up, migration002Down)
}

func migration002Up(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		DROP TYPE IF EXISTS user_origin;
		CREATE TYPE user_origin AS ENUM ('local', 'saml');

		DROP TYPE IF EXISTS user_status;
		CREATE TYPE user_status AS ENUM ('not-confirmed', 'confirmed');

		DROP TYPE IF EXISTS user_auth_method;
		CREATE TYPE user_auth_method AS ENUM ('local', 'saml');
	`)
	if err != nil {
		return err
	}

	table := &struct {
		bun.BaseModel           `bun:"table:users"`
		ID                      string    `bun:"id,type:uuid,pk"`
		CreatedAt               time.Time `bun:"created_at,type:timestamptz,notnull"`
		UpdatedAt               time.Time `bun:"updated_at,type:timestamptz,notnull"`
		LastLogin               time.Time `bun:"last_login,type:timestamptz,nullzero"`
		Origin                  string    `bun:"origin,type:user_origin,notnull"`
		ExternalID              string    `bun:"external_id,type:varchar,nullzero"`
		Status                  string    `bun:"status,type:user_status,notnull"`
		Name                    string    `bun:"name,type:varchar(64),notnull"`
		Username                string    `bun:"username,type:varchar(32),notnull,unique"`
		Email                   string    `bun:"email,type:varchar(320),notnull,unique"`
		SecurityEmail           string    `bun:"security_email,type:varchar(320),nullzero"`
		PasswordDigest          string    `bun:"password_digest,type:char(72),notnull"`
		AuthMethods             []string  `bun:"auth_methods,type:user_auth_method[],array,notnull"`
		NamespaceOwnershipLimit int       `bun:"namespace_ownership_limit,type:integer,notnull"`
		EmailMarketing          bool      `bun:"email_marketing,notnull,default:false"`
		PreferredNamespaceID    string    `bun:"preferred_namespace_id,type:uuid,nullzero"`
		Admin                   bool      `bun:"admin,notnull,default:false"`
	}{}

	if _, err := db.
		NewCreateTable().
		Model(table).
		IfNotExists().
		ForeignKey(`("preferred_namespace_id") REFERENCES namespaces("id") ON DELETE SET NULL`).
		Exec(ctx); err != nil {
		log.WithError(err).Error("failed to apply migration 002")

		return err
	}

	return nil
}

func migration002Down(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		DROP TABLE IF EXISTS users;
		DROP TYPE IF EXISTS user_origin;
		DROP TYPE IF EXISTS user_status;
		DROP TYPE IF EXISTS user_auth_method;
	`)
	if err != nil {
		log.WithError(err).Error("failed to revert migration 002")

		return err
	}

	return nil
}
