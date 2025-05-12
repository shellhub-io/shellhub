package pg

import (
	"context" //nolint:gosec
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func (pg *Pg) DeviceCreate(ctx context.Context, device *models.Device) (string, error) {
	device.CreatedAt = clock.Now()
	device.UpdatedAt = clock.Now()

	e := entity.DeviceFromModel(device)
	if _, err := pg.driver.NewInsert().Model(e).Exec(ctx); err != nil {
		return "", fromSqlError(err)
	}

	return e.ID, nil
}

func (pg *Pg) DeviceConflicts(ctx context.Context, target *models.DeviceConflicts) ([]string, bool, error) {
	devices := make([]map[string]any, 0)
	if err := pg.driver.NewSelect().Model((*entity.Device)(nil)).Column("name").Where("name = ?", target.Name).Scan(ctx, &devices); err != nil {
		return nil, false, fromSqlError(err)
	}

	conflicts := make([]string, 0)
	for _, device := range devices {
		if device["name"] == target.Name {
			conflicts = append(conflicts, "name")
		}
	}

	return conflicts, len(conflicts) > 0, nil
}

func (pg *Pg) DeviceList(ctx context.Context, opts ...store.QueryOption) ([]models.Device, int, error) {
	entities := make([]entity.Device, 0)

	query := pg.driver.
		NewSelect().
		Model(&entities).
		Column("device.*").
		Relation("Namespace").
		ColumnExpr(`
			CASE
				WHEN "device"."disconnected_at" IS NULL AND "device"."seen_at" > ?
				THEN true
				ELSE false
			END AS "online"`,
			time.Now().Add(-2*time.Minute),
		).
		ColumnExpr(`
			CASE
				WHEN "device"."status" <> 'accepted'
				THEN true
				ELSE false
			END AS "acceptable"`,
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
		devices[i] = *entity.DeviceToModel(&e)
	}

	return devices, count, nil
}

func (pg *Pg) DeviceGet(ctx context.Context, ident store.DeviceIdent, val string) (*models.Device, error) {
	d := new(entity.Device)

	query := pg.driver.
		NewSelect().
		Model(d).
		Where("? = ?", bun.Ident("device."+ident), val).
		Column("device.*").
		Relation("Namespace").
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

	return entity.DeviceToModel(d), nil
}

func (pg *Pg) DeviceSave(ctx context.Context, device *models.Device) error {
	d := entity.DeviceFromModel(device)
	d.UpdatedAt = clock.Now()
	_, err := pg.driver.NewUpdate().Model(d).WherePK().Exec(ctx)

	return fromSqlError(err)
}

func (pg *Pg) DeviceUpdateSeenAt(ctx context.Context, ids []string, to time.Time) (int64, error) {
	r, err := pg.driver.NewUpdate().
		Model((*entity.Device)(nil)).
		Set("seen_at = ?", to).
		Set("disconnected_at = NULL").
		TableExpr("(SELECT unnest(?::varchar[]) as id) as _data", pgdialect.Array(ids)).
		Where("device.id = _data.id").
		Exec(ctx)
	if err != nil {
		return 0, fromSqlError(err)
	}

	return r.RowsAffected()
}

func (pg *Pg) DeviceDelete(ctx context.Context, device *models.Device) error {
	d := entity.DeviceFromModel(device)
	_, err := pg.driver.NewDelete().Model(d).WherePK().Exec(ctx)

	return fromSqlError(err)
}
