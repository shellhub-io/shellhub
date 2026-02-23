package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
)

func (pg *Pg) MembershipInvitationCreate(ctx context.Context, invitation *models.MembershipInvitation) error {
	db := pg.GetConnection(ctx)

	now := clock.Now()
	invitation.ID = uuid.Generate()
	invitation.CreatedAt = now
	invitation.UpdatedAt = now
	invitation.StatusUpdatedAt = now

	e := entity.MembershipInvitationFromModel(invitation)

	if _, err := db.NewInsert().Model(e).Exec(ctx); err != nil {
		return fromSQLError(err)
	}

	return nil
}

func (pg *Pg) MembershipInvitationResolve(ctx context.Context, tenantID, userID string) (*models.MembershipInvitation, error) {
	db := pg.GetConnection(ctx)

	invitation := new(entity.MembershipInvitation)

	err := db.NewSelect().
		Model(invitation).
		Relation("Namespace").
		Relation("User").
		Relation("UserInvitation").
		Where("membership_invitation.tenant_id = ?", tenantID).
		Where("membership_invitation.user_id = ?", userID).
		Order("membership_invitation.id DESC").
		Limit(1).
		Scan(ctx)
	if err != nil {
		return nil, fromSQLError(err)
	}

	return entity.MembershipInvitationToModel(invitation), nil
}

func (pg *Pg) MembershipInvitationUpdate(ctx context.Context, invitation *models.MembershipInvitation) error {
	db := pg.GetConnection(ctx)

	invitation.UpdatedAt = clock.Now()

	e := entity.MembershipInvitationFromModel(invitation)

	r, err := db.NewUpdate().Model(e).WherePK().Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return nil
}
