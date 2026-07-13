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
	db := pg.GetConnection(ctx)

	user.CreatedAt = clock.Now()
	if user.ID == "" {
		user.ID = uuid.Generate()
	}

	if _, err := db.NewInsert().Model(entity.UserFromModel(user)).Exec(ctx); err != nil {
		return "", fromSQLError(err)
	}

	return user.ID, nil
}

func (pg *Pg) UserList(ctx context.Context, opts ...store.QueryOption) ([]models.User, int, error) {
	db := pg.GetConnection(ctx)

	entities := make([]entity.User, 0)
	query := UserSelectQuery(db.NewSelect().Model(&entities))

	var err error
	query, err = applyOptions(ctx, query, opts...)
	if err != nil {
		return nil, 0, err
	}

	count, err := query.ScanAndCount(ctx)
	if err != nil {
		return nil, 0, fromSQLError(err)
	}

	users := make([]models.User, len(entities))
	for i, e := range entities {
		users[i] = *entity.UserToModel(&e)
	}

	return users, count, nil
}

func (pg *Pg) UserResolve(ctx context.Context, resolver store.UserResolver, val string, opts ...store.QueryOption) (*models.User, error) {
	db := pg.GetConnection(ctx)

	column, err := UserResolverToString(resolver)
	if err != nil {
		return nil, err
	}

	u := new(entity.User)
	query := UserSelectQuery(db.NewSelect().Model(u)).
		Where("? = ?", bun.Ident(column), val)

	query, err = applyOptions(ctx, query, opts...)
	if err != nil {
		return nil, err
	}

	if err = query.Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.UserToModel(u), nil
}

func (pg *Pg) UserGetInfo(ctx context.Context, userID string) (userInfo *models.UserInfo, err error) {
	db := pg.GetConnection(ctx)

	var namespaceEntities []entity.Namespace
	err = db.NewSelect().
		Model(&namespaceEntities).
		Relation("Memberships.User").
		Where("owner_id = ? OR EXISTS (SELECT 1 FROM memberships WHERE memberships.namespace_id = namespace.id AND memberships.user_id = ?)", userID, userID).
		Scan(ctx)
	if err != nil {
		return nil, fromSQLError(err)
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
	db := pg.GetConnection(ctx)

	u := entity.UserFromModel(user)
	u.UpdatedAt = clock.Now()

	r, err := db.NewUpdate().Model(u).WherePK().Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return fromSQLError(err)
}

func (pg *Pg) UserUpdatePreferredNamespace(ctx context.Context, userID, tenantID string) error {
	db := pg.GetConnection(ctx)

	q := db.NewUpdate().
		Model((*entity.User)(nil)).
		Where("id = ?", userID)
	if tenantID == "" {
		q = q.Set("preferred_namespace_id = NULL")
	} else {
		q = q.Set("preferred_namespace_id = ?", tenantID)
	}

	r, err := q.Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

func (pg *Pg) UserDelete(ctx context.Context, user *models.User) error {
	db := pg.GetConnection(ctx)

	u := entity.UserFromModel(user)

	r, err := db.NewDelete().Model(u).WherePK().Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return fromSQLError(err)
}

// UserSelectQuery applies the standard user SELECT decorations: all columns
// plus the computed namespaces count. The caller provides the base query
// with the desired model (core or cloud entity).
func UserSelectQuery(q *bun.SelectQuery) *bun.SelectQuery {
	return q.
		ColumnExpr(`"user".*`).
		ColumnExpr(`(SELECT COUNT(*) FROM namespaces WHERE owner_id = "user".id) AS namespaces`)
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
