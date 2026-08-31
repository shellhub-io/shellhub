package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/pg/entity"
)

func (pg *Pg) NamespaceCreateMembership(ctx context.Context, sc scope.Scope, membership *models.Member) error {
	db := pg.GetConnection(ctx)

	tenantID, err := requireBounded(sc)
	if err != nil {
		return err
	}

	membership.AddedAt = clock.Now()
	entity := entity.MembershipFromModel(tenantID, membership)
	if _, err := db.NewInsert().Model(entity).Exec(ctx); err != nil {
		return fromSQLError(err)
	}

	return nil
}

func (pg *Pg) NamespaceUpdateMembership(ctx context.Context, sc scope.Scope, member *models.Member) error {
	db := pg.GetConnection(ctx)

	tenantID, err := requireBounded(sc)
	if err != nil {
		return err
	}

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

func (pg *Pg) NamespaceDeleteMembership(ctx context.Context, sc scope.Scope, member *models.Member) error {
	db := pg.GetConnection(ctx)

	tenantID, err := requireBounded(sc)
	if err != nil {
		return err
	}

	e := entity.MembershipFromModel(tenantID, member)
	r, err := db.NewDelete().Model(e).WherePK().Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if count, err := r.RowsAffected(); err != nil || count == 0 {
		return store.ErrNoDocuments
	}

	if _, err := db.NewUpdate().
		Model((*entity.User)(nil)).
		Set("preferred_namespace_id = NULL").
		Where("id = ? AND preferred_namespace_id = ?", member.ID, tenantID).
		Exec(ctx); err != nil {
		return fromSQLError(err)
	}

	return nil
}
