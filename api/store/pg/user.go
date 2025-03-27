package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/pg/internal/entity"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/uptrace/bun"
)

func (pg *pg) UserCreate(ctx context.Context, user *models.User) (string, error) {
	user.ID = uuid.Generate()
	user.CreatedAt = clock.Now()
	user.UpdatedAt = clock.Now()

	if _, err := pg.driver.NewInsert().Model(&entity.User{User: *user}).Exec(ctx); err != nil {
		return "", err
	}

	return user.ID, nil
}

func (pg *pg) UserCreateInvited(ctx context.Context, email string) (string, error) {
	// TODO: unify create methods
	return "", nil
}

func (pg *pg) UserConflicts(ctx context.Context, target *models.UserConflicts) ([]string, bool, error) {
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

func (pg *pg) UserList(ctx context.Context, paginator query.Paginator, filters query.Filters) ([]models.User, int, error) {
	return nil, 0, nil
}

func (pg *pg) UserGet(ctx context.Context, ident store.UserIdent, val string) (*models.User, error) {
	u := new(entity.User)
	if err := pg.driver.NewSelect().Model(u).Relation("Memberships").Where("? = ?", bun.Ident(ident), val).Scan(ctx); err != nil {
		return nil, fromSqlError(err)
	}

	return &u.User, nil
}

func (pg *pg) UserPreferredNamespace(ctx context.Context, ident store.UserIdent, val string) (*models.Namespace, error) {
	ns := new(entity.Namespace)
	if err := pg.driver.NewSelect().
		Model(ns).
		Relation("Memberships").
		Join("JOIN users AS u ON namespace.id = u.preferred_namespace_id OR namespace.id IN (SELECT namespace_id FROM memberships WHERE user_id = u.id)").
		Where("u.? = ?", bun.Ident(ident), val).
		OrderExpr("CASE WHEN namespace.id = u.preferred_namespace_id THEN 0 ELSE 1 END").
		Limit(1).
		Scan(ctx); err != nil {
		return nil, fromSqlError(err)
	}

	return &ns.Namespace, nil
}

func (pg *pg) UserGetInfo(ctx context.Context, id string) (userInfo *models.UserInfo, err error) {
	// TODO: unify get methods
	return nil, nil
}
