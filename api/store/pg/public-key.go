package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/uptrace/bun"
)

func (pg *Pg) PublicKeyCreate(ctx context.Context, publicKey *models.PublicKey) (string, error) {
	publicKey.ID = uuid.Generate()
	publicKey.CreatedAt = clock.Now()
	publicKey.UpdatedAt = clock.Now()

	if _, err := pg.driver.NewInsert().Model(entity.PublicKeyFromModel(publicKey)).Exec(ctx); err != nil {
		return "", fromSqlError(err)
	}

	return publicKey.ID, nil
}

func (pg *Pg) PublicKeyList(ctx context.Context, opts ...store.QueryOption) ([]models.PublicKey, int, error) {
	entities := make([]entity.PublicKey, 0)

	query := pg.driver.NewSelect().Model(&entities)
	if err := applyOptions(ctx, query, opts...); err != nil {
		return nil, 0, fromSqlError(err)
	}

	count, err := query.ScanAndCount(ctx)
	if err != nil {
		return nil, 0, fromSqlError(err)
	}

	publicKeys := make([]models.PublicKey, len(entities))
	for i, e := range entities {
		publicKeys[i] = *entity.PublicKeyToModel(&e)
	}

	return publicKeys, count, nil
}

func (pg *Pg) PublicKeyGet(ctx context.Context, ident store.PublicKeyIdent, val string, tenantID string) (*models.PublicKey, error) {
	a := new(entity.PublicKey)
	if err := pg.driver.NewSelect().Model(a).Where("? = ?", bun.Ident(ident), val).Scan(ctx); err != nil {
		return nil, fromSqlError(err)
	}

	return entity.PublicKeyToModel(a), nil
}

func (pg *Pg) PublicKeySave(ctx context.Context, publicKey *models.PublicKey) error {
	a := entity.PublicKeyFromModel(publicKey)
	a.UpdatedAt = clock.Now()
	_, err := pg.driver.NewUpdate().Model(a).WherePK().Exec(ctx)

	return fromSqlError(err)
}

func (pg *Pg) PublicKeyDelete(ctx context.Context, publicKey *models.PublicKey) error {
	a := entity.PublicKeyFromModel(publicKey)
	_, err := pg.driver.NewDelete().Model(a).WherePK().Exec(ctx)

	return fromSqlError(err)
}
