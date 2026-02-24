package migrations

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func init() {
	migrations.MustRegister(migration016Up, migration016Down)
}

func migration016Up(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		DROP TYPE IF EXISTS membership_invitation_status;
		CREATE TYPE membership_invitation_status AS ENUM ('pending', 'accepted', 'rejected', 'cancelled');
	`)
	if err != nil {
		return err
	}

	table := &struct {
		bun.BaseModel   `bun:"table:membership_invitations"`
		ID              string     `bun:"id,type:uuid,pk,default:gen_random_uuid()"`
		TenantID        string     `bun:"tenant_id,type:uuid,notnull"`
		UserID          string     `bun:"user_id,type:uuid,notnull"`
		InvitedBy       string     `bun:"invited_by,type:uuid,notnull"`
		Role            string     `bun:"role,type:varchar(32),notnull"`
		Status          string     `bun:"status,type:membership_invitation_status,notnull,default:'pending'"`
		StatusUpdatedAt time.Time  `bun:"status_updated_at,type:timestamptz,notnull"`
		ExpiresAt       *time.Time `bun:"expires_at,type:timestamptz,nullzero"`
		Invitations     int        `bun:"invitations,notnull,default:1"`
		CreatedAt       time.Time  `bun:"created_at,type:timestamptz,notnull"`
		UpdatedAt       time.Time  `bun:"updated_at,type:timestamptz,notnull"`
	}{}

	if _, err := db.NewCreateTable().
		Model(table).
		IfNotExists().
		ForeignKey(`("tenant_id") REFERENCES namespaces("id") ON DELETE CASCADE`).
		Exec(ctx); err != nil {
		log.WithError(err).Error("failed to apply migration 016")

		return err
	}

	_, err = db.NewCreateIndex().
		Model((*struct {
			bun.BaseModel `bun:"table:membership_invitations"`
		})(nil)).
		Index("membership_invitations_tenant_user_idx").
		Column("tenant_id", "user_id").
		Exec(ctx)
	if err != nil {
		log.WithError(err).Error("failed to create tenant_user index in migration 016")

		return err
	}

	_, err = db.NewCreateIndex().
		Model((*struct {
			bun.BaseModel `bun:"table:membership_invitations"`
		})(nil)).
		Index("membership_invitations_status_idx").
		Column("status").
		Exec(ctx)
	if err != nil {
		log.WithError(err).Error("failed to create status index in migration 016")

		return err
	}

	return nil
}

func migration016Down(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		DROP TABLE IF EXISTS membership_invitations;
		DROP TYPE IF EXISTS membership_invitation_status;
	`)
	if err != nil {
		log.WithError(err).Error("failed to revert migration 016")

		return err
	}

	return nil
}
