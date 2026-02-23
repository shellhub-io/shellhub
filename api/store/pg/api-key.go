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
	db := pg.GetConnection(ctx)

	apiKey.CreatedAt = clock.Now()
	apiKey.UpdatedAt = clock.Now()
	if _, err := db.NewInsert().Model(entity.APIKeyFromModel(apiKey)).Exec(ctx); err != nil {
		return "", fromSQLError(err)
	}

	return apiKey.ID, nil
}

func (pg *Pg) APIKeyConflicts(ctx context.Context, tenantID string, target *models.APIKeyConflicts) ([]string, bool, error) {
	db := pg.GetConnection(ctx)

	if target.ID == "" && target.Name == "" {
		return []string{}, false, nil
	}

	apiKeys := make([]entity.APIKey, 0)
	query := db.NewSelect().
		Model(&apiKeys).
		Column("key_digest", "name").
		Where("namespace_id = ?", tenantID)

	// Add OR conditions for ID and Name within the same tenant
	if target.ID != "" && target.Name != "" {
		query = query.Where("key_digest = ? OR name = ?", target.ID, target.Name)
	} else if target.ID != "" {
		query = query.Where("key_digest = ?", target.ID)
	} else if target.Name != "" {
		query = query.Where("name = ?", target.Name)
	}

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
	db := pg.GetConnection(ctx)

	entities := make([]entity.APIKey, 0)

	query := db.NewSelect().Model(&entities)
	var err error
	query, err = applyOptions(ctx, query, opts...)
	if err != nil {
		return nil, 0, err
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
	db := pg.GetConnection(ctx)

	column, err := APIKeyResolverToString(resolver)
	if err != nil {
		return nil, err
	}

	apKey := new(entity.APIKey)
	query := db.NewSelect().Model(apKey).Where("? = ?", bun.Ident(column), val)
	query, err = applyOptions(ctx, query, opts...)
	if err != nil {
		return nil, err
	}

	if err = query.Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.APIKeyToModel(apKey), nil
}

func (pg *Pg) APIKeyUpdate(ctx context.Context, apiKey *models.APIKey) error {
	db := pg.GetConnection(ctx)

	a := entity.APIKeyFromModel(apiKey)
	a.UpdatedAt = clock.Now()
	r, err := db.NewUpdate().Model(a).OmitZero().WherePK().Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

func (pg *Pg) APIKeyDelete(ctx context.Context, apiKey *models.APIKey) error {
	db := pg.GetConnection(ctx)

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
		return "key_digest", nil
	case store.APIKeyNameResolver:
		return "name", nil
	default:
		return "", store.ErrResolverNotFound
	}
}
