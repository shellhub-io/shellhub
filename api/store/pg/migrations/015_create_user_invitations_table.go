package migrations

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func init() {
	migrations.MustRegister(migration015Up, migration015Down)
}

func migration015Up(ctx context.Context, db *bun.DB) error {
	table := &struct {
		bun.BaseModel `bun:"table:user_invitations"`
		ID            string    `bun:"id,type:uuid,pk,default:gen_random_uuid()"`
		Email         string    `bun:"email,type:varchar(254),notnull,unique"`
		Status        string    `bun:"status,type:varchar(32),notnull,default:'pending'"`
		Invitations   int       `bun:"invitations,notnull,default:1"`
		CreatedAt     time.Time `bun:"created_at,type:timestamptz,notnull"`
		UpdatedAt     time.Time `bun:"updated_at,type:timestamptz,notnull"`
	}{}

	if _, err := db.NewCreateTable().
		Model(table).
		IfNotExists().
		Exec(ctx); err != nil {
		log.WithError(err).Error("failed to apply migration 015")

		return err
	}

	_, err := db.NewCreateIndex().
		Model((*struct {
			bun.BaseModel `bun:"table:user_invitations"`
		})(nil)).
		Index("user_invitations_email_idx").
		Column("email").
		Exec(ctx)
	if err != nil {
		log.WithError(err).Error("failed to create email index in migration 015")

		return err
	}

	return nil
}

func migration015Down(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS user_invitations;`)
	if err != nil {
		log.WithError(err).Error("failed to revert migration 015")

		return err
	}

	return nil
}
