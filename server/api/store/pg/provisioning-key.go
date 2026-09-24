package pg

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/pg/entity"
)

// ProvisioningKeyCreate implements [store.ProvisioningKeyStore].
func (pg *Pg) ProvisioningKeyCreate(ctx context.Context, provisioningKey *models.ProvisioningKey) (string, error) {
	db := pg.GetConnection(ctx)

	provisioningKey.CreatedAt = clock.Now()
	provisioningKey.UpdatedAt = clock.Now()
	if _, err := db.NewInsert().Model(entity.ProvisioningKeyFromModel(provisioningKey)).Exec(ctx); err != nil {
		return "", fromSQLError(err)
	}

	return provisioningKey.ID, nil
}

// ProvisioningKeyConflicts implements [store.ProvisioningKeyStore].
func (pg *Pg) ProvisioningKeyConflicts(ctx context.Context, sc scope.Scope, target *models.ProvisioningKeyConflicts) ([]string, bool, error) {
	db := pg.GetConnection(ctx)

	if target.ID == "" && target.Name == "" {
		return []string{}, false, nil
	}

	provisioningKeys := make([]entity.ProvisioningKey, 0)
	query := db.NewSelect().
		Model(&provisioningKeys).
		Column("key_digest", "name")

	query, err := applyScopedOptions(ctx, query, sc)
	if err != nil {
		return nil, false, err
	}

	switch {
	case target.ID != "" && target.Name != "":
		query = query.Where("key_digest = ? OR name = ?", target.ID, target.Name)
	case target.ID != "":
		query = query.Where("key_digest = ?", target.ID)
	case target.Name != "":
		query = query.Where("name = ?", target.Name)
	}

	if err := query.Scan(ctx); err != nil {
		return nil, false, fromSQLError(err)
	}

	seen := make(map[string]bool)
	for _, provisioningKey := range provisioningKeys {
		if target.ID != "" && provisioningKey.KeyDigest == target.ID {
			seen["id"] = true
		}

		if target.Name != "" && provisioningKey.Name == target.Name {
			seen["name"] = true
		}
	}

	conflicts := make([]string, 0, len(seen))
	for field := range seen {
		conflicts = append(conflicts, field)
	}

	return conflicts, len(conflicts) > 0, nil
}

const pendingDevicesExpr = `(
	SELECT COUNT(*) FROM devices d
	WHERE d.provisioning_key_id = provisioning_key.key_digest
	  AND d.namespace_id = provisioning_key.namespace_id
	  AND d.status = 'pending'
) AS pending_devices`

// ProvisioningKeyList implements [store.ProvisioningKeyStore].
func (pg *Pg) ProvisioningKeyList(ctx context.Context, sc scope.Scope, opts ...store.QueryOption) ([]models.ProvisioningKey, int, error) {
	db := pg.GetConnection(ctx)

	entities := make([]entity.ProvisioningKey, 0)

	query := db.NewSelect().
		Model(&entities).
		ColumnExpr("provisioning_key.*").
		ColumnExpr(pendingDevicesExpr).
		OrderExpr("(type = 'user') ASC, (type = 'pairing') ASC")
	var err error
	query, err = applyScopedOptions(ctx, query, sc, opts...)
	if err != nil {
		return nil, 0, err
	}

	count, err := query.ScanAndCount(ctx)
	if err != nil {
		return nil, 0, fromSQLError(err)
	}

	provisioningKeys := make([]models.ProvisioningKey, len(entities))
	for i, e := range entities {
		provisioningKeys[i] = *entity.ProvisioningKeyToModel(&e)
	}

	return provisioningKeys, count, nil
}

// ProvisioningKeyResolve implements [store.ProvisioningKeyStore].
func (pg *Pg) ProvisioningKeyResolve(ctx context.Context, sc scope.Scope, resolver store.ProvisioningKeyResolver, val string, opts ...store.QueryOption) (*models.ProvisioningKey, error) {
	column, err := ProvisioningKeyResolverToString(resolver)
	if err != nil {
		return nil, err
	}

	provisioningKey, err := resolveUnique[entity.ProvisioningKey](ctx, pg.GetConnection(ctx), sc, column, val, opts...)
	if err != nil {
		return nil, err
	}

	return entity.ProvisioningKeyToModel(provisioningKey), nil
}

// ProvisioningKeyResolveSystem implements [store.ProvisioningKeyStore].
func (pg *Pg) ProvisioningKeyResolveSystem(ctx context.Context, sc scope.Scope) (*models.ProvisioningKey, error) {
	return pg.provisioningKeyResolveSystem(ctx, sc, models.ProvisioningKeyTypeLegacy)
}

// ProvisioningKeyResolveSystemPairing implements [store.ProvisioningKeyStore].
func (pg *Pg) ProvisioningKeyResolveSystemPairing(ctx context.Context, sc scope.Scope) (*models.ProvisioningKey, error) {
	return pg.provisioningKeyResolveSystem(ctx, sc, models.ProvisioningKeyTypePairing)
}

func (pg *Pg) provisioningKeyResolveSystem(ctx context.Context, sc scope.Scope, keyType models.ProvisioningKeyType) (*models.ProvisioningKey, error) {
	db := pg.GetConnection(ctx)

	tenantID, err := requireBounded(sc)
	if err != nil {
		return nil, err
	}

	provisioningKey := new(entity.ProvisioningKey)
	query := db.NewSelect().
		Model(provisioningKey).
		Where("type = ?", string(keyType)).
		Where("namespace_id = ?", tenantID)

	if err := query.Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.ProvisioningKeyToModel(provisioningKey), nil
}

// ProvisioningKeyUpdate implements [store.ProvisioningKeyStore].
func (pg *Pg) ProvisioningKeyUpdate(ctx context.Context, provisioningKey *models.ProvisioningKey) error {
	db := pg.GetConnection(ctx)

	s := entity.ProvisioningKeyFromModel(provisioningKey)
	s.UpdatedAt = clock.Now()

	r, err := db.NewUpdate().Model(s).WherePK().Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

// ProvisioningKeyIncrementUsage implements [store.ProvisioningKeyStore].
func (pg *Pg) ProvisioningKeyIncrementUsage(ctx context.Context, provisioningKey *models.ProvisioningKey) error {
	db := pg.GetConnection(ctx)

	r, err := db.NewUpdate().
		Model((*entity.ProvisioningKey)(nil)).
		Set("used_times = used_times + 1").
		Set("last_used_at = ?", clock.Now()).
		Set("updated_at = ?", clock.Now()).
		Where("key_digest = ? AND namespace_id = ?", provisioningKey.ID, provisioningKey.TenantID).
		Where("usage_limit = 0 OR used_times < usage_limit").
		Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

// ProvisioningKeyEventCreate implements [store.ProvisioningKeyStore].
func (pg *Pg) ProvisioningKeyEventCreate(ctx context.Context, event *models.ProvisioningKeyEvent) error {
	db := pg.GetConnection(ctx)

	e := entity.ProvisioningKeyEventFromModel(event)
	e.ID = uuid.Generate()
	e.CreatedAt = clock.Now()

	if _, err := db.NewInsert().Model(e).Exec(ctx); err != nil {
		return fromSQLError(err)
	}

	return nil
}

// ProvisioningKeyEventStampDecision implements [store.ProvisioningKeyStore].
func (pg *Pg) ProvisioningKeyEventStampDecision(ctx context.Context, sc scope.Scope, deviceUID string, status models.DeviceStatus, at time.Time) error {
	db := pg.GetConnection(ctx)

	newest := db.NewSelect().
		Model((*entity.ProvisioningKeyEvent)(nil)).
		Column("id").
		Where("device_uid = ?", deviceUID).
		Order("created_at DESC").
		Limit(1)

	tenantID, err := requireBounded(sc)
	if err != nil {
		return err
	}

	newest = newest.Where("namespace_id = ?", tenantID)

	if _, err := db.NewUpdate().
		Model((*entity.ProvisioningKeyEvent)(nil)).
		Set("decided_status = ?", string(status)).
		Set("decided_at = ?", at).
		Where("id IN (?)", newest).
		Exec(ctx); err != nil {
		return fromSQLError(err)
	}

	return nil
}

// ProvisioningKeyEventList implements [store.ProvisioningKeyStore].
func (pg *Pg) ProvisioningKeyEventList(ctx context.Context, sc scope.Scope, keyDigest string, opts ...store.QueryOption) ([]models.ProvisioningKeyEvent, int, error) {
	db := pg.GetConnection(ctx)

	ctx = context.WithValue(ctx, CtxTableAlias, "e")

	entities := make([]entity.ProvisioningKeyEvent, 0)
	query := db.NewSelect().
		Model(&entities).
		ModelTableExpr("provisioning_key_events AS e").
		ColumnExpr("e.*").
		ColumnExpr("(SELECT status FROM devices d WHERE d.id = e.device_uid) AS device_status").
		ColumnExpr("(e.created_at = MAX(e.created_at) OVER (PARTITION BY e.device_uid)) AS is_current").
		Where("e.provisioning_key_id = ?", keyDigest)

	var err error
	query, err = applyScopedOptions(ctx, query, sc, opts...)
	if err != nil {
		return nil, 0, err
	}

	count, err := query.ScanAndCount(ctx)
	if err != nil {
		return nil, 0, fromSQLError(err)
	}

	events := make([]models.ProvisioningKeyEvent, len(entities))
	for i, e := range entities {
		events[i] = *entity.ProvisioningKeyEventToModel(&e)
	}

	return events, count, nil
}

// EnrollmentCallbackRedeem implements [store.ProvisioningKeyStore].
func (pg *Pg) EnrollmentCallbackRedeem(ctx context.Context, jti string, at time.Time) (bool, error) {
	db := pg.GetConnection(ctx)

	res, err := db.NewInsert().
		Model(&entity.EnrollmentCallbackRedemption{JTI: jti, RedeemedAt: at}).
		On("CONFLICT (jti) DO NOTHING").
		Exec(ctx)
	if err != nil {
		return false, fromSQLError(err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, fromSQLError(err)
	}

	return affected > 0, nil
}

// EnrollmentCallbackCleanup implements [store.ProvisioningKeyStore].
func (pg *Pg) EnrollmentCallbackCleanup(ctx context.Context, before time.Time) (int64, error) {
	db := pg.GetConnection(ctx)

	res, err := db.NewDelete().
		Model((*entity.EnrollmentCallbackRedemption)(nil)).
		Where("redeemed_at < ?", before).
		Exec(ctx)
	if err != nil {
		return 0, fromSQLError(err)
	}

	return res.RowsAffected()
}

// ProvisioningKeyResolverToString returns the column resolver selects, reporting
// [store.ErrResolverNotFound] for one this store does not implement.
func ProvisioningKeyResolverToString(resolver store.ProvisioningKeyResolver) (string, error) {
	switch resolver {
	case store.ProvisioningKeyIDResolver:
		return "key_digest", nil
	case store.ProvisioningKeyNameResolver:
		return "name", nil
	default:
		return "", store.ErrResolverNotFound
	}
}
