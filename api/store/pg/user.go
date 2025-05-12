package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/uptrace/bun"
)

func (pg *Pg) UserCreate(ctx context.Context, user *models.User) (string, error) {
	user.ID = uuid.Generate()
	user.CreatedAt = clock.Now()
	user.UpdatedAt = clock.Now()

	if _, err := pg.driver.NewInsert().Model(entity.UserFromModel(user)).Exec(ctx); err != nil {
		return "", err
	}

	return user.ID, nil
}

func (pg *Pg) UserConflicts(ctx context.Context, target *models.UserConflicts) ([]string, bool, error) {
	users := make([]map[string]any, 0)
	if err := pg.driver.NewSelect().Model((*entity.User)(nil)).Column("email").Where("email = ?", target.Email).Scan(ctx, &users); err != nil {
		return nil, false, err
	}

	conflicts := make([]string, 0)
	for _, user := range users {
		if user["email"] == target.Email {
			conflicts = append(conflicts, "email")
		}
	}

	return conflicts, len(conflicts) > 0, nil
}

func (pg *Pg) UserList(ctx context.Context, paginator query.Paginator, filters query.Filters) ([]models.User, int, error) {
	return nil, 0, nil
}

func (pg *Pg) UserGet(ctx context.Context, ident store.UserIdent, val string) (*models.User, error) {
	u := new(entity.User)
	if err := pg.driver.NewSelect().Model(u).Where("? = ?", bun.Ident(ident), val).Scan(ctx); err != nil {
		return nil, fromSqlError(err)
	}

	return entity.UserToModel(u), nil
}

func (pg *Pg) UserGetInfo(ctx context.Context, id string) (userInfo *models.UserInfo, err error) {
	// TODO: unify get methods
	return nil, nil
}

func (pg *Pg) UserPreferredNamespace(ctx context.Context, ident store.UserIdent, val string) (*models.Namespace, error) {
	ns := new(entity.Namespace)
	if err := pg.driver.NewSelect().
		Model(ns).
		Relation("Memberships.User").
		Join("JOIN users").
		JoinOn("namespace.id = users.preferred_namespace_id OR namespace.id IN (SELECT namespace_id FROM memberships WHERE user_id = users.id)"). // TODO: subquery
		Where("users.? = ?", bun.Ident(ident), val).
		OrderExpr("CASE WHEN namespace.id = users.preferred_namespace_id THEN 0 ELSE 1 END"). // TODO: segunda ordenacao pela membership mais recente
		Limit(1).
		Scan(ctx); err != nil {
		return nil, fromSqlError(err)
	}

	return entity.NamespaceToModel(ns), nil
}

func (pg *Pg) UserSave(ctx context.Context, user *models.User) error {
	u := entity.UserFromModel(user)
	u.UpdatedAt = clock.Now()

	r, err := pg.driver.NewUpdate().Model(u).WherePK().Exec(ctx)
	if err != nil {
		return fromSqlError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return fromSqlError(err)
}

func (pg *Pg) UserDelete(ctx context.Context, user *models.User) error {
	u := entity.UserFromModel(user)

	r, err := pg.driver.NewDelete().Model(u).WherePK().Exec(ctx)
	if err != nil {
		return fromSqlError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return fromSqlError(err)
}
