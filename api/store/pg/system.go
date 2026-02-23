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
	db := pg.GetConnection(ctx)

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
	db := pg.GetConnection(ctx)

	// Get existing system (should be only one)
	existingSystem := new(entity.System)
	err := db.NewSelect().Model(existingSystem).Limit(1).Scan(ctx)

	systemEntity := entity.SystemFromModel(system)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		// No system exists, create new one
		if systemEntity.ID == "" {
			systemEntity.ID = uuid.Generate()
		}
		_, err = db.NewInsert().Model(systemEntity).Exec(ctx)
	case err == nil:
		// System exists, update it (use existing ID)
		systemEntity.ID = existingSystem.ID
		_, err = db.NewUpdate().Model(systemEntity).WherePK().Exec(ctx)
	}

	return err
}
