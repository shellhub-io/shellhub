package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

func (pg *Pg) PublicKeyCreate(ctx context.Context, publicKey *models.PublicKey) (string, error) {
	db := pg.getConnection(ctx)

	publicKey.CreatedAt = clock.Now()
	e := entity.PublicKeyFromModel(publicKey)

	if _, err := db.NewInsert().Model(e).Exec(ctx); err != nil {
		return "", fromSQLError(err)
	}

	return e.Fingerprint, nil
}

func (pg *Pg) PublicKeyList(ctx context.Context, opts ...store.QueryOption) ([]models.PublicKey, int, error) {
	db := pg.getConnection(ctx)

	entities := make([]entity.PublicKey, 0)

	query := db.NewSelect().Model(&entities).Relation("Tags")
	var err error
	query, err = applyOptions(ctx, query, opts...)
	if err != nil {
		return nil, 0, err
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

	// Filter by both fingerprint and namespace_id to match MongoDB behavior
	r, err := db.NewUpdate().
		Model(a).
		Where("fingerprint = ?", publicKey.Fingerprint).
		Where("namespace_id = ?", publicKey.TenantID).
		Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

func (pg *Pg) PublicKeyResolve(ctx context.Context, resolver store.PublicKeyResolver, value string, opts ...store.QueryOption) (*models.PublicKey, error) {
	db := pg.getConnection(ctx)

	column, err := PublicKeyResolverToString(resolver)
	if err != nil {
		return nil, err
	}

	a := new(entity.PublicKey)
	query := db.NewSelect().Model(a).
		Relation("Tags").
		Where("? = ?", bun.Ident(column), value)
	query, err = applyOptions(ctx, query, opts...)
	if err != nil {
		return nil, err
	}

	if err = query.Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.PublicKeyToModel(a), nil
}

func (pg *Pg) PublicKeyDelete(ctx context.Context, publicKey *models.PublicKey) error {
	db := pg.getConnection(ctx)

	a := entity.PublicKeyFromModel(publicKey)

	// Filter by both fingerprint and namespace_id to match MongoDB behavior
	r, err := db.NewDelete().
		Model(a).
		Where("fingerprint = ?", publicKey.Fingerprint).
		Where("namespace_id = ?", publicKey.TenantID).
		Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

func PublicKeyResolverToString(resolver store.PublicKeyResolver) (string, error) {
	switch resolver {
	case store.PublicKeyFingerprintResolver:
		return "fingerprint", nil
	default:
		return "", store.ErrResolverNotFound
	}
}
