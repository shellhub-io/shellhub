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
	"github.com/uptrace/bun"
)

// InstallKeyCreate implements [store.InstallKeyStore].
func (pg *Pg) InstallKeyCreate(ctx context.Context, installKey *models.InstallKey) (string, error) {
	db := pg.GetConnection(ctx)

	installKey.CreatedAt = clock.Now()
	installKey.UpdatedAt = clock.Now()
	if _, err := db.NewInsert().Model(entity.InstallKeyFromModel(installKey)).Exec(ctx); err != nil {
		return "", fromSQLError(err)
	}

	return installKey.ID, nil
}

// InstallKeyConflicts implements [store.InstallKeyStore].
func (pg *Pg) InstallKeyConflicts(ctx context.Context, sc scope.Scope, target *models.InstallKeyConflicts) ([]string, bool, error) {
	db := pg.GetConnection(ctx)

	if target.ID == "" && target.Name == "" {
		return []string{}, false, nil
	}

	installKeys := make([]entity.InstallKey, 0)
	query := db.NewSelect().
		Model(&installKeys).
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
	for _, installKey := range installKeys {
		if target.ID != "" && installKey.KeyDigest == target.ID {
			seen["id"] = true
		}

		if target.Name != "" && installKey.Name == target.Name {
			seen["name"] = true
		}
	}

	conflicts := make([]string, 0, len(seen))
	for field := range seen {
		conflicts = append(conflicts, field)
	}

	return conflicts, len(conflicts) > 0, nil
}

// InstallKeyList implements [store.InstallKeyStore].
func (pg *Pg) InstallKeyList(ctx context.Context, sc scope.Scope, opts ...store.QueryOption) ([]models.InstallKey, int, error) {
	db := pg.GetConnection(ctx)

	entities := make([]entity.InstallKey, 0)

	query := db.NewSelect().
		Model(&entities).
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

	installKeys := make([]models.InstallKey, len(entities))
	for i, e := range entities {
		installKeys[i] = *entity.InstallKeyToModel(&e)
	}

	return installKeys, count, nil
}

// InstallKeyResolve implements [store.InstallKeyStore].
func (pg *Pg) InstallKeyResolve(ctx context.Context, sc scope.Scope, resolver store.InstallKeyResolver, val string, opts ...store.QueryOption) (*models.InstallKey, error) {
	db := pg.GetConnection(ctx)

	column, err := InstallKeyResolverToString(resolver)
	if err != nil {
		return nil, err
	}

	installKey := new(entity.InstallKey)
	query := db.NewSelect().Model(installKey).Where("? = ?", bun.Ident(column), val)
	query, err = applyScopedOptions(ctx, query, sc, opts...)
	if err != nil {
		return nil, err
	}

	if err = query.Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.InstallKeyToModel(installKey), nil
}

// InstallKeyResolveSystem implements [store.InstallKeyStore].
func (pg *Pg) InstallKeyResolveSystem(ctx context.Context, sc scope.Scope) (*models.InstallKey, error) {
	return pg.installKeyResolveSystem(ctx, sc, models.InstallKeyTypeLegacy)
}

// InstallKeyResolveSystemPairing implements [store.InstallKeyStore].
func (pg *Pg) InstallKeyResolveSystemPairing(ctx context.Context, sc scope.Scope) (*models.InstallKey, error) {
	return pg.installKeyResolveSystem(ctx, sc, models.InstallKeyTypePairing)
}

func (pg *Pg) installKeyResolveSystem(ctx context.Context, sc scope.Scope, keyType models.InstallKeyType) (*models.InstallKey, error) {
	db := pg.GetConnection(ctx)

	tenantID, err := requireBounded(sc)
	if err != nil {
		return nil, err
	}

	installKey := new(entity.InstallKey)
	query := db.NewSelect().
		Model(installKey).
		Where("type = ?", string(keyType)).
		Where("namespace_id = ?", tenantID)

	if err := query.Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.InstallKeyToModel(installKey), nil
}

// InstallKeyUpdate implements [store.InstallKeyStore].
func (pg *Pg) InstallKeyUpdate(ctx context.Context, installKey *models.InstallKey) error {
	db := pg.GetConnection(ctx)

	s := entity.InstallKeyFromModel(installKey)
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

// InstallKeyIncrementUsage implements [store.InstallKeyStore].
func (pg *Pg) InstallKeyIncrementUsage(ctx context.Context, installKey *models.InstallKey) error {
	db := pg.GetConnection(ctx)

	r, err := db.NewUpdate().
		Model((*entity.InstallKey)(nil)).
		Set("used_times = used_times + 1").
		Set("last_used_at = ?", clock.Now()).
		Set("updated_at = ?", clock.Now()).
		Where("key_digest = ? AND namespace_id = ?", installKey.ID, installKey.TenantID).
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

// InstallKeyDecrementUsage implements [store.InstallKeyStore].
func (pg *Pg) InstallKeyDecrementUsage(ctx context.Context, installKey *models.InstallKey) error {
	db := pg.GetConnection(ctx)

	r, err := db.NewUpdate().
		Model((*entity.InstallKey)(nil)).
		Set("used_times = used_times - 1").
		Set("updated_at = ?", clock.Now()).
		Where("key_digest = ? AND namespace_id = ?", installKey.ID, installKey.TenantID).
		Where("used_times > 0").
		Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

// InstallKeyEventCreate implements [store.InstallKeyStore].
func (pg *Pg) InstallKeyEventCreate(ctx context.Context, event *models.InstallKeyEvent) error {
	db := pg.GetConnection(ctx)

	e := entity.InstallKeyEventFromModel(event)
	e.ID = uuid.Generate()
	e.CreatedAt = clock.Now()

	if _, err := db.NewInsert().Model(e).Exec(ctx); err != nil {
		return fromSQLError(err)
	}

	return nil
}

// InstallKeyEventStampDecision implements [store.InstallKeyStore].
func (pg *Pg) InstallKeyEventStampDecision(ctx context.Context, sc scope.Scope, deviceUID string, status models.DeviceStatus, at time.Time) error {
	db := pg.GetConnection(ctx)

	newest := db.NewSelect().
		Model((*entity.InstallKeyEvent)(nil)).
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
		Model((*entity.InstallKeyEvent)(nil)).
		Set("decided_status = ?", string(status)).
		Set("decided_at = ?", at).
		Where("id IN (?)", newest).
		Exec(ctx); err != nil {
		return fromSQLError(err)
	}

	return nil
}

// InstallKeyEventList implements [store.InstallKeyStore].
func (pg *Pg) InstallKeyEventList(ctx context.Context, sc scope.Scope, keyDigest string, opts ...store.QueryOption) ([]models.InstallKeyEvent, int, error) {
	db := pg.GetConnection(ctx)

	ctx = context.WithValue(ctx, CtxTableAlias, "e")

	entities := make([]entity.InstallKeyEvent, 0)
	query := db.NewSelect().
		Model(&entities).
		ModelTableExpr("install_key_events AS e").
		ColumnExpr("e.*").
		ColumnExpr("(SELECT status FROM devices d WHERE d.id = e.device_uid) AS device_status").
		ColumnExpr("(e.created_at = MAX(e.created_at) OVER (PARTITION BY e.device_uid)) AS is_current").
		Where("e.install_key_id = ?", keyDigest)

	var err error
	query, err = applyScopedOptions(ctx, query, sc, opts...)
	if err != nil {
		return nil, 0, err
	}

	count, err := query.ScanAndCount(ctx)
	if err != nil {
		return nil, 0, fromSQLError(err)
	}

	events := make([]models.InstallKeyEvent, len(entities))
	for i, e := range entities {
		events[i] = *entity.InstallKeyEventToModel(&e)
	}

	return events, count, nil
}

// EnrollmentCallbackRedeem implements [store.InstallKeyStore].
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

// EnrollmentCallbackCleanup implements [store.InstallKeyStore].
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

// InstallKeyResolverToString returns the column resolver selects, reporting
// [store.ErrResolverNotFound] for one this store does not implement.
func InstallKeyResolverToString(resolver store.InstallKeyResolver) (string, error) {
	switch resolver {
	case store.InstallKeyIDResolver:
		return "key_digest", nil
	case store.InstallKeyNameResolver:
		return "name", nil
	default:
		return "", store.ErrResolverNotFound
	}
}
