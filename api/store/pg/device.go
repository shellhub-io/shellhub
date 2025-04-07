package pg

import (
	"context" //nolint:gosec
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/pg/internal/entity"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func (pg *pg) DeviceCreate(ctx context.Context, device *models.Device) (string, error) {
	device.CreatedAt = clock.Now()
	device.UpdatedAt = clock.Now()

	if _, err := pg.driver.NewInsert().Model(&entity.Device{Device: *device}).Exec(ctx); err != nil {
		return "", fromSqlError(err)
	}

	if device.Info != nil {
		device.Info.DeviceID = device.ID
		if _, err := pg.driver.NewInsert().Model(&entity.DeviceInfo{DeviceInfo: *device.Info}).Exec(ctx); err != nil {
			return "", fromSqlError(err)
		}
	}

	if device.Position != nil {
		device.Position.DeviceID = device.ID
		if _, err := pg.driver.NewInsert().Model(&entity.DevicePosition{DevicePosition: *device.Position}).Exec(ctx); err != nil {
			return "", fromSqlError(err)
		}
	}

	return device.ID, nil
}

func (pg *pg) DeviceList(ctx context.Context, opts ...store.QueryOption) ([]models.Device, int, error) {
	entities := make([]entity.Device, 0)

	query := pg.driver.
		NewSelect().
		Model(&entities).
		Column("device.*").
		ColumnExpr(`
			CASE
				WHEN "device"."disconnected_at" IS NULL AND "device"."seen_at" > ?
				THEN true
				ELSE false
			END AS "online"`,
			time.Now().Add(-2*time.Minute),
		)

	if err := applyOptions(ctx, query, opts...); err != nil {
		return nil, 0, fromSqlError(err)
	}

	count, err := query.ScanAndCount(ctx)
	if err != nil {
		return nil, 0, fromSqlError(err)
	}

	devices := make([]models.Device, len(entities))
	for i, e := range entities {
		devices[i] = e.Device
	}

	return devices, count, nil
}

func (pg *pg) DeviceGet(ctx context.Context, ident store.DeviceIdent, val string) (*models.Device, error) {
	d := new(entity.Device)

	query := pg.driver.
		NewSelect().
		Model(d).
		Where("? = ?", bun.Ident(ident), val).
		Column("device.*").
		ColumnExpr(`
			CASE
				WHEN "device"."disconnected_at" IS NULL AND "device"."seen_at" > ?
				THEN true
				ELSE false
			END AS "online"`,
			time.Now().Add(-2*time.Minute),
		)

	if err := query.Scan(ctx); err != nil {
		return nil, fromSqlError(err)
	}

	return &d.Device, nil
}

func (pg *pg) DeviceDelete(ctx context.Context, uid models.UID) error {
	return nil
}

func (pg *pg) DeviceRename(ctx context.Context, uid models.UID, hostname string) error {
	return nil
}

func (pg *pg) DeviceLookup(ctx context.Context, namespace, hostname string) (*models.Device, error) {
	return nil, nil
}

// DeviceUpdateStatus updates the status of a specific device in the devices collection
func (pg *pg) DeviceUpdateStatus(ctx context.Context, uid models.UID, status models.DeviceStatus) error {
	return nil
}

func (pg *pg) DeviceListByUsage(ctx context.Context, tenant string) ([]models.UID, error) {
	return nil, nil
}

func (pg *pg) DeviceGetByMac(ctx context.Context, mac string, tenantID string, status models.DeviceStatus) (*models.Device, error) {
	return nil, nil
}

func (pg *pg) DeviceGetByName(ctx context.Context, name string, tenantID string, status models.DeviceStatus) (*models.Device, error) {
	return nil, nil
}

func (pg *pg) DeviceGetByUID(ctx context.Context, uid models.UID, tenantID string) (*models.Device, error) {
	return nil, nil
}

func (pg *pg) DeviceSetPosition(ctx context.Context, uid models.UID, position models.DevicePosition) error {
	return nil
}

func (pg *pg) DeviceChooser(ctx context.Context, tenantID string, chosen []string) error {
	return nil
}

func (pg *pg) DeviceConflicts(ctx context.Context, target *models.DeviceConflicts) ([]string, bool, error) {
	return nil, false, nil
}

func (pg *pg) DeviceUpdateSeenAt(ctx context.Context, ids []string, to time.Time) (int64, error) {
	r, err := pg.driver.NewUpdate().
		Model((*entity.Device)(nil)).
		Set("seen_at = ?", to).
		TableExpr("(SELECT unnest(?::varchar[]) as id) as _data", pgdialect.Array(ids)).
		Where("device.id = _data.id").
		Exec(ctx)
	if err != nil {
		return 0, fromSqlError(err)
	}

	return r.RowsAffected()
}

func (pg *pg) DeviceRemovedCount(ctx context.Context, tenant string) (int64, error) {
	return int64(0), nil
}

func (pg *pg) DeviceRemovedGet(ctx context.Context, tenant string, uid models.UID) (*models.DeviceRemoved, error) {
	return nil, nil
}

func (pg *pg) DeviceRemovedInsert(ctx context.Context, tenant string, device *models.Device) error { //nolint:revive
	return nil
}

func (pg *pg) DeviceRemovedDelete(ctx context.Context, tenant string, uid models.UID) error {
	return nil
}

func (pg *pg) DeviceRemovedList(ctx context.Context, tenant string, paginator query.Paginator, filters query.Filters, sorter query.Sorter) ([]models.DeviceRemoved, int, error) {
	return nil, 0, nil
}

func (pg *pg) DevicePushTag(ctx context.Context, uid models.UID, tag string) error {
	return nil
}

func (pg *pg) DevicePullTag(ctx context.Context, uid models.UID, tag string) error {
	return nil
}

func (pg *pg) DeviceSetTags(ctx context.Context, uid models.UID, tags []string) (int64, int64, error) {
	return int64(0), int64(0), nil
}

func (pg *pg) DeviceBulkRenameTag(ctx context.Context, tenant, currentTag, newTag string) (int64, error) {
	return int64(0), nil
}

func (pg *pg) DeviceBulkDeleteTag(ctx context.Context, tenant, tag string) (int64, error) {
	return int64(0), nil
}

func (pg *pg) DeviceGetTags(ctx context.Context, tenant string) ([]string, int, error) {
	return nil, 0, nil
}
