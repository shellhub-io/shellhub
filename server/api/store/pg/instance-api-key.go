package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/pg/entity"
	"github.com/uptrace/bun"
)

// InstanceAPIKeyCreate implements [store.InstanceAPIKeyStore].
func (pg *Pg) InstanceAPIKeyCreate(ctx context.Context, apiKey *models.InstanceAPIKey) (string, error) {
	db := pg.GetConnection(ctx)

	apiKey.CreatedAt = clock.Now()
	apiKey.UpdatedAt = clock.Now()

	if _, err := db.NewInsert().Model(entity.InstanceAPIKeyFromModel(apiKey)).Exec(ctx); err != nil {
		return "", fromSQLError(err)
	}

	return apiKey.ID, nil
}

// InstanceAPIKeyResolve implements [store.InstanceAPIKeyStore].
func (pg *Pg) InstanceAPIKeyResolve(ctx context.Context, resolver store.InstanceAPIKeyResolver, value string, opts ...store.QueryOption) (*models.InstanceAPIKey, error) {
	db := pg.GetConnection(ctx)

	column, err := InstanceAPIKeyResolverToString(resolver)
	if err != nil {
		return nil, err
	}

	apiKey := new(entity.InstanceAPIKey)

	query := db.NewSelect().Model(apiKey).Where("? = ?", bun.Ident(column), value)
	if query, err = applyOptions(ctx, query, opts...); err != nil {
		return nil, err
	}

	if err := query.Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.InstanceAPIKeyToModel(apiKey), nil
}

// InstanceAPIKeyList implements [store.InstanceAPIKeyStore].
func (pg *Pg) InstanceAPIKeyList(ctx context.Context, opts ...store.QueryOption) ([]models.InstanceAPIKey, int, error) {
	db := pg.GetConnection(ctx)

	entities := make([]entity.InstanceAPIKey, 0)

	query, err := applyOptions(ctx, db.NewSelect().Model(&entities), opts...)
	if err != nil {
		return nil, 0, err
	}

	count, err := query.ScanAndCount(ctx)
	if err != nil {
		return nil, 0, fromSQLError(err)
	}

	apiKeys := make([]models.InstanceAPIKey, len(entities))
	for i, e := range entities {
		apiKeys[i] = *entity.InstanceAPIKeyToModel(&e)
	}

	return apiKeys, count, nil
}

// InstanceAPIKeyDelete implements [store.InstanceAPIKeyStore].
func (pg *Pg) InstanceAPIKeyDelete(ctx context.Context, name string) error {
	db := pg.GetConnection(ctx)

	r, err := db.NewDelete().Model((*entity.InstanceAPIKey)(nil)).Where("name = ?", name).Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

// InstanceAPIKeyResolverToString maps a resolver onto the column it looks up.
func InstanceAPIKeyResolverToString(resolver store.InstanceAPIKeyResolver) (string, error) {
	switch resolver {
	case store.InstanceAPIKeyIDResolver:
		return "key_digest", nil
	case store.InstanceAPIKeyNameResolver:
		return "name", nil
	default:
		return "", store.ErrResolverNotFound
	}
}
