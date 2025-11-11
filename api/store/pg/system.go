package pg

import (
	"context"
	"database/sql"
	"errors"

	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
)

func (pg *Pg) SystemGet(ctx context.Context) (*models.System, error) {
	db := pg.getConnection(ctx)

	system := new(entity.System)
	if err := db.NewSelect().Model(system).Limit(1).Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			system := &models.System{
				Setup: false,
				Authentication: &models.SystemAuthentication{
					Local: &models.SystemAuthenticationLocal{
						Enabled: true,
					},
					SAML: &models.SystemAuthenticationSAML{
						Enabled: false,
						Idp:     &models.SystemIdpSAML{Binding: &models.SystemAuthenticationBinding{}},
						Sp:      &models.SystemSpSAML{},
					},
				},
			}

			return system, nil
		}

		return nil, err
	}

	return entity.SystemToModel(system), nil
}

func (pg *Pg) SystemSet(ctx context.Context, system *models.System) error {
	systemEntity := entity.SystemFromModel(system)
	if systemEntity.ID == "" {
		systemEntity.ID = uuid.Generate()
	}

	db := pg.getConnection(ctx)
	exists, err := db.NewSelect().Model((*entity.System)(nil)).Where("id = ?", systemEntity.ID).Exists(ctx)
	switch {
	case err == nil && !exists:
		_, err = pg.driver.NewInsert().Model(systemEntity).Exec(ctx)
	case err == nil && exists:
		_, err = pg.driver.NewUpdate().Model(systemEntity).Where("id = ?", systemEntity.ID).Exec(ctx)
	}

	return err
}
