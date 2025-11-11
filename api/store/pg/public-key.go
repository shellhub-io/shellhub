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
	db := pg.getConnection(ctx)

	publicKey.CreatedAt = clock.Now()
	e := entity.PublicKeyFromModel(publicKey)
	e.ID = uuid.Generate()

	if _, err := db.NewInsert().Model(e).Exec(ctx); err != nil {
		return "", fromSQLError(err)
	}

	return e.ID, nil // TODO: ID no model
}

func (pg *Pg) PublicKeyList(ctx context.Context, opts ...store.QueryOption) ([]models.PublicKey, int, error) {
	db := pg.getConnection(ctx)

	entities := make([]entity.PublicKey, 0)

	query := db.NewSelect().Model(&entities)
	if err := applyOptions(ctx, query, opts...); err != nil {
		return nil, 0, fromSQLError(err)
	}

	count, err := query.ScanAndCount(ctx)
	if err != nil {
		return nil, 0, fromSQLError(err)
	}

	publicKeys := make([]models.PublicKey, len(entities))
	for i, e := range entities {
		publicKeys[i] = *entity.PublicKeyToModel(&e)
	}

	return publicKeys, count, nil
}

func (pg *Pg) PublicKeyUpdate(ctx context.Context, publicKey *models.PublicKey) error {
	db := pg.getConnection(ctx)

	a := entity.PublicKeyFromModel(publicKey)
	a.UpdatedAt = clock.Now()
	_, err := db.NewUpdate().Model(a).WherePK().Exec(ctx)

	return fromSQLError(err)
}

func (pg *Pg) PublicKeyResolve(ctx context.Context, resolver store.PublicKeyResolver, value string, opts ...store.QueryOption) (*models.PublicKey, error) {
	db := pg.getConnection(ctx)

	column, err := PublicKeyResolverToString(resolver)
	if err != nil {
		return nil, err
	}

	a := new(entity.PublicKey)
	query := db.NewSelect().Model(a).Where("? = ?", bun.Ident(column), value)
	if err := applyOptions(ctx, query, opts...); err != nil {
		return nil, fromSQLError(err)
	}

	if err := query.Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.PublicKeyToModel(a), nil
}

func (pg *Pg) PublicKeyDelete(ctx context.Context, publicKey *models.PublicKey) error {
	db := pg.getConnection(ctx)

	a := entity.PublicKeyFromModel(publicKey)
	_, err := db.NewDelete().Model(a).WherePK().Exec(ctx)

	return fromSQLError(err)
}

func PublicKeyResolverToString(resolver store.PublicKeyResolver) (string, error) {
	switch resolver {
	case store.PublicKeyFingerprintResolver:
		return "fingerprint", nil
	default:
		return "", store.ErrResolverNotFound
	}
}
