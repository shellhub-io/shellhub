package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
)

func (pg *Pg) PublicKeyCreate(ctx context.Context, publicKey *models.PublicKey) (string, error) {
	publicKey.CreatedAt = clock.Now()
	e := entity.PublicKeyFromModel(publicKey)
	e.ID = uuid.Generate()

	if _, err := pg.driver.NewInsert().Model(e).Exec(ctx); err != nil {
		return "", fromSqlError(err)
	}

	return e.ID, nil // TODO: ID no model
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

func (pg *Pg) PublicKeyGet(ctx context.Context, fingerprint, tenantID string) (*models.PublicKey, error) {
	a := new(entity.PublicKey)
	if err := pg.driver.NewSelect().Model(a).Where("namespace_id = ?", tenantID).Where("fingerprint = ?", fingerprint).Scan(ctx); err != nil {
		return nil, fromSqlError(err)
	}

	return entity.PublicKeyToModel(a), nil
}

func (pg *Pg) PublicKeyUpdate(ctx context.Context, publicKey *models.PublicKey) error {
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
