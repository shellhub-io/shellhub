package pg

import (
	"context"
	"database/sql"
	"errors"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/server/api/store/pg/entity"
	"github.com/uptrace/bun"
)

// SystemGet implements [store.SystemStore].
func (pg *Pg) SystemGet(ctx context.Context) (*models.System, error) {
	db := pg.GetConnection(ctx)

	system := new(entity.System)
	if err := selectSystem(db, system).Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			system := &models.System{
				Setup: false,
				Authentication: &models.SystemAuthentication{
					Local: &models.SystemAuthenticationLocal{
						Enabled: true,
					},
				},
				Telemetry: &models.SystemTelemetry{Enabled: true},
			}

			return system, nil
		}

		return nil, err
	}

	return entity.SystemToModel(system), nil
}

func selectSystem(db bun.IDB, system *entity.System) *bun.SelectQuery {
	return db.NewSelect().Model(system).OrderExpr("setup DESC, instance_tenant_id IS NULL, id").Limit(1)
}

var systemColumnsSetElsewhere = []string{"instance_id", "telemetry_enabled"}

// SystemSet implements [store.SystemStore].
func (pg *Pg) SystemSet(ctx context.Context, system *models.System) error {
	db := pg.GetConnection(ctx)

	existingSystem := new(entity.System)
	err := selectSystem(db, existingSystem).Scan(ctx)

	systemEntity := entity.SystemFromModel(system)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		if systemEntity.ID == "" {
			systemEntity.ID = uuid.Generate()
		}
		_, err = db.NewInsert().Model(systemEntity).ExcludeColumn(systemColumnsSetElsewhere...).Exec(ctx)
	case err == nil:
		systemEntity.ID = existingSystem.ID
		_, err = db.NewUpdate().Model(systemEntity).ExcludeColumn(systemColumnsSetElsewhere...).WherePK().Exec(ctx)
	}

	return err
}
