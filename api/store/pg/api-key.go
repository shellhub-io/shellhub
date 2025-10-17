package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (pg *Pg) APIKeyCreate(ctx context.Context, apiKey *models.APIKey) (string, error) {
	apiKey.CreatedAt = clock.Now()
	apiKey.UpdatedAt = clock.Now()

	if _, err := pg.driver.NewInsert().Model(entity.APIKeyFromModel(apiKey)).Exec(ctx); err != nil {
		return "", fromSqlError(err)
	}

	return apiKey.ID, nil
}

func (pg *Pg) APIKeyConflicts(ctx context.Context, tenantID string, target *models.APIKeyConflicts) ([]string, bool, error) {
	apiKeys := make([]map[string]any, 0)
	if err := pg.driver.NewSelect().Model((*entity.Namespace)(nil)).Column("name").Where("name = ?", target.Name).Scan(ctx, &apiKeys); err != nil {
		return nil, false, fromSqlError(err)
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
	entities := make([]entity.APIKey, 0)

	query := pg.driver.NewSelect().Model(&entities)
	if err := applyOptions(ctx, query, opts...); err != nil {
		return nil, 0, fromSqlError(err)
	}

	count, err := query.ScanAndCount(ctx)
	if err != nil {
		return nil, 0, fromSqlError(err)
	}

	apiKeys := make([]models.APIKey, len(entities))
	for i, e := range entities {
		apiKeys[i] = *entity.APIKeyToModel(&e)
	}

	return apiKeys, count, nil
}

func (pg *Pg) APIKeyGet(ctx context.Context, id string) (*models.APIKey, error) {
	a := new(entity.APIKey)
	if err := pg.driver.NewSelect().Model(a).Where("id = ?", id).Scan(ctx); err != nil {
		return nil, fromSqlError(err)
	}

	return entity.APIKeyToModel(a), nil
}

func (pg *Pg) APIKeyGetByName(ctx context.Context, tenantID string, name string) (*models.APIKey, error) {
	a := new(entity.APIKey)
	if err := pg.driver.NewSelect().Model(a).Where("namespace_id = ?", tenantID).Where("name = ?", name).Scan(ctx); err != nil {
		return nil, fromSqlError(err)
	}

	return entity.APIKeyToModel(a), nil
}

func (pg *Pg) APIKeyUpdate(ctx context.Context, apiKey *models.APIKey) error {
	a := entity.APIKeyFromModel(apiKey)
	a.UpdatedAt = clock.Now()
	_, err := pg.driver.NewUpdate().Model(a).WherePK().Exec(ctx)

	return fromSqlError(err)
}

func (pg *Pg) APIKeyDelete(ctx context.Context, apiKey *models.APIKey) error {
	a := entity.APIKeyFromModel(apiKey)
	_, err := pg.driver.NewDelete().Model(a).WherePK().Exec(ctx)

	return fromSqlError(err)
}
