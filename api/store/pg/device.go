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
	db := pg.GetConnection(ctx)

	device.CreatedAt = clock.Now()

	e := entity.DeviceFromModel(device)
	if _, err := db.NewInsert().Model(e).Exec(ctx); err != nil {
		return "", fromSQLError(err)
	}

	return e.ID, nil
}

func (pg *Pg) DeviceConflicts(ctx context.Context, target *models.DeviceConflicts) ([]string, bool, error) {
	db := pg.GetConnection(ctx)

	if target.Name == "" {
		return []string{}, false, nil
	}

	devices := make([]entity.Device, 0)
	query := db.NewSelect().
		Model(&devices).
		Column("name").
		Where("status != ?", models.DeviceStatusRemoved).
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
	for _, device := range devices {
		if target.Name != "" && device.Name == target.Name {
			seen["name"] = true
		}
	}

	conflicts := make([]string, 0, len(seen))
	for field := range seen {
		conflicts = append(conflicts, field)
	}

	return conflicts, len(conflicts) > 0, nil
}

func (pg *Pg) DeviceList(ctx context.Context, acceptable store.DeviceAcceptable, opts ...store.QueryOption) ([]models.Device, int, error) {
	db := pg.GetConnection(ctx)

	entities := make([]entity.Device, 0)

	onlineExpr, onlineThreshold := deviceExprOnline(time.Now().Add(-2 * time.Minute))
	query := db.
		NewSelect().
		Model(&entities).
		Column("device.*").
		Relation("Namespace").
		Relation("Tags").
		ColumnExpr(onlineExpr, onlineThreshold).
		ColumnExpr(deviceExprAcceptable(acceptable))

	var err error
	query, err = applyOptions(ctx, query, opts...)
	if err != nil {
		return nil, 0, err
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
	db := pg.GetConnection(ctx)

	column, err := DeviceResolverToString(resolver)
	if err != nil {
		return nil, err
	}

	d := new(entity.Device)

	onlineExpr, onlineThreshold := deviceExprOnline(time.Now().Add(-2 * time.Minute))
	query := db.
		NewSelect().
		Model(d).
		Where("? = ?", bun.Ident("device."+column), val).
		Column("device.*").
		Relation("Namespace").
		Relation("Tags").
		ColumnExpr(onlineExpr, onlineThreshold)

	query, err = applyOptions(ctx, query, opts...)
	if err != nil {
		return nil, err
	}

	if err = query.Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.DeviceToModel(d), nil
}

func (pg *Pg) DeviceUpdate(ctx context.Context, device *models.Device) error {
	db := pg.GetConnection(ctx)

	d := entity.DeviceFromModel(device)
	d.UpdatedAt = clock.Now()

	r, err := db.NewUpdate().Model(d).Where("id = ?", d.ID).Where("namespace_id = ?", d.NamespaceID).Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

func (pg *Pg) DeviceHeartbeat(ctx context.Context, ids []string, lastSeen time.Time) (int64, error) {
	db := pg.GetConnection(ctx)

	unnestExpr, unnestIDs := deviceExprUnnestIDs(ids)
	r, err := db.NewUpdate().
		Model((*entity.Device)(nil)).
		Set("last_seen = ?", lastSeen).
		Set("disconnected_at = NULL").
		TableExpr(unnestExpr, unnestIDs).
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
	db := pg.GetConnection(ctx)
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
		r, err := tx.NewDelete().Model((*entity.Device)(nil)).Where("id IN (?)", bun.List(uids)).Exec(ctx)
		if err != nil {
			return 0, fromSQLError(err)
		}

		count, _ := r.RowsAffected()

		if _, err := tx.NewDelete().
			Model((*entity.Session)(nil)).
			Where("device_id IN (?)", bun.List(uids)).
			Exec(ctx); err != nil {
			return 0, fromSQLError(err)
		}

		return count, nil
	}
}

// deviceExprOnline returns the SQL expression for the "online" field.
func deviceExprOnline(threshold time.Time) (string, time.Time) {
	return `CASE
		WHEN "device"."disconnected_at" IS NULL AND "device"."last_seen" > ?
		THEN true
		ELSE false
		END AS "online"`, threshold
}

// deviceExprAcceptable returns the SQL expression for the "acceptable" field
// based on the provided store.DeviceAcceptable mode.
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

// deviceExprUnnestIDs returns a table expression for batch operations with IDs.
func deviceExprUnnestIDs(ids []string) (string, any) {
	return "(SELECT unnest(?::varchar[]) as id) as _data", pgdialect.Array(ids)
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
