package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/pg/entity"
)

// MembershipInvitationCreate implements [store.MembershipInvitationStore].
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

// MembershipInvitationResolve implements [store.MembershipInvitationStore].
func (pg *Pg) MembershipInvitationResolve(ctx context.Context, sc scope.Scope, userID string) (*models.MembershipInvitation, error) {
	db := pg.GetConnection(ctx)

	inv := new(entity.MembershipInvitation)

	query := db.NewSelect().
		Model(inv).
		Relation("Namespace").
		Relation("User").
		Relation("UserInvitation").
		Where("membership_invitation.user_id = ?", userID).
		Order("membership_invitation.created_at DESC").
		Limit(1)

	ctx = context.WithValue(ctx, CtxTableAlias, "membership_invitation")
	ctx = context.WithValue(ctx, CtxNamespaceColumn, "tenant_id")

	query, err := applyScopedOptions(ctx, query, sc)
	if err != nil {
		return nil, err
	}

	if err := query.Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.MembershipInvitationToModel(inv), nil
}

// MembershipInvitationResolveBySig implements [store.MembershipInvitationStore].
func (pg *Pg) MembershipInvitationResolveBySig(ctx context.Context, sig string) (*models.MembershipInvitation, error) {
	db := pg.GetConnection(ctx)

	inv := new(entity.MembershipInvitation)

	err := db.NewSelect().
		Model(inv).
		Relation("Namespace").
		Relation("User").
		Relation("UserInvitation").
		Where("membership_invitation.sig = ?", sig).
		Where("membership_invitation.status = ?", models.MembershipInvitationStatusPending).
		Where("membership_invitation.expires_at > ?", clock.Now()).
		Limit(1).
		Scan(ctx)
	if err != nil {
		return nil, fromSQLError(err)
	}

	return entity.MembershipInvitationToModel(inv), nil
}

// MembershipInvitationUpdate implements [store.MembershipInvitationStore].
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

// MembershipInvitationDelete implements [store.MembershipInvitationStore].
func (pg *Pg) MembershipInvitationDelete(ctx context.Context, invitation *models.MembershipInvitation) error {
	db := pg.GetConnection(ctx)

	e := entity.MembershipInvitationFromModel(invitation)

	r, err := db.NewDelete().Model(e).WherePK().Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

// UserMembershipInvitationList implements [store.MembershipInvitationStore].
func (pg *Pg) UserMembershipInvitationList(ctx context.Context, userID string, opts ...store.QueryOption) ([]models.MembershipInvitation, int64, error) {
	db := pg.GetConnection(ctx)

	entities := make([]entity.MembershipInvitation, 0)
	q := db.NewSelect().
		Model(&entities).
		Relation("Namespace").
		Relation("User").
		Relation("UserInvitation").
		Where("membership_invitation.user_id = ?", userID)

	var err error
	ctx = context.WithValue(ctx, CtxTableAlias, "membership_invitation")
	q, err = ApplyOptions(ctx, q, opts...)
	if err != nil {
		return nil, 0, err
	}

	count, err := q.ScanAndCount(ctx)
	if err != nil {
		return nil, 0, fromSQLError(err)
	}

	if count == 0 {
		return []models.MembershipInvitation{}, 0, nil
	}

	invitations := make([]models.MembershipInvitation, len(entities))
	for i, e := range entities {
		invitations[i] = *entity.MembershipInvitationToModel(&e)
	}

	return invitations, int64(count), nil
}

// NamespaceMembershipInvitationList implements [store.MembershipInvitationStore].
func (pg *Pg) NamespaceMembershipInvitationList(ctx context.Context, sc scope.Scope, opts ...store.QueryOption) ([]models.MembershipInvitation, int64, error) {
	db := pg.GetConnection(ctx)

	entities := make([]entity.MembershipInvitation, 0)
	q := db.NewSelect().
		Model(&entities).
		Relation("Namespace").
		Relation("User").
		Relation("UserInvitation")

	var err error
	ctx = context.WithValue(ctx, CtxTableAlias, "membership_invitation")
	ctx = context.WithValue(ctx, CtxNamespaceColumn, "tenant_id")

	q, err = applyScopedOptions(ctx, q, sc, opts...)
	if err != nil {
		return nil, 0, err
	}

	count, err := q.ScanAndCount(ctx)
	if err != nil {
		return nil, 0, fromSQLError(err)
	}

	if count == 0 {
		return []models.MembershipInvitation{}, 0, nil
	}

	invitations := make([]models.MembershipInvitation, len(entities))
	for i, e := range entities {
		invitations[i] = *entity.MembershipInvitationToModel(&e)
	}

	return invitations, int64(count), nil
}
