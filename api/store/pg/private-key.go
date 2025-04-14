package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store/pg/internal/entity"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (pg *pg) PrivateKeyCreate(ctx context.Context, key *models.PrivateKey) error {
	key.CreatedAt = clock.Now()
	key.UpdatedAt = clock.Now()

	if _, err := pg.driver.NewInsert().Model(entity.PrivateKeyFromModel(key)).Exec(ctx); err != nil {
		return fromSqlError(err)
	}

	return nil
}

func (pg *pg) PrivateKeyGet(ctx context.Context, fingerprint string) (*models.PrivateKey, error) {
	k := new(entity.PrivateKey)
	if err := pg.driver.NewSelect().Model(k).Where("fingerprint = ?", fingerprint).Scan(ctx); err != nil {
		return nil, fromSqlError(err)
	}

	return entity.PrivateKeyToModel(k), nil
}
