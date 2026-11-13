package pg

import (
	"context"
	"errors"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/server/api/store/pg/internal"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// DeviceCreate implements [store.DeviceStore].
func (pg *Pg) DeviceCreate(ctx context.Context, sc scope.Scope, device *models.Device) (string, error) {
	if err := requireBoundedTo(sc, device.TenantID); err != nil {
		return "", err
	}

	db := pg.GetConnection(ctx)

	device.CreatedAt = pg.clock.Now()

	e := entity.DeviceFromModel(device)
	if _, err := db.NewInsert().Model(e).Exec(ctx); err != nil {
		return "", fromSQLError(err)
	}

	return e.ID, nil
}

// DeviceConflicts implements [store.DeviceStore].
func (pg *Pg) DeviceConflicts(ctx context.Context, sc scope.Scope, target *models.DeviceConflicts, opts ...store.QueryOption) ([]string, bool, error) {
	db := pg.GetConnection(ctx)

	if target.Name == "" {
		return []string{}, false, nil
	}

	devices := make([]entity.Device, 0)
	query := db.NewSelect().
		Model(&devices).
		Column("name").
		Where("status = ?", models.DeviceStatusAccepted).
		Where("name = ?", target.Name)

	var err error
	if query, err = applyScopedOptions(ctx, query, sc, opts...); err != nil {
		return nil, false, err
	}

	if err := query.Scan(ctx); err != nil {
		return nil, false, fromSQLError(err)
	}

	seen := make(map[string]bool)
	for _, device := range devices {
		if device.Name == target.Name {
			seen["name"] = true
		}
	}

	conflicts := make([]string, 0, len(seen))
	for field := range seen {
		conflicts = append(conflicts, field)
	}

	return conflicts, len(conflicts) > 0, nil
}

// DeviceList implements [store.DeviceStore].
func (pg *Pg) DeviceList(ctx context.Context, sc scope.Scope, acceptable store.DeviceAcceptable, opts ...store.QueryOption) ([]models.Device, int64, error) {
	entities := make([]entity.Device, 0)

	ctx = context.WithValue(ctx, CtxTableAlias, "device")

	selection, err := applyScopedOptions(ctx, pg.deviceListSelect(ctx, &entities, acceptable), sc, opts...)
	if err != nil {
		return nil, 0, err
	}

	count, err := ScanThenCount(ctx, selection)
	if err != nil {
		return nil, 0, fromSQLError(err)
	}

	return devicesToModels(entities), count, nil
}

// DeviceLocations implements [store.DeviceStore].
func (pg *Pg) DeviceLocations(ctx context.Context, sc scope.Scope, opts ...store.QueryOption) ([]models.DeviceLocation, error) {
	locations := make([]models.DeviceLocation, 0)

	ctx = context.WithValue(ctx, CtxTableAlias, "device")

	selection := pg.GetConnection(ctx).
		NewSelect().
		TableExpr("devices AS device").
		ColumnExpr("round(device.latitude, ?)::float8 AS latitude", models.DeviceLocationPrecision).
		ColumnExpr("round(device.longitude, ?)::float8 AS longitude", models.DeviceLocationPrecision).
		ColumnExpr("count(*) AS devices").
		ColumnExpr("count(*) FILTER (WHERE device.disconnected_at IS NULL AND device.last_seen > ?) AS online", OnlineCutoff(pg.clock)).
		Where("device.status = ?", models.DeviceStatusAccepted).
		Where(internal.DeviceLocated).
		GroupExpr("1, 2")

	selection, err := applyScopedOptions(ctx, selection, sc, opts...)
	if err != nil {
		return nil, err
	}

	if err := selection.Scan(ctx, &locations); err != nil {
		return nil, fromSQLError(err)
	}

	return locations, nil
}

// DeviceListAfter implements [store.DeviceStore].
func (pg *Pg) DeviceListAfter(ctx context.Context, sc scope.Scope, acceptable store.DeviceAcceptable, after *query.Cursor, size int64, opts ...store.QueryOption) ([]models.Device, *query.Cursor, error) {
	size = max(size, query.MinPageSize)
	entities := make([]entity.Device, 0, size+1)

	ctx = context.WithValue(ctx, CtxTableAlias, "device")

	selection := pg.deviceListSelect(ctx, &entities, acceptable).
		OrderExpr("device.created_at ASC, device.id ASC").
		Limit(size + 1)

	if after != nil {
		selection = selection.Where("(device.created_at, device.id) > (?, ?)", after.CreatedAt, after.ID)
	}

	selection, err := applyScopedOptions(ctx, selection, sc, opts...)
	if err != nil {
		return nil, nil, err
	}

	if err := selection.Scan(ctx); err != nil {
		return nil, nil, fromSQLError(err)
	}

	var next *query.Cursor
	if int64(len(entities)) > size {
		entities = entities[:size]
		last := entities[size-1]
		next = &query.Cursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}

	return devicesToModels(entities), next, nil
}

func (pg *Pg) deviceListSelect(ctx context.Context, entities *[]entity.Device, acceptable store.DeviceAcceptable) *bun.SelectQuery {
	onlineExpr, onlineThreshold := deviceExprOnline(OnlineCutoff(pg.clock))

	return pg.GetConnection(ctx).
		NewSelect().
		Model(entities).
		Column("device.*").
		Relation("Namespace").
		Relation("Tags").
		ColumnExpr(onlineExpr, onlineThreshold).
		ColumnExpr(deviceExprAcceptable(acceptable))
}

func devicesToModels(entities []entity.Device) []models.Device {
	devices := make([]models.Device, len(entities))
	for i, e := range entities {
		devices[i] = *entity.DeviceToModel(&e)
	}

	return devices
}

// DeviceListExpiredEphemeral implements [store.DeviceStore].
func (pg *Pg) DeviceListExpiredEphemeral(ctx context.Context, sc scope.Scope) ([]models.Device, error) {
	db := pg.GetConnection(ctx)
	now := pg.clock.Now()

	entities := make([]entity.Device, 0)
	query, err := applyScopedOptions(ctx, db.NewSelect().
		Model(&entities).
		Where("ephemeral = TRUE").
		Where("NOT (disconnected_at IS NULL AND last_seen > ?)", OnlineCutoff(pg.clock)).
		Where("status = ? OR disconnected_at > status_updated_at", models.DeviceStatusAccepted).
		Where("GREATEST(last_seen, status_updated_at) + make_interval(mins => ephemeral_timeout) < ?", now).
		Limit(1000), sc)
	if err != nil {
		return nil, err
	}

	if err := query.Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	devices := make([]models.Device, len(entities))
	for i, e := range entities {
		devices[i] = *entity.DeviceToModel(&e)
	}

	return devices, nil
}

// DeviceResolve implements [store.DeviceStore].
func (pg *Pg) DeviceResolve(ctx context.Context, sc scope.Scope, resolver store.DeviceResolver, val string, opts ...store.QueryOption) (*models.Device, error) {
	db := pg.GetConnection(ctx)

	column, err := DeviceResolverToString(resolver)
	if err != nil {
		return nil, err
	}

	d := new(entity.Device)

	onlineExpr, onlineThreshold := deviceExprOnline(OnlineCutoff(pg.clock))
	query := db.
		NewSelect().
		Model(d).
		Where("? = ?", bun.Ident("device."+column), val).
		Column("device.*").
		Relation("Namespace").
		Relation("Tags").
		ColumnExpr(onlineExpr, onlineThreshold)

	ctx = context.WithValue(ctx, CtxTableAlias, "device")

	query, err = applyScopedOptions(ctx, query, sc, opts...)
	if err != nil {
		return nil, err
	}

	if err = query.Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.DeviceToModel(d), nil
}

// DeviceUpdate implements [store.DeviceStore].
func (pg *Pg) DeviceUpdate(ctx context.Context, sc scope.Scope, device *models.Device) error {
	return pg.deviceUpdate(ctx, sc, device, false)
}

// DeviceUpdateUnlessRemoved implements [store.DeviceStore].
func (pg *Pg) DeviceUpdateUnlessRemoved(ctx context.Context, sc scope.Scope, device *models.Device) error {
	return pg.deviceUpdate(ctx, sc, device, true)
}

// DeviceAssertNotRemoved implements [store.DeviceStore].
func (pg *Pg) DeviceAssertNotRemoved(ctx context.Context, sc scope.Scope, device *models.Device) error {
	if err := requireBoundedTo(sc, device.TenantID); err != nil {
		return err
	}

	query, err := whereBounded(pg.GetConnection(ctx).NewSelect().
		Model((*entity.Device)(nil)).
		Where("id = ?", device.UID).
		Where("removed_at IS NULL"), sc, "namespace_id")
	if err != nil {
		return err
	}

	live, err := query.Exists(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if !live {
		return store.ErrNoDocuments
	}

	return nil
}

func (pg *Pg) deviceUpdate(ctx context.Context, sc scope.Scope, device *models.Device, unlessRemoved bool) error {
	if err := requireBoundedTo(sc, device.TenantID); err != nil {
		return err
	}

	db := pg.GetConnection(ctx)

	d := entity.DeviceFromModel(device)
	d.UpdatedAt = pg.clock.Now()

	query, err := whereBounded(db.NewUpdate().Model(d).Where("id = ?", d.ID), sc, "namespace_id")
	if err != nil {
		return err
	}

	if unlessRemoved {
		query = query.Where("removed_at IS NULL")
	}

	r, err := query.Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	if device.Status == models.DeviceStatusRemoved {
		return pg.deviceForgetKeyVerification(ctx, d.ID, d.NamespaceID)
	}

	if d.KeyVerifiedAt == nil {
		return nil
	}

	return pg.deviceRecordKeyVerification(ctx, d.ID, d.NamespaceID, d.KeyVerifiedAt)
}

func (pg *Pg) deviceRecordKeyVerification(ctx context.Context, id, namespaceID string, at *time.Time) error {
	if _, err := pg.GetConnection(ctx).NewUpdate().
		Model((*entity.Device)(nil)).
		Set("key_verified_at = ?", at).
		Where("id = ?", id).
		Where("namespace_id = ?", namespaceID).
		Where("key_verified_at IS NULL").
		Where("removed_at IS NULL").
		Exec(ctx); err != nil {
		return fromSQLError(err)
	}

	return nil
}

func (pg *Pg) deviceForgetKeyVerification(ctx context.Context, id, namespaceID string) error {
	if _, err := pg.GetConnection(ctx).NewUpdate().
		Model((*entity.Device)(nil)).
		Set("key_verified_at = NULL").
		Where("id = ?", id).
		Where("namespace_id = ?", namespaceID).
		Exec(ctx); err != nil {
		return fromSQLError(err)
	}

	return nil
}

// DeviceSetCustomField implements [store.DeviceStore].
func (pg *Pg) DeviceSetCustomField(ctx context.Context, sc scope.Scope, uid, key, value string) error {
	db := pg.GetConnection(ctx)

	query, err := whereBounded(db.NewUpdate().
		Model((*entity.Device)(nil)).
		Set("custom_fields = jsonb_set(coalesce(custom_fields, '{}'::jsonb), ?, to_jsonb(?::text))", pgdialect.Array([]string{key}), value).
		Where("id = ?", uid), sc, "namespace_id")
	if err != nil {
		return err
	}

	r, err := query.Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

// DeviceDeleteCustomField implements [store.DeviceStore].
func (pg *Pg) DeviceDeleteCustomField(ctx context.Context, sc scope.Scope, uid, key string) error {
	db := pg.GetConnection(ctx)

	query, err := whereBounded(db.NewUpdate().
		Model((*entity.Device)(nil)).
		Set("custom_fields = coalesce(custom_fields, '{}'::jsonb) - ?", key).
		Where("id = ?", uid), sc, "namespace_id")
	if err != nil {
		return err
	}

	r, err := query.Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

// DeviceHeartbeat implements [store.DeviceStore].
func (pg *Pg) DeviceHeartbeat(ctx context.Context, sc scope.Scope, beats []store.DeviceBeat, skipWithin time.Duration) ([]string, error) {
	if !sc.IsValid() {
		return nil, store.ErrInvalidScope
	}

	db := pg.GetConnection(ctx)

	ids, namespaceIDs, times := deviceBeatsInScope(sc, beats)

	present := []string{}
	if err := db.NewRaw(
		`WITH _data AS (
			SELECT id, namespace_id, at
			FROM unnest(?::varchar[], ?::uuid[], ?::timestamptz[]) AS _data(id, namespace_id, at)),
		_written AS (
			UPDATE devices AS device
			SET last_seen = _data.at,
				disconnected_at = CASE WHEN device.disconnected_at > _data.at THEN device.disconnected_at END
			FROM _data
			WHERE device.id = _data.id
				AND device.namespace_id = _data.namespace_id
				AND device.status <> ?
				AND (device.disconnected_at IS NOT NULL OR device.last_seen <= _data.at - make_interval(secs => ?))
		)
		SELECT device.id FROM devices AS device
			JOIN _data ON device.id = _data.id AND device.namespace_id = _data.namespace_id
			WHERE device.status <> ?`,
		pgdialect.Array(ids), pgdialect.Array(namespaceIDs), pgdialect.Array(times),
		models.DeviceStatusRemoved, max(skipWithin, 0).Seconds(), models.DeviceStatusRemoved,
	).Scan(ctx, &present); err != nil {
		return nil, fromSQLError(err)
	}

	found := make(map[string]struct{}, len(present))
	for _, id := range present {
		found[id] = struct{}{}
	}

	gone := []string{}
	for _, beat := range beats {
		if _, ok := found[beat.UID]; !ok {
			gone = append(gone, beat.UID)
		}
	}

	return gone, nil
}

// DeviceOffline implements [store.DeviceStore].
func (pg *Pg) DeviceOffline(ctx context.Context, sc scope.Scope, uid string, disconnectedAt time.Time) error {
	db := pg.GetConnection(ctx)

	query, err := whereBounded(db.NewUpdate().
		Model((*entity.Device)(nil)).
		Set("disconnected_at = ?", disconnectedAt).
		Where("id = ?", uid), sc, "namespace_id")
	if err != nil {
		return err
	}

	r, err := query.Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

// DeviceSweepStalePresence implements [store.DeviceStore].
func (pg *Pg) DeviceSweepStalePresence(ctx context.Context, sc scope.Scope, staleBefore time.Time, limit int64) (int64, error) {
	db := pg.GetConnection(ctx)

	unclosedSince := func(q bun.QueryBuilder) bun.QueryBuilder {
		return q.
			Where("disconnected_at IS NULL").
			Where("last_seen > 'epoch'").
			Where("last_seen < ?", staleBefore).
			Where("status <> ?", models.DeviceStatusRemoved)
	}

	stale, err := whereInScope(db.NewSelect().
		Model((*entity.Device)(nil)).
		Column("id").
		ApplyQueryBuilder(unclosedSince).
		Limit(limit), sc, "namespace_id")
	if err != nil {
		return 0, err
	}

	r, err := db.NewUpdate().
		Model((*entity.Device)(nil)).
		Set("disconnected_at = last_seen + make_interval(secs => ?)", models.DeviceOnlineWindow.Seconds()).
		Where("id IN (?)", stale).
		ApplyQueryBuilder(unclosedSince).
		Exec(ctx)
	if err != nil {
		return 0, fromSQLError(err)
	}

	stamped, err := r.RowsAffected()
	if err != nil {
		return 0, fromSQLError(err)
	}

	return stamped, nil
}

// DeviceLockMAC implements [store.DeviceStore].
func (pg *Pg) DeviceLockMAC(ctx context.Context, sc scope.Scope, mac string) error {
	tenantID, err := requireBounded(sc)
	if err != nil {
		return err
	}

	return pg.transactionLock(ctx, "device-mac:"+tenantID+":"+mac)
}

// DeviceLockAcceptance implements [store.DeviceStore].
func (pg *Pg) DeviceLockAcceptance(ctx context.Context, sc scope.Scope) error {
	switch {
	case sc.IsBounded():
		return pg.transactionLock(ctx, "device-accept:"+canonicalTenant(sc.TenantID()))
	case sc.IsValid():
		return pg.transactionLock(ctx, "device-accept")
	default:
		return store.ErrInvalidScope
	}
}

// DeviceDelete implements [store.DeviceStore].
func (pg *Pg) DeviceDelete(ctx context.Context, sc scope.Scope, device *models.Device) error {
	if err := requireBoundedTo(sc, device.TenantID); err != nil {
		return err
	}

	deleted, err := pg.deviceDeleteIn(ctx, sc, []string{device.UID})
	switch {
	case err != nil:
		return err
	case len(deleted) < 1:
		return store.ErrNoDocuments
	default:
		return nil
	}
}

// DeviceDeleteMany implements [store.DeviceStore].
func (pg *Pg) DeviceDeleteMany(ctx context.Context, sc scope.Scope, uids []string) ([]store.DeletedDevice, error) {
	if !sc.IsValid() {
		return nil, store.ErrInvalidScope
	}

	return pg.deviceDeleteIn(ctx, sc, uids)
}

func (pg *Pg) deviceDeleteIn(ctx context.Context, sc scope.Scope, uids []string) ([]store.DeletedDevice, error) {
	db := pg.GetConnection(ctx)
	fn := pg.deviceDeleteManyFn(ctx, sc, uids)

	if tx, ok := db.(bun.Tx); ok {
		return fn(tx)
	} else {
		tx, err := pg.driver.BeginTx(ctx, nil)
		if err != nil {
			return nil, fromSQLError(err)
		}

		defer func() {
			if p := recover(); p != nil {
				_ = tx.Rollback()
				panic(p)
			}
		}()

		deleted, err := fn(tx)
		if err != nil {
			_ = tx.Rollback()

			return nil, err
		}

		if err := tx.Commit(); err != nil {
			return nil, fromSQLError(err)
		}

		return deleted, nil
	}
}

func (pg *Pg) deviceDeleteManyFn(ctx context.Context, sc scope.Scope, uids []string) func(tx bun.Tx) ([]store.DeletedDevice, error) {
	return func(tx bun.Tx) ([]store.DeletedDevice, error) {
		deleted := []store.DeletedDevice{}
		if len(uids) == 0 {
			return deleted, nil
		}

		rows := []entity.Device{}
		query, err := whereInScope(tx.NewDelete().
			Model(&rows).
			Where("id IN (?)", bun.List(uids)).
			Returning("id, namespace_id, status"), sc, "namespace_id")
		if errors.Is(err, store.ErrNoDocuments) {
			return deleted, nil
		}

		if err != nil {
			return nil, err
		}

		if err := query.Scan(ctx); err != nil {
			return nil, fromSQLError(err)
		}

		if len(rows) == 0 {
			return deleted, nil
		}

		ids := make([]string, len(rows))
		for i, row := range rows {
			ids[i] = row.ID
			deleted = append(deleted, store.DeletedDevice{UID: row.ID, TenantID: row.NamespaceID, Status: models.DeviceStatus(row.Status)})
		}

		if _, err := tx.NewDelete().
			Model((*entity.Session)(nil)).
			Where("device_id IN (?)", bun.List(ids)).
			Exec(ctx); err != nil {
			return nil, fromSQLError(err)
		}

		return deleted, nil
	}
}

// OnlineCutoff returns the instant, at c's now, that a device's last_seen must be after to count
// as online: [models.DeviceOnlineWindow] before now.
func OnlineCutoff(c clock.Clock) time.Time {
	return c.Now().Add(-models.DeviceOnlineWindow)
}

// DeviceIsOnline reports whether the device should be considered online
// based on the given threshold time. A device is online when it has not
// been explicitly disconnected and was last seen after the threshold.
func DeviceIsOnline(d *entity.Device, threshold time.Time) bool {
	return d.DisconnectedAt.IsZero() && d.LastSeen.After(threshold)
}

func deviceExprOnline(threshold time.Time) (string, time.Time) {
	return `CASE
		WHEN "device"."disconnected_at" IS NULL AND "device"."last_seen" > ?
		THEN true
		ELSE false
		END AS "online"`, threshold
}

func deviceExprAcceptable(mode store.DeviceAcceptable) string {
	switch mode {
	case store.DeviceAcceptableFromRemoved:
		return `"device"."status" = 'removed' AS "acceptable"`
	case store.DeviceAcceptableAsFalse:
		return `false AS "acceptable"`
	case store.DeviceAcceptableIfNotAccepted:
		return `CASE WHEN "device"."status" <> 'accepted' THEN true ELSE false END AS "acceptable"`
	default:
		return `true AS "acceptable"`
	}
}

func deviceBeatsInScope(sc scope.Scope, beats []store.DeviceBeat) (ids, namespaceIDs []string, times []time.Time) {
	bounded := ""
	if sc.IsBounded() {
		parsed, err := uuid.Parse(sc.TenantID())
		if err != nil {
			return nil, nil, nil
		}

		bounded = parsed.String()
	}

	ids = make([]string, 0, len(beats))
	namespaceIDs = make([]string, 0, len(beats))
	times = make([]time.Time, 0, len(beats))
	for _, beat := range beats {
		parsed, err := uuid.Parse(beat.TenantID)
		if err != nil {
			continue
		}

		namespaceID := parsed.String()
		if sc.IsBounded() && namespaceID != bounded {
			continue
		}

		ids = append(ids, beat.UID)
		namespaceIDs = append(namespaceIDs, namespaceID)
		times = append(times, beat.At)
	}

	return ids, namespaceIDs, times
}

// DeviceResolverToString returns the column resolver selects, reporting
// [store.ErrResolverNotFound] for one this store does not implement.
func DeviceResolverToString(resolver store.DeviceResolver) (string, error) {
	switch resolver {
	case store.DeviceUIDResolver:
		return "id", nil
	case store.DeviceHostnameResolver:
		return "name", nil
	case store.DeviceMACResolver:
		return "mac", nil
	case store.DevicePublicKeyResolver:
		return "public_key", nil
	default:
		return "", store.ErrResolverNotFound
	}
}

// DeviceSetOwner implements [store.DeviceStore].
func (pg *Pg) DeviceSetOwner(ctx context.Context, sc scope.Scope, uid, ownerID string) error {
	db := pg.GetConnection(ctx)

	var owner any
	if ownerID != "" {
		owner = ownerID
	}

	query, err := whereBounded(db.NewUpdate().
		Model((*entity.Device)(nil)).
		Set("owner_id = ?", owner).
		Where("id = ?", uid), sc, "namespace_id")
	if err != nil {
		return err
	}

	r, err := query.Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return nil
}
