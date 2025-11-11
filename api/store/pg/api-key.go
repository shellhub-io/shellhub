package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

func (pg *Pg) APIKeyCreate(ctx context.Context, apiKey *models.APIKey) (string, error) {
	db := pg.getConnection(ctx)

	apiKey.CreatedAt = clock.Now()
	apiKey.UpdatedAt = clock.Now()
	if _, err := db.NewInsert().Model(entity.APIKeyFromModel(apiKey)).Exec(ctx); err != nil {
		return "", fromSQLError(err)
	}

	return apiKey.ID, nil
}

func (pg *Pg) APIKeyConflicts(ctx context.Context, tenantID string, target *models.APIKeyConflicts) ([]string, bool, error) {
	db := pg.getConnection(ctx)

	apiKeys := make([]map[string]any, 0)
	if err := db.NewSelect().Model((*entity.Namespace)(nil)).Column("name").Where("name = ?", target.Name).Scan(ctx, &apiKeys); err != nil {
		return nil, false, fromSQLError(err)
	}

	conflicts := make([]string, 0)
	for _, apiKey := range apiKeys {
		if apiKey["name"] == target.Name {
			conflicts = append(conflicts, "name")
		}
	}

	return conflicts, len(conflicts) > 0, nil
}

func (pg *Pg) APIKeyList(ctx context.Context, opts ...store.QueryOption) ([]models.APIKey, int, error) {
	db := pg.getConnection(ctx)

	entities := make([]entity.APIKey, 0)

	query := db.NewSelect().Model(&entities)
	if err := applyOptions(ctx, query, opts...); err != nil {
		return nil, 0, fromSQLError(err)
	}

	count, err := query.ScanAndCount(ctx)
	if err != nil {
		return nil, 0, fromSQLError(err)
	}

	apiKeys := make([]models.APIKey, len(entities))
	for i, e := range entities {
		apiKeys[i] = *entity.APIKeyToModel(&e)
	}

	return apiKeys, count, nil
}

func (pg *Pg) APIKeyResolve(ctx context.Context, resolver store.APIKeyResolver, val string, opts ...store.QueryOption) (*models.APIKey, error) {
	db := pg.getConnection(ctx)

	column, err := APIKeyResolverToString(resolver)
	if err != nil {
		return nil, err
	}

	apKey := new(entity.APIKey)
	query := db.NewSelect().Model(apKey).Where("? = ?", bun.Ident(column), val)
	if err := applyOptions(ctx, query, opts...); err != nil {
		return nil, fromSQLError(err)
	}

	if err := query.Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.APIKeyToModel(apKey), nil
}

func (pg *Pg) APIKeyUpdate(ctx context.Context, apiKey *models.APIKey) error {
	db := pg.getConnection(ctx)

	a := entity.APIKeyFromModel(apiKey)
	a.UpdatedAt = clock.Now()
	_, err := db.NewUpdate().Model(a).WherePK().Exec(ctx)

	return fromSQLError(err)
}

func (pg *Pg) APIKeyDelete(ctx context.Context, apiKey *models.APIKey) error {
	db := pg.getConnection(ctx)

	a := entity.APIKeyFromModel(apiKey)
	_, err := db.NewDelete().Model(a).WherePK().Exec(ctx)

	return fromSQLError(err)
}

func APIKeyResolverToString(resolver store.APIKeyResolver) (string, error) {
	switch resolver {
	case store.APIKeyIDResolver:
		return "id", nil
	case store.APIKeyNameResolver:
		return "name", nil
	default:
		return "", store.ErrResolverNotFound
	}
}
