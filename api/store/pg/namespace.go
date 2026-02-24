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
	db := pg.GetConnection(ctx)

	namespace.CreatedAt = clock.Now()
	if namespace.TenantID == "" {
		namespace.TenantID = uuid.Generate()
	}

	nsEntity := entity.NamespaceFromModel(namespace)
	if _, err := db.NewInsert().Model(nsEntity).Exec(ctx); err != nil {
		return "", fromSQLError(err)
	}

	if len(nsEntity.Memberships) > 0 {
		if _, err := db.NewInsert().Model(&nsEntity.Memberships).Exec(ctx); err != nil {
			return "", fromSQLError(err)
		}
	}

	return namespace.TenantID, nil
}

func (pg *Pg) NamespaceConflicts(ctx context.Context, target *models.NamespaceConflicts) ([]string, bool, error) {
	db := pg.GetConnection(ctx)

	if target.Name == "" {
		return []string{}, false, nil
	}

	namespaces := make([]entity.Namespace, 0)
	query := db.NewSelect().
		Model(&namespaces).
		Column("name").
		WhereGroup(" OR ", func(q *bun.SelectQuery) *bun.SelectQuery {
			if target.Name != "" {
				q = q.Where("name = ?", target.Name)
			}

			return q
		})

	if err := query.Scan(ctx); err != nil {
		return nil, false, fromSQLError(err)
	}

	seen := make(map[string]bool)
	for _, ns := range namespaces {
		if target.Name != "" && ns.Name == target.Name {
			seen["name"] = true
		}
	}

	conflicts := make([]string, 0, len(seen))
	for field := range seen {
		conflicts = append(conflicts, field)
	}

	return conflicts, len(conflicts) > 0, nil
}

func (pg *Pg) NamespaceList(ctx context.Context, opts ...store.QueryOption) ([]models.Namespace, int, error) {
	db := pg.GetConnection(ctx)

	entities := make([]entity.Namespace, 0)
	query := db.NewSelect().Model(&entities)

	var err error
	query, err = applyOptions(ctx, query, opts...)
	if err != nil {
		return nil, 0, err
	}

	count, err := query.ScanAndCount(ctx)
	if err != nil {
		return nil, 0, fromSQLError(err)
	}

	namespaces := make([]models.Namespace, len(entities))
	for i, e := range entities {
		namespaces[i] = *entity.NamespaceToModel(&e)
	}

	return namespaces, count, nil
}

func (pg *Pg) NamespaceResolve(ctx context.Context, resolver store.NamespaceResolver, val string) (*models.Namespace, error) {
	db := pg.GetConnection(ctx)

	column, err := NamespaceResolverToString(resolver)
	if err != nil {
		return nil, err
	}

	ns := new(entity.Namespace)
	query := db.NewSelect().Model(ns).Relation("Memberships.User").Where("? = ?", bun.Ident(column), val)
	if err := query.Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.NamespaceToModel(ns), nil
}

func (pg *Pg) NamespaceGetPreferred(ctx context.Context, userID string) (*models.Namespace, error) {
	db := pg.GetConnection(ctx)

	ns := new(entity.Namespace)
	if err := db.NewSelect().
		Model(ns).
		Relation("Memberships.User").
		Join("JOIN users").
		JoinOn("namespace.id = users.preferred_namespace_id OR namespace.id IN (SELECT namespace_id FROM memberships WHERE user_id = users.id)").
		Where("users.id = ?", userID).
		OrderExpr(namespaceExprPreferredOrder()).
		Limit(1).
		Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.NamespaceToModel(ns), nil
}

func (pg *Pg) NamespaceUpdate(ctx context.Context, namespace *models.Namespace) error {
	db := pg.GetConnection(ctx)

	// First check if namespace exists
	exists, err := db.NewSelect().Model((*entity.Namespace)(nil)).Where("id = ?", namespace.TenantID).Exists(ctx)
	if err != nil {
		return fromSQLError(err)
	}
	if !exists {
		return store.ErrNoDocuments
	}

	n := entity.NamespaceFromModel(namespace)
	n.UpdatedAt = clock.Now()

	r, err := db.NewUpdate().Model(n).WherePK().Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

func (pg *Pg) NamespaceIncrementDeviceCount(ctx context.Context, tenantID string, status models.DeviceStatus, count int64) error {
	db := pg.GetConnection(ctx)

	column := "devices_" + string(status) + "_count"
	result, err := db.NewUpdate().
		Model((*entity.Namespace)(nil)).
		Set("? = ? + ?", bun.Ident(column), bun.Ident(column), count).
		Where("id = ?", tenantID).
		Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

func (pg *Pg) NamespaceSyncDeviceCounts(ctx context.Context) error {
	db := pg.GetConnection(ctx)

	_, err := db.NewRaw(`
		UPDATE namespaces SET
			devices_accepted_count = COALESCE(c.accepted, 0),
			devices_pending_count  = COALESCE(c.pending, 0),
			devices_rejected_count = COALESCE(c.rejected, 0),
			devices_removed_count  = COALESCE(c.removed, 0)
		FROM (
			SELECT
				namespace_id,
				COUNT(*) FILTER (WHERE status = 'accepted') AS accepted,
				COUNT(*) FILTER (WHERE status = 'pending')  AS pending,
				COUNT(*) FILTER (WHERE status = 'rejected') AS rejected,
				COUNT(*) FILTER (WHERE status = 'removed')  AS removed
			FROM devices
			GROUP BY namespace_id
		) c
		WHERE namespaces.id = c.namespace_id
	`).Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	_, err = db.NewUpdate().
		Model((*entity.Namespace)(nil)).
		Set("devices_accepted_count = 0").
		Set("devices_pending_count = 0").
		Set("devices_rejected_count = 0").
		Set("devices_removed_count = 0").
		Where("id NOT IN (SELECT DISTINCT namespace_id FROM devices)").
		Exec(ctx)
	if err != nil {
		return fromSQLError(err)
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
	db := pg.GetConnection(ctx)
	fn := pg.namespaceDeleteManyFn(ctx, tenantIDs)

	if tx, ok := db.(bun.Tx); ok {
		return fn(tx)
	} else { // nolint:revive
		tx, err := pg.driver.BeginTx(ctx, nil)
		if err != nil {
			return 0, fromSQLError(err)
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
			return 0, fromSQLError(err)
		}

		return count, nil
	}
}

func (pg *Pg) namespaceDeleteManyFn(ctx context.Context, tenantIDs []string) func(tx bun.Tx) (int64, error) {
	return func(tx bun.Tx) (int64, error) {
		if _, err := tx.NewDelete().
			Model((*entity.Session)(nil)).
			Where("device_id IN (SELECT id FROM devices WHERE namespace_id IN (?))", bun.List(tenantIDs)).
			Exec(ctx); err != nil {
			return 0, fromSQLError(err)
		}

		res, err := tx.NewDelete().Model((*entity.Namespace)(nil)).Where("id IN (?)", bun.List(tenantIDs)).Exec(ctx)
		if err != nil {
			return 0, fromSQLError(err)
		}

		count, _ := res.RowsAffected()

		entities := []any{
			(*entity.Device)(nil),
			(*entity.PublicKey)(nil),
			(*entity.APIKey)(nil),
			(*entity.Tunnel)(nil),
		}

		for _, e := range entities {
			if _, err := tx.NewDelete().Model(e).Where("namespace_id IN (?)", bun.List(tenantIDs)).Exec(ctx); err != nil {
				return 0, fromSQLError(err)
			}
		}

		if _, err := tx.NewUpdate().
			Model((*entity.User)(nil)).
			Set("preferred_namespace_id = NULL").
			Where("preferred_namespace_id IN (?)", bun.List(tenantIDs)).
			Exec(ctx); err != nil {
			return 0, fromSQLError(err)
		}

		return count, nil
	}
}

// namespaceExprPreferredOrder returns the SQL expression for ordering by preferred namespace.
func namespaceExprPreferredOrder() string {
	return "CASE WHEN namespace.id = users.preferred_namespace_id THEN 0 ELSE 1 END"
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
