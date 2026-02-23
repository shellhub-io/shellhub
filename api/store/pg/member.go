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

	// Check if user has this namespace as preferred and clear it if so
	user := new(entity.User)
	err = db.NewSelect().Model(user).Where("id = ? AND preferred_namespace_id = ?", member.ID, tenantID).Limit(1).Scan(ctx)

	// If user was found (no error), clear the preferred namespace
	if err == nil && user.ID != "" {
		user.Preferences.PreferredNamespace = ""
		if _, err := db.NewUpdate().Model(user).Column("preferred_namespace_id").WherePK().Exec(ctx); err != nil {
			return fromSQLError(err)
		}
	}
	// If user not found (err != nil), that's OK - just means they don't have this as preferred

	return nil
}
