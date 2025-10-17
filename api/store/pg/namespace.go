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

func (pg *Pg) NamespaceCreate(ctx context.Context, namespace *models.Namespace) (string, error) {
	if namespace.TenantID == "" {
		namespace.TenantID = uuid.Generate()
	}

	namespace.CreatedAt = clock.Now()

	if _, err := pg.driver.NewInsert().Model(entity.NamespaceFromModel(namespace)).Exec(ctx); err != nil {
		return "", fromSqlError(err)
	}

	return namespace.TenantID, nil
}

func (pg *Pg) NamespaceConflicts(ctx context.Context, target *models.NamespaceConflicts) ([]string, bool, error) {
	namespaces := make([]map[string]any, 0)
	if err := pg.driver.NewSelect().Model((*entity.Namespace)(nil)).Column("name").Where("name = ?", target.Name).Scan(ctx, &namespaces); err != nil {
		return nil, false, fromSqlError(err)
	}

	conflicts := make([]string, 0)
	for _, user := range namespaces {
		if user["name"] == target.Name {
			conflicts = append(conflicts, "name")
		}
	}

	return conflicts, len(conflicts) > 0, nil
}

func (pg *Pg) NamespaceList(ctx context.Context, opts ...store.QueryOption) ([]models.Namespace, int, error) {
	entities := make([]entity.Namespace, 0)
	query := pg.driver.NewSelect().Model(&entities).Relation("Memberships.User")
	if err := applyOptions(ctx, query, opts...); err != nil {
		return nil, 0, fromSqlError(err)
	}

	count, err := query.ScanAndCount(ctx)
	if err != nil {
		return nil, 0, fromSqlError(err)
	}

	namespaces := make([]models.Namespace, len(entities))
	for i, e := range entities {
		namespaces[i] = *entity.NamespaceToModel(&e)
	}

	return namespaces, count, nil
}

func (pg *Pg) NamespaceResolve(ctx context.Context, resolver store.NamespaceResolver, val string, opts ...store.QueryOption) (*models.Namespace, error) {
	ns := new(entity.Namespace)

	query := pg.driver.NewSelect().Model(ns).Relation("Memberships.User").Where("? = ?", bun.Ident(string(resolver)), val)
	if err := applyOptions(ctx, query, opts...); err != nil {
		return nil, fromSqlError(err)
	}

	if err := query.Scan(ctx); err != nil {
		return nil, fromSqlError(err)
	}

	return entity.NamespaceToModel(ns), nil
}

func (pg *Pg) NamespaceGetPreferred(ctx context.Context, userID string) (*models.Namespace, error) {
	ns := new(entity.Namespace)
	if err := pg.driver.NewSelect().
		Model(ns).
		Relation("Memberships.User").
		Join("JOIN users").
		JoinOn("namespace.id = users.preferred_namespace_id OR namespace.id IN (SELECT namespace_id FROM memberships WHERE user_id = users.id)").
		Where("users.id = ?", userID).
		OrderExpr("CASE WHEN namespace.id = users.preferred_namespace_id THEN 0 ELSE 1 END").
		Limit(1).
		Scan(ctx); err != nil {
		return nil, fromSqlError(err)
	}

	return entity.NamespaceToModel(ns), nil
}

func (pg *Pg) NamespaceUpdate(ctx context.Context, namespace *models.Namespace) error {
	n := entity.NamespaceFromModel(namespace)
	n.UpdatedAt = clock.Now()

	_, err := pg.driver.NewUpdate().Model(n).WherePK().Exec(ctx)

	return fromSqlError(err)
}

func (pg *Pg) NamespaceIncrementDeviceCount(ctx context.Context, tenantID string, status models.DeviceStatus, count int64) error {
	column := "devices" + string(status) + "count"
	result, err := pg.driver.NewUpdate().
		Model((*entity.Namespace)(nil)).
		Set("? = ? + ?", bun.Ident(column), bun.Ident(column), count).
		Where("id = ?", tenantID).
		Exec(ctx)
	if err != nil {
		return fromSqlError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fromSqlError(err)
	}

	if rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

func (pg *Pg) NamespaceDelete(ctx context.Context, namespace *models.Namespace) error {
	n := entity.NamespaceFromModel(namespace)
	_, err := pg.driver.NewDelete().Model(n).WherePK().Exec(ctx)

	return fromSqlError(err)
}
