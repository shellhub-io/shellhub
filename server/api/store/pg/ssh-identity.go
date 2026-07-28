package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/pg/entity"
	"github.com/uptrace/bun"
)

func (pg *Pg) SSHIdentityCreate(ctx context.Context, identity *models.SSHIdentity) (string, error) {
	db := pg.GetConnection(ctx)

	if identity.ID == "" {
		identity.ID = uuid.Generate()
	}

	if identity.CreatedAt.IsZero() {
		identity.CreatedAt = clock.Now()
	}

	e := entity.SSHIdentityFromModel(identity)

	if _, err := db.NewInsert().Model(e).Exec(ctx); err != nil {
		return "", fromSQLError(err)
	}

	return e.ID, nil
}

func (pg *Pg) SSHIdentityList(ctx context.Context, sc scope.Scope, opts ...store.QueryOption) ([]models.SSHIdentity, int, error) {
	db := pg.GetConnection(ctx)

	entities := make([]entity.SSHIdentity, 0)

	query := db.NewSelect().Model(&entities).Relation("User").Order("created_at ASC")

	query, err := applyScopedOptions(ctx, query, sc, opts...)
	if err != nil {
		return nil, 0, err
	}

	count, err := query.ScanAndCount(ctx)
	if err != nil {
		return nil, 0, fromSQLError(err)
	}

	identities := make([]models.SSHIdentity, len(entities))
	for i, e := range entities {
		identities[i] = *entity.SSHIdentityToModel(&e)
	}

	return identities, count, nil
}

func (pg *Pg) SSHIdentityResolve(ctx context.Context, sc scope.Scope, resolver store.SSHIdentityResolver, value string, opts ...store.QueryOption) (*models.SSHIdentity, error) {
	db := pg.GetConnection(ctx)

	column, err := sshIdentityResolverToString(resolver)
	if err != nil {
		return nil, err
	}

	e := new(entity.SSHIdentity)
	query := db.NewSelect().Model(e).
		Where("? = ?", bun.Ident(column), value)

	query, err = applyScopedOptions(ctx, query, sc, opts...)
	if err != nil {
		return nil, err
	}

	if err := query.Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.SSHIdentityToModel(e), nil
}

func (pg *Pg) SSHIdentityUpdate(ctx context.Context, identity *models.SSHIdentity) error {
	db := pg.GetConnection(ctx)

	r, err := db.NewUpdate().
		Model((*entity.SSHIdentity)(nil)).
		Set("name = ?", identity.Name).
		Where("id = ?", identity.ID).
		Where("namespace_id = ?", identity.TenantID).
		Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

func (pg *Pg) SSHIdentityDelete(ctx context.Context, identity *models.SSHIdentity) error {
	db := pg.GetConnection(ctx)

	r, err := db.NewDelete().
		Model((*entity.SSHIdentity)(nil)).
		Where("id = ?", identity.ID).
		Where("namespace_id = ?", identity.TenantID).
		Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

func (pg *Pg) SSHIdentityTouchLastUsed(ctx context.Context, tenantID, fingerprint string) error {
	db := pg.GetConnection(ctx)

	now := clock.Now()

	if _, err := db.NewUpdate().
		Model((*entity.SSHIdentity)(nil)).
		Set("last_used_at = ?", now).
		Where("namespace_id = ?", tenantID).
		Where("fingerprint = ?", fingerprint).
		Exec(ctx); err != nil {
		return fromSQLError(err)
	}

	return nil
}

func (pg *Pg) SSHIdentityTouchReauth(ctx context.Context, tenantID, fingerprint string) error {
	db := pg.GetConnection(ctx)

	now := clock.Now()

	if _, err := db.NewUpdate().
		Model((*entity.SSHIdentity)(nil)).
		Set("last_reauth_at = ?", now).
		Where("namespace_id = ?", tenantID).
		Where("fingerprint = ?", fingerprint).
		Exec(ctx); err != nil {
		return fromSQLError(err)
	}

	return nil
}

func (pg *Pg) SSHIdentityConsume(ctx context.Context, tenantID, fingerprint string) (bool, error) {
	db := pg.GetConnection(ctx)

	// Atomic burn: only the first caller flips consumed_at from NULL, so the
	// affected-row count decides the race between concurrent single-use sessions.
	res, err := db.NewUpdate().
		Model((*entity.SSHIdentity)(nil)).
		Set("consumed_at = ?", clock.Now()).
		Where("namespace_id = ?", tenantID).
		Where("fingerprint = ?", fingerprint).
		Where("consumed_at IS NULL").
		Exec(ctx)
	if err != nil {
		return false, fromSQLError(err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, fromSQLError(err)
	}

	return affected == 1, nil
}

func sshIdentityResolverToString(resolver store.SSHIdentityResolver) (string, error) {
	switch resolver {
	case store.SSHIdentityIDResolver:
		return "id", nil
	case store.SSHIdentityFingerprintResolver:
		return "fingerprint", nil
	default:
		return "", store.ErrResolverNotFound
	}
}
