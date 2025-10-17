package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/uptrace/bun"
)

func (pg *Pg) UserCreate(ctx context.Context, user *models.User) (string, error) {
	db := pg.getExecutor(ctx)

	user.ID = uuid.Generate()
	user.CreatedAt = clock.Now()

	if _, err := db.NewInsert().Model(entity.UserFromModel(user)).Exec(ctx); err != nil {
		return "", err
	}

	return user.ID, nil
}

func (pg *Pg) UserCreateInvited(ctx context.Context, email string) (string, error) {
	return "", nil
}

func (pg *Pg) UserConflicts(ctx context.Context, target *models.UserConflicts) ([]string, bool, error) {
	db := pg.getExecutor(ctx)

	users := make([]map[string]any, 0)
	if err := db.NewSelect().Model((*entity.User)(nil)).Column("email").Where("email = ?", target.Email).Scan(ctx, &users); err != nil {
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

func (pg *Pg) UserList(ctx context.Context, opts ...store.QueryOption) ([]models.User, int, error) {
	db := pg.getExecutor(ctx)

	entities := make([]entity.User, 0)
	query := db.NewSelect().Model(&entities)
	if err := applyOptions(ctx, query, opts...); err != nil {
		return nil, 0, fromSqlError(err)
	}

	count, err := query.ScanAndCount(ctx)
	if err != nil {
		return nil, 0, fromSqlError(err)
	}

	users := make([]models.User, len(entities))
	for i, e := range entities {
		users[i] = *entity.UserToModel(&e)
	}

	return users, count, nil
}

func (pg *Pg) UserResolve(ctx context.Context, resolver store.UserResolver, val string, opts ...store.QueryOption) (*models.User, error) {
	db := pg.getExecutor(ctx)

	column, err := UserResolverToString(resolver)
	if err != nil {
		return nil, err
	}

	u := new(entity.User)
	if err := db.NewSelect().Model(u).Where("? = ?", bun.Ident(column), val).Scan(ctx); err != nil {
		return nil, fromSqlError(err)
	}

	return entity.UserToModel(u), nil
}

func (pg *Pg) UserGetInfo(ctx context.Context, userID string) (userInfo *models.UserInfo, err error) {
	db := pg.getExecutor(ctx)

	var namespaceEntities []entity.Namespace
	err = db.NewSelect().
		Model(&namespaceEntities).
		Relation("Memberships.User").
		Where("owner_id = ? OR EXISTS (SELECT 1 FROM memberships WHERE memberships.namespace_id = namespace.id AND memberships.user_id = ?)", userID, userID).
		Scan(ctx)
	if err != nil {
		return nil, fromSqlError(err)
	}

	userInfo = &models.UserInfo{
		OwnedNamespaces:      make([]models.Namespace, 0),
		AssociatedNamespaces: make([]models.Namespace, 0),
	}

	for _, nsEntity := range namespaceEntities {
		ns := entity.NamespaceToModel(&nsEntity)

		if nsEntity.OwnerID == userID {
			userInfo.OwnedNamespaces = append(userInfo.OwnedNamespaces, *ns)
		} else {
			userInfo.AssociatedNamespaces = append(userInfo.AssociatedNamespaces, *ns)
		}
	}

	return userInfo, nil
}

func (pg *Pg) UserUpdate(ctx context.Context, user *models.User) error {
	db := pg.getExecutor(ctx)

	u := entity.UserFromModel(user)
	u.UpdatedAt = clock.Now()

	r, err := db.NewUpdate().Model(u).WherePK().Exec(ctx)
	if err != nil {
		return fromSqlError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return fromSqlError(err)
}

func (pg *Pg) UserDelete(ctx context.Context, user *models.User) error {
	db := pg.getExecutor(ctx)

	u := entity.UserFromModel(user)

	r, err := db.NewDelete().Model(u).WherePK().Exec(ctx)
	if err != nil {
		return fromSqlError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return fromSqlError(err)
}

func UserResolverToString(resolver store.UserResolver) (string, error) {
	switch resolver {
	case store.UserIDResolver:
		return "id", nil
	case store.UserEmailResolver:
		return "email", nil
	case store.UserUsernameResolver:
		return "username", nil
	default:
		return "", store.ErrResolverNotFound
	}
}
