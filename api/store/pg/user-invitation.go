package pg

import (
	"context"
	"strings"

	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/uuid"
)

func (pg *Pg) UserInvitationsUpsert(ctx context.Context, email string) (string, error) {
	db := pg.getConnection(ctx)

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

	_, err := db.NewInsert().
		Model(invitation).
		On("CONFLICT (email) DO UPDATE").
		Set("invitations = user_invitations.invitations + 1").
		Set("updated_at = EXCLUDED.updated_at").
		Returning("id").
		Exec(ctx)
	if err != nil {
		return "", fromSQLError(err)
	}

	return invitation.ID, nil
}
