package pg

import (
	"context"
	"strings"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
)

func (pg *Pg) UserInvitationsUpsert(ctx context.Context, email string) (string, error) {
	db := pg.GetConnection(ctx)

	now := clock.Now()
	normalizedEmail := strings.ToLower(email)

	invitation := &entity.UserInvitation{
		ID:          uuid.Generate(),
		Email:       normalizedEmail,
		Status:      "pending",
		Invitations: 1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result := &entity.UserInvitation{}
	err := db.NewInsert().
		Model(invitation).
		On("CONFLICT (email) DO UPDATE").
		Set("invitations = user_invitations.invitations + 1").
		Set("updated_at = EXCLUDED.updated_at").
		Returning("*").
		Scan(ctx, result)
	if err != nil {
		return "", fromSQLError(err)
	}

	return result.ID, nil
}

func (pg *Pg) UserInvitationGet(ctx context.Context, resolver store.UserInvitationResolver, value string) (*models.UserInvitation, error) {
	db := pg.GetConnection(ctx)

	e := new(entity.UserInvitation)
	q := db.NewSelect().Model(e)

	switch resolver {
	case store.UserInvitationIDResolver:
		q = q.Where("id = ?", value)
	case store.UserInvitationEmailResolver:
		q = q.Where("email = ?", value)
	}

	if err := q.Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.UserInvitationToModel(e), nil
}

func (pg *Pg) UserInvitationUpdate(ctx context.Context, invitation *models.UserInvitation) error {
	db := pg.GetConnection(ctx)

	e := entity.UserInvitationFromModel(invitation)

	r, err := db.NewUpdate().
		Model(e).
		Column("email", "status", "invitations", "updated_at").
		WherePK().
		Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	count, err := r.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return store.ErrNoDocuments
	}

	return nil
}
