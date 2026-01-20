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

	if target.ID == "" && target.Name == "" {
		return nil, false, nil
	}

	apiKeys := make([]entity.APIKey, 0)
	query := db.NewSelect().
		Model(&apiKeys).
		Column("key_digest", "name").
		Where("namespace_id = ?", tenantID).
		WhereGroup(" OR ", func(q *bun.SelectQuery) *bun.SelectQuery {
			if target.ID != "" {
				q = q.Where("key_digest = ?", target.ID)
			}

			if target.Name != "" {
				q = q.Where("name = ?", target.Name)
			}

			return q
		})

	if err := query.Scan(ctx); err != nil {
		return nil, false, fromSQLError(err)
	}

	seen := make(map[string]bool)
	for _, apiKey := range apiKeys {
		if target.ID != "" && apiKey.KeyDigest == target.ID {
			seen["id"] = true
		}

		if target.Name != "" && apiKey.Name == target.Name {
			seen["name"] = true
		}
	}

	conflicts := make([]string, 0, len(seen))
	for field := range seen {
		conflicts = append(conflicts, field)
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
	r, err := db.NewUpdate().Model(a).WherePK().Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

func (pg *Pg) APIKeyDelete(ctx context.Context, apiKey *models.APIKey) error {
	db := pg.getConnection(ctx)

	a := entity.APIKeyFromModel(apiKey)
	r, err := db.NewDelete().Model(a).WherePK().Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return nil
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
