package pg

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/uptrace/bun"
)

func (pg *Pg) InstallKeyCreate(ctx context.Context, installKey *models.InstallKey) (string, error) {
	db := pg.GetConnection(ctx)

	installKey.CreatedAt = clock.Now()
	installKey.UpdatedAt = clock.Now()
	if _, err := db.NewInsert().Model(entity.InstallKeyFromModel(installKey)).Exec(ctx); err != nil {
		return "", fromSQLError(err)
	}

	return installKey.ID, nil
}

func (pg *Pg) InstallKeyConflicts(ctx context.Context, tenantID string, target *models.InstallKeyConflicts) ([]string, bool, error) {
	db := pg.GetConnection(ctx)

	if target.ID == "" && target.Name == "" {
		return []string{}, false, nil
	}

	installKeys := make([]entity.InstallKey, 0)
	query := db.NewSelect().
		Model(&installKeys).
		Column("key_digest", "name").
		Where("namespace_id = ?", tenantID)

	if target.ID != "" && target.Name != "" {
		query = query.Where("key_digest = ? OR name = ?", target.ID, target.Name)
	} else if target.ID != "" {
		query = query.Where("key_digest = ?", target.ID)
	} else if target.Name != "" {
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

func (pg *Pg) InstallKeyList(ctx context.Context, opts ...store.QueryOption) ([]models.InstallKey, int, error) {
	db := pg.GetConnection(ctx)

	entities := make([]entity.InstallKey, 0)

	// The system keys are pinned first (they are the keyless/pairing enrollment queues), the legacy
	// one ahead of the pairing one; the caller's sort applies within each group. Both predicates score
	// the same for every user key, so their relative order is left to the caller's sort.
	query := db.NewSelect().
		Model(&entities).
		OrderExpr("(type = 'user') ASC, (type = 'pairing') ASC")
	var err error
	query, err = applyOptions(ctx, query, opts...)
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

func (pg *Pg) InstallKeyResolve(ctx context.Context, resolver store.InstallKeyResolver, val string, opts ...store.QueryOption) (*models.InstallKey, error) {
	db := pg.GetConnection(ctx)

	column, err := InstallKeyResolverToString(resolver)
	if err != nil {
		return nil, err
	}

	installKey := new(entity.InstallKey)
	query := db.NewSelect().Model(installKey).Where("? = ?", bun.Ident(column), val)
	query, err = applyOptions(ctx, query, opts...)
	if err != nil {
		return nil, err
	}

	if err = query.Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.InstallKeyToModel(installKey), nil
}

func (pg *Pg) InstallKeyResolveSystem(ctx context.Context, tenantID string) (*models.InstallKey, error) {
	return pg.installKeyResolveSystem(ctx, tenantID, models.InstallKeyTypeLegacy)
}

func (pg *Pg) InstallKeyResolveSystemPairing(ctx context.Context, tenantID string) (*models.InstallKey, error) {
	return pg.installKeyResolveSystem(ctx, tenantID, models.InstallKeyTypePairing)
}

// installKeyResolveSystem fetches one of the namespace's system keys by type (legacy or pairing).
func (pg *Pg) installKeyResolveSystem(ctx context.Context, tenantID string, keyType models.InstallKeyType) (*models.InstallKey, error) {
	db := pg.GetConnection(ctx)

	installKey := new(entity.InstallKey)
	if err := db.NewSelect().Model(installKey).Where("namespace_id = ? AND type = ?", tenantID, string(keyType)).Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.InstallKeyToModel(installKey), nil
}

func (pg *Pg) InstallKeyUpdate(ctx context.Context, installKey *models.InstallKey) error {
	db := pg.GetConnection(ctx)

	s := entity.InstallKeyFromModel(installKey)
	s.UpdatedAt = clock.Now()

	// A full-model update (not OmitZero) so zero values are written: a nil expires_at clears the
	// expiry to "never", disabled=false re-enables, usage_limit=0 makes it unlimited. The usage
	// counters (used_times/last_used_at) carry `skipupdate` on the entity, so they stay out of the
	// SET clause and a concurrent enrollment's increment is never clobbered.
	r, err := db.NewUpdate().Model(s).WherePK().Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

func (pg *Pg) InstallKeyIncrementUsage(ctx context.Context, installKey *models.InstallKey) error {
	db := pg.GetConnection(ctx)

	// Guard the increment with the usage limit in the same statement so two concurrent enrollments
	// can't both consume the last use of a limited key. usage_limit = 0 means unlimited.
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

func (pg *Pg) InstallKeyDecrementUsage(ctx context.Context, installKey *models.InstallKey) error {
	db := pg.GetConnection(ctx)

	// Guard at zero so releasing a reservation can never drive used_times negative.
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

func (pg *Pg) InstallKeyEventStampDecision(ctx context.Context, tenantID, deviceUID string, status models.DeviceStatus, at time.Time) error {
	db := pg.GetConnection(ctx)

	// Stamp the device's newest event, so a re-registered device keeps each event's own decision. A
	// device with no event (enrolled before the history, or keyless with no legacy key) updates nothing.
	newest := db.NewSelect().
		Model((*entity.InstallKeyEvent)(nil)).
		Column("id").
		Where("namespace_id = ?", tenantID).
		Where("device_uid = ?", deviceUID).
		Order("created_at DESC").
		Limit(1)

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

func (pg *Pg) InstallKeyEventList(ctx context.Context, tenantID, keyDigest string, opts ...store.QueryOption) ([]models.InstallKeyEvent, int, error) {
	db := pg.GetConnection(ctx)

	entities := make([]entity.InstallKeyEvent, 0)
	// Join the device's current status and when it was last set live via correlated subqueries, so the
	// history can offer an accept/reject action and show the decision time distinct from the enrollment
	// time. Both are null when the device was hard-deleted.
	query := db.NewSelect().
		Model(&entities).
		ModelTableExpr("install_key_events AS e").
		ColumnExpr("e.*").
		ColumnExpr("(SELECT status FROM devices d WHERE d.id = e.device_uid) AS device_status").
		// A device that was removed and re-registered has several events sharing one device row. Mark
		// only the newest event per device_uid as current; the live status + accept/reject action apply
		// there alone. (The decision itself is frozen per-event in decided_status/decided_at.)
		ColumnExpr("(e.created_at = MAX(e.created_at) OVER (PARTITION BY e.device_uid)) AS is_current").
		Where("e.namespace_id = ?", tenantID).
		Where("e.install_key_id = ?", keyDigest)

	var err error
	query, err = applyOptions(ctx, query, opts...)
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

func (pg *Pg) EnrollmentCallbackRedeem(ctx context.Context, jti string, at time.Time) (bool, error) {
	db := pg.GetConnection(ctx)

	// INSERT ... ON CONFLICT DO NOTHING is the atomic claim: the first redemption inserts the row (one
	// row affected), a replay conflicts on the jti primary key and affects none. No read-then-write
	// race window.
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
