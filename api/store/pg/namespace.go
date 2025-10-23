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
	db := pg.getExecutor(ctx)

	if namespace.TenantID == "" {
		namespace.TenantID = uuid.Generate()
	}

	namespace.CreatedAt = clock.Now()

	if _, err := db.NewInsert().Model(entity.NamespaceFromModel(namespace)).Exec(ctx); err != nil {
		return "", fromSqlError(err)
	}

	return namespace.TenantID, nil
}

func (pg *Pg) NamespaceConflicts(ctx context.Context, target *models.NamespaceConflicts) ([]string, bool, error) {
	db := pg.getExecutor(ctx)

	namespaces := make([]map[string]any, 0)
	if err := db.NewSelect().Model((*entity.Namespace)(nil)).Column("name").Where("name = ?", target.Name).Scan(ctx, &namespaces); err != nil {
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
	db := pg.getExecutor(ctx)

	entities := make([]entity.Namespace, 0)
	query := db.NewSelect().Model(&entities).Relation("Memberships.User")
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

func (pg *Pg) NamespaceResolve(ctx context.Context, resolver store.NamespaceResolver, val string) (*models.Namespace, error) {
	db := pg.getExecutor(ctx)

	column, err := NamespaceResolverToString(resolver)
	if err != nil {
		return nil, err
	}

	ns := new(entity.Namespace)
	query := db.NewSelect().Model(ns).Relation("Memberships.User").Where("? = ?", bun.Ident(column), val)
	if err := query.Scan(ctx); err != nil {
		return nil, fromSqlError(err)
	}

	return entity.NamespaceToModel(ns), nil
}

func (pg *Pg) NamespaceGetPreferred(ctx context.Context, userID string) (*models.Namespace, error) {
	db := pg.getExecutor(ctx)

	ns := new(entity.Namespace)
	if err := db.NewSelect().
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
	db := pg.getExecutor(ctx)

	n := entity.NamespaceFromModel(namespace)
	n.UpdatedAt = clock.Now()

	_, err := db.NewUpdate().Model(n).WherePK().Exec(ctx)

	return fromSqlError(err)
}

func (pg *Pg) NamespaceIncrementDeviceCount(ctx context.Context, tenantID string, status models.DeviceStatus, count int64) error {
	db := pg.getExecutor(ctx)

	column := "devices" + string(status) + "count"
	result, err := db.NewUpdate().
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
	deletedCount, err := pg.NamespaceDeleteMany(ctx, []string{namespace.TenantID})
	switch {
	case err != nil:
		return err
	case deletedCount < 1:
		return store.ErrNoDocuments
	default:
		return nil
	}
}

func (pg *Pg) NamespaceDeleteMany(ctx context.Context, tenantIDs []string) (int64, error) {
	db := pg.getExecutor(ctx)
	fn := pg.namespaceDeleteManyFn(ctx, tenantIDs)

	if tx, ok := db.(bun.Tx); ok {
		return fn(tx)
	} else { // nolint:revive
		tx, err := pg.driver.BeginTx(ctx, nil)
		if err != nil {
			return 0, fromSqlError(err)
		}

		defer func() {
			if p := recover(); p != nil {
				_ = tx.Rollback()
				panic(p)
			}
		}()

		count, err := fn(tx)
		if err != nil {
			_ = tx.Rollback()
			return 0, err
		}

		if err := tx.Commit(); err != nil {
			return 0, fromSqlError(err)
		}

		return count, nil

	}
}

func (pg *Pg) namespaceDeleteManyFn(ctx context.Context, tenantIDs []string) func(tx bun.Tx) (int64, error) {
	return func(tx bun.Tx) (int64, error) {
		res, err := tx.NewDelete().Model((*entity.Namespace)(nil)).Where("id IN (?)", bun.In(tenantIDs)).Exec(ctx)
		if err != nil {
			return 0, fromSqlError(err)
		}

		count, _ := res.RowsAffected()

		entities := []any{
			(*entity.Device)(nil),
			// (*entity.Session)(nil),
			// (*entity.FirewallRule)(nil),
			(*entity.PublicKey)(nil),
			// (*entity.RecordedSession)(nil),
			(*entity.APIKey)(nil),
		}

		for _, e := range entities {
			if _, err := tx.NewDelete().Model(e).Where("namespace_id IN (?)", bun.In(tenantIDs)).Exec(ctx); err != nil {
				return 0, fromSqlError(err)
			}
		}

		if _, err := tx.NewUpdate().
			Model((*entity.User)(nil)).
			Set("preferred_namespace_id = NULL").
			Where("preferred_namespace_id IN (?)", bun.In(tenantIDs)).
			Exec(ctx); err != nil {
			return 0, fromSqlError(err)
		}

		return count, nil
	}
}

func NamespaceResolverToString(resolver store.NamespaceResolver) (string, error) {
	switch resolver {
	case store.NamespaceTenantIDResolver:
		return "id", nil
	case store.NamespaceNameResolver:
		return "name", nil
	default:
		return "", store.ErrResolverNotFound
	}
}
