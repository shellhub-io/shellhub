package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (pg *Pg) PrivateKeyCreate(ctx context.Context, privateKey *models.PrivateKey) error {
	privateKey.CreatedAt = clock.Now()

	if _, err := pg.driver.NewInsert().Model(entity.PrivateKeyFromModel(privateKey)).Exec(ctx); err != nil {
		return fromSqlError(err)
	}

	return nil
}

func (pg *Pg) PrivateKeyGet(ctx context.Context, fingerprint string) (*models.PrivateKey, error) {
	privateKey := new(entity.PrivateKey)
	if err := pg.driver.NewSelect().Model(privateKey).Where("fingerprint = ?", fingerprint).Scan(ctx); err != nil {
		return nil, fromSqlError(err)
	}

	return entity.PrivateKeyToModel(privateKey), nil
}
