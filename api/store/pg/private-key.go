package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (pg *Pg) PrivateKeyCreate(ctx context.Context, privateKey *models.PrivateKey) error {
	db := pg.GetConnection(ctx)

	privateKey.CreatedAt = clock.Now()

	if _, err := db.NewInsert().Model(entity.PrivateKeyFromModel(privateKey)).Exec(ctx); err != nil {
		return fromSQLError(err)
	}

	return nil
}

func (pg *Pg) PrivateKeyGet(ctx context.Context, fingerprint string) (*models.PrivateKey, error) {
	db := pg.GetConnection(ctx)

	privateKey := new(entity.PrivateKey)
	if err := db.NewSelect().Model(privateKey).Where("fingerprint = ?", fingerprint).Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.PrivateKeyToModel(privateKey), nil
}
