package pg

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/uptrace/bun"
)

func (pg *Pg) NamespaceCreate(ctx context.Context, namespace *models.Namespace) (string, error) {
	namespace.CreatedAt = clock.Now()

	// Single-namespace binding: once the instance is bound to a namespace
	// (systems.instance_tenant_id, set in Community deployments), refuse any further namespace
	// with a specific error — distinct from a duplicate-name conflict. Enterprise/Cloud keep the
	// binding empty (their store wrapper strips it), so this never triggers there. Setup creates
	// the first namespace while the binding is still empty.
	system, err := pg.SystemGet(ctx)
	if err != nil {
		return "", err
	}

	if system.InstanceTenantID != "" {
		return "", store.ErrNamespaceSingle
	}

	if namespace.TenantID == "" {
		namespace.TenantID = uuid.Generate()
	}

	// Identity-first: a namespace created without an explicit SSH access mode is
	// born identity, with the owner starter policy seeded below.
	if namespace.Settings == nil {
		namespace.Settings = &models.NamespaceSettings{}
	}

	if namespace.Settings.SSHAccessMode == "" {
		namespace.Settings.SSHAccessMode = models.SSHAccessModeIdentity
	}

	// Insert the namespace, its memberships, and its legacy install key atomically, so a failure can't
	// leave a namespace without the legacy key that keyless enrollments attribute to. InstallKeyCreate
	// resolves its connection from ctx, so it joins this transaction transparently.
	if err := pg.WithTransaction(ctx, func(ctx context.Context) error {
		db := pg.GetConnection(ctx)

		nsEntity := entity.NamespaceFromModel(namespace)
		if _, err := db.NewInsert().Model(nsEntity).Exec(ctx); err != nil {
			return fromSQLError(err)
		}

		if len(nsEntity.Memberships) > 0 {
			if _, err := db.NewInsert().Model(&nsEntity.Memberships).Exec(ctx); err != nil {
				return fromSQLError(err)
			}
		}

		// Every namespace gets its two system-managed install keys, so every keyless enrollment has a
		// source to attribute to. Created here, at the single namespace-creation chokepoint every path
		// funnels through — the API, setup, the CLI, and the cloud/enterprise store that delegates here —
		// so no path can skip them. Digests are derived from the tenant; agents never present them, they
		// are resolved by type.
		//
		// legacy: tenant-only keyless enrollment (a device presenting only the tenant ID). Manual mode,
		// so such devices land pending.
		legacyDigest := sha256.Sum256([]byte("system:" + namespace.TenantID))
		if _, err := pg.InstallKeyCreate(ctx, &models.InstallKey{
			ID:        hex.EncodeToString(legacyDigest[:]),
			Name:      string(models.InstallKeyTypeLegacy),
			TenantID:  namespace.TenantID,
			Mode:      models.InstallKeyModeManual,
			Reusable:  true,
			Type:      models.InstallKeyTypeLegacy,
			CreatedBy: namespace.Owner,
		}); err != nil {
			return err
		}

		// pairing: code-pairing enrollment (a tenant-less agent accepted via its printed code). Automatic
		// mode, since acceptance is the code itself; the pairing flow accepts the device explicitly.
		pairingDigest := sha256.Sum256([]byte("system:pairing:" + namespace.TenantID))
		if _, err := pg.InstallKeyCreate(ctx, &models.InstallKey{
			ID:        hex.EncodeToString(pairingDigest[:]),
			Name:      string(models.InstallKeyTypePairing),
			TenantID:  namespace.TenantID,
			Mode:      models.InstallKeyModeAutomatic,
			Reusable:  true,
			Type:      models.InstallKeyTypePairing,
			CreatedBy: namespace.Owner,
		}); err != nil {
			return err
		}

		// An identity-mode namespace with zero policies denies every SSH login
		// (default-deny), so a namespace born identity gets the owner starter
		// policy atomically with its creation (see NewOwnerAccessPolicy).
		// AccessPolicyCreate joins this transaction via ctx.
		if namespace.Settings.IsIdentityAccess() && namespace.Owner != "" {
			if _, err := pg.AccessPolicyCreate(ctx, models.NewOwnerAccessPolicy(namespace.TenantID, namespace.Owner)); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return "", err
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

	// namespaces.id is a uuid-typed column; a malformed value would otherwise reach
	// Postgres and fail with SQLSTATE 22P02. Treat it as not-found to match the prior
	// Mongo behavior and avoid logging a misleading SQL error (see #6404).
	if resolver == store.NamespaceTenantIDResolver {
		if _, err := uuid.Parse(val); err != nil {
			return nil, store.ErrNoDocuments
		}
	}

	ns := new(entity.Namespace)
	query := db.NewSelect().Model(ns).Relation("Memberships.User").Where("? = ?", bun.Ident(column), val)
	if err := query.Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.NamespaceToModel(ns), nil
}

func (pg *Pg) NamespaceGetMembers(ctx context.Context, tenantID string, opts ...store.QueryOption) ([]models.MemberView, int, error) {
	db := pg.GetConnection(ctx)

	entities := make([]entity.Membership, 0)
	query := db.NewSelect().
		Model(&entities).
		Relation("User").
		Where("membership.namespace_id = ?", tenantID).
		// Service accounts are not human members; keep them out of the members list.
		Where("membership.user_id IN (SELECT id FROM users WHERE type != ?)", string(models.UserTypeService)).
		OrderExpr("membership.created_at ASC")

	var err error
	query, err = applyOptions(ctx, query, opts...)
	if err != nil {
		return nil, 0, err
	}

	count, err := query.ScanAndCount(ctx)
	if err != nil {
		return nil, 0, fromSQLError(err)
	}

	members := make([]models.MemberView, len(entities))
	for i := range entities {
		members[i] = *entity.MembershipToMemberView(&entities[i])
	}

	return members, count, nil
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
