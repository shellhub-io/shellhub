package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (pg *Pg) NamespaceCreateMembership(ctx context.Context, tenantID string, membership *models.Member) error {
	db := pg.getConnection(ctx)

	membership.AddedAt = clock.Now()
	entity := entity.MembershipFromModel(tenantID, membership)
	if _, err := db.NewInsert().Model(entity).Exec(ctx); err != nil {
		return fromSQLError(err)
	}

	return nil
}

func (pg *Pg) NamespaceUpdateMembership(ctx context.Context, tenantID string, member *models.Member) error {
	db := pg.getConnection(ctx)

	e := entity.MembershipFromModel(tenantID, member)
	e.UpdatedAt = clock.Now()
	_, err := db.NewUpdate().Model(e).WherePK().Exec(ctx)

	return fromSQLError(err)
}

func (pg *Pg) NamespaceDeleteMembership(ctx context.Context, tenantID string, member *models.Member) error {
	db := pg.getConnection(ctx)

	e := entity.MembershipFromModel(tenantID, member)
	r, err := db.NewDelete().Model(e).WherePK().Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if count, err := r.RowsAffected(); err != nil || count == 0 {
		return store.ErrNoDocuments
	}

	user := new(entity.User)
	if err := db.NewSelect().Model(user).Where("id = ? AND preferred_namespace_id = ?", member.ID, tenantID).Limit(1).Scan(ctx); err != nil {
		return fromSQLError(err)
	}

	if user != nil && user.ID != "" {
		user.Preferences.PreferredNamespace = ""
		if _, err := db.NewUpdate().Model(user).Column("preferred_namespace_id").WherePK().Exec(ctx); err != nil {
			return fromSQLError(err)
		}
	}

	return nil
}
