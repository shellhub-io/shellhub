package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (pg *Pg) NamespaceCreateMembership(ctx context.Context, tenantID string, membership *models.Member) error {
	db := pg.GetConnection(ctx)

	membership.AddedAt = clock.Now()
	entity := entity.MembershipFromModel(tenantID, membership)
	if _, err := db.NewInsert().Model(entity).Exec(ctx); err != nil {
		return fromSQLError(err)
	}

	return nil
}

func (pg *Pg) NamespaceUpdateMembership(ctx context.Context, tenantID string, member *models.Member) error {
	db := pg.GetConnection(ctx)

	e := entity.MembershipFromModel(tenantID, member)
	e.UpdatedAt = clock.Now()
	r, err := db.NewUpdate().Model(e).WherePK().Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if count, err := r.RowsAffected(); err != nil || count == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

func (pg *Pg) NamespaceDeleteMembership(ctx context.Context, tenantID string, member *models.Member) error {
	db := pg.GetConnection(ctx)

	e := entity.MembershipFromModel(tenantID, member)
	r, err := db.NewDelete().Model(e).WherePK().Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if count, err := r.RowsAffected(); err != nil || count == 0 {
		return store.ErrNoDocuments
	}

	// Clear the removed member's preferred namespace if it points at this one. Targeted Set, not a
	// full-model update, since preferred_namespace_id is skipupdate.
	if _, err := db.NewUpdate().
		Model((*entity.User)(nil)).
		Set("preferred_namespace_id = NULL").
		Where("id = ? AND preferred_namespace_id = ?", member.ID, tenantID).
		Exec(ctx); err != nil {
		return fromSQLError(err)
	}

	return nil
}
