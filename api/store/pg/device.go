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
	db := pg.getConnection(ctx)

	device.CreatedAt = clock.Now()

	e := entity.DeviceFromModel(device)
	if _, err := db.NewInsert().Model(e).Exec(ctx); err != nil {
		return "", fromSQLError(err)
	}

	return e.ID, nil
}

func (pg *Pg) DeviceConflicts(ctx context.Context, target *models.DeviceConflicts) ([]string, bool, error) {
	db := pg.getConnection(ctx)

	devices := make([]map[string]any, 0)
	if err := db.NewSelect().Model((*entity.Device)(nil)).Column("name").Where("name = ?", target.Name).Scan(ctx, &devices); err != nil {
		return nil, false, fromSQLError(err)
	}

	conflicts := make([]string, 0)
	for _, device := range devices {
		if device["name"] == target.Name {
			conflicts = append(conflicts, "name")
		}
	}

	return conflicts, len(conflicts) > 0, nil
}

func (pg *Pg) DeviceList(ctx context.Context, acceptable store.DeviceAcceptable, opts ...store.QueryOption) ([]models.Device, int, error) {
	db := pg.getConnection(ctx)

	entities := make([]entity.Device, 0)

	query := db.
		NewSelect().
		Model(&entities).
		Column("device.*").
		Relation("Namespace").
		ColumnExpr(string(deviceExprOnline), time.Now().Add(-2*time.Minute)).
		ColumnExpr(deviceExprAcepptable(acceptable))

	if err := applyOptions(ctx, query, opts...); err != nil {
		return nil, 0, fromSQLError(err)
	}

	count, err := query.ScanAndCount(ctx)
	if err != nil {
		return nil, 0, fromSQLError(err)
	}

	devices := make([]models.Device, len(entities))
	for i, e := range entities {
		devices[i] = *entity.DeviceToModel(&e)
	}

	return devices, count, nil
}

func (pg *Pg) DeviceResolve(ctx context.Context, resolver store.DeviceResolver, val string, opts ...store.QueryOption) (*models.Device, error) {
	db := pg.getConnection(ctx)

	column, err := DeviceResolverToString(resolver)
	if err != nil {
		return nil, err
	}

	d := new(entity.Device)

	query := db.
		NewSelect().
		Model(d).
		Where("? = ?", bun.Ident("device."+column), val).
		Column("device.*").
		Relation("Namespace").
		ColumnExpr(string(deviceExprOnline), time.Now().Add(-2*time.Minute))

	if err := query.Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.DeviceToModel(d), nil
}

func (pg *Pg) DeviceUpdate(ctx context.Context, device *models.Device) error {
	db := pg.getConnection(ctx)

	d := entity.DeviceFromModel(device)
	d.UpdatedAt = clock.Now()
	_, err := db.NewUpdate().Model(d).WherePK().Exec(ctx)

	return fromSQLError(err)
}

func (pg *Pg) DeviceHeartbeat(ctx context.Context, ids []string, lastSeen time.Time) (int64, error) {
	db := pg.getConnection(ctx)

	r, err := db.NewUpdate().
		Model((*entity.Device)(nil)).
		Set("seen_at = ?", lastSeen).
		Set("disconnected_at = NULL").
		TableExpr("(SELECT unnest(?::varchar[]) as id) as _data", pgdialect.Array(ids)).
		Where("device.id = _data.id").
		Exec(ctx)
	if err != nil {
		return 0, fromSQLError(err)
	}

	return r.RowsAffected()
}

func (pg *Pg) DeviceDelete(ctx context.Context, device *models.Device) error {
	deletedCount, err := pg.DeviceDeleteMany(ctx, []string{device.UID})
	switch {
	case err != nil:
		return err
	case deletedCount < 1:
		return store.ErrNoDocuments
	default:
		return nil
	}
}

func (pg *Pg) DeviceDeleteMany(ctx context.Context, uids []string) (int64, error) {
	db := pg.getConnection(ctx)
	fn := pg.deviceDeleteManyFn(ctx, uids)

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

func (pg *Pg) deviceDeleteManyFn(ctx context.Context, uids []string) func(tx bun.Tx) (int64, error) {
	return func(tx bun.Tx) (int64, error) {
		r, err := tx.NewDelete().Model((*entity.Device)(nil)).Where("id IN (?)", bun.In(uids)).Exec(ctx)
		if err != nil {
			return 0, fromSQLError(err)
		}

		count, _ := r.RowsAffected()

		// if _, err := tx.NewDelete().
		// 	Model((*entity.Session)(nil)).
		// 	Where("device_uid IN (?)", bun.In(uids)).
		// 	Exec(ctx); err != nil {
		// 	return 0, fromSQLError(err)
		// }
		//
		// if _, err := tx.NewDelete().
		// 	Model((*entity.Tunnel)(nil)).
		// 	Where("device IN (?)", bun.In(uids)).
		// 	Exec(ctx); err != nil {
		// 	return 0, fromSQLError(err)
		// }

		return count, nil
	}
}

type deviceExpr string

const (
	deviceExprOnline deviceExpr = `
		CASE
		WHEN "device"."disconnected_at" IS NULL AND "device"."seen_at" > ?
		THEN true
		ELSE false
		END AS "online"`
)

// deviceExprAcepptable generates the SQL expression for the "acceptable" field
// based on the provided store.DeviceAcceptable mode.
func deviceExprAcepptable(mode store.DeviceAcceptable) string {
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

func DeviceResolverToString(resolver store.DeviceResolver) (string, error) {
	switch resolver {
	case store.DeviceUIDResolver:
		return "id", nil
	case store.DeviceHostnameResolver:
		return "name", nil
	case store.DeviceMACResolver:
		return "mac", nil
	default:
		return "", store.ErrResolverNotFound
	}
}
