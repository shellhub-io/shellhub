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

func (pg *Pg) AccessPolicyCreate(ctx context.Context, accessPolicy *models.AccessPolicy) (string, error) {
	db := pg.GetConnection(ctx)

	now := clock.Now()
	accessPolicy.CreatedAt = now
	accessPolicy.UpdatedAt = now

	if accessPolicy.ID == "" {
		accessPolicy.ID = uuid.Generate()
	}

	e := entity.AccessPolicyFromModel(accessPolicy)

	if _, err := db.NewInsert().Model(e).Exec(ctx); err != nil {
		return "", fromSQLError(err)
	}

	for _, tag := range e.Tags {
		apTag := entity.NewAccessPolicyTag(tag.ID, e.ID)
		apTag.CreatedAt = now

		if _, err := db.NewInsert().
			Model(apTag).
			On("CONFLICT (access_policy_id, tag_id) DO NOTHING").
			Exec(ctx); err != nil {
			return "", fromSQLError(err)
		}
	}

	return e.ID, nil
}

func (pg *Pg) AccessPolicyList(ctx context.Context, opts ...store.QueryOption) ([]models.AccessPolicy, int, error) {
	db := pg.GetConnection(ctx)

	entities := make([]entity.AccessPolicy, 0)

	query := db.NewSelect().Model(&entities).Relation("Tags")

	query, err := applyOptions(ctx, query, opts...)
	if err != nil {
		return nil, 0, err
	}

	count, err := query.ScanAndCount(ctx)
	if err != nil {
		return nil, 0, fromSQLError(err)
	}

	accessPolicies := make([]models.AccessPolicy, len(entities))
	for i, e := range entities {
		accessPolicies[i] = *entity.AccessPolicyToModel(&e)
	}

	return accessPolicies, count, nil
}

func (pg *Pg) AccessPolicyResolve(ctx context.Context, resolver store.AccessPolicyResolver, value string, opts ...store.QueryOption) (*models.AccessPolicy, error) {
	db := pg.GetConnection(ctx)

	column, err := AccessPolicyResolverToString(resolver)
	if err != nil {
		return nil, err
	}

	e := new(entity.AccessPolicy)
	query := db.NewSelect().Model(e).
		Relation("Tags").
		Where("? = ?", bun.Ident(column), value)

	query, err = applyOptions(ctx, query, opts...)
	if err != nil {
		return nil, err
	}

	if err := query.Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.AccessPolicyToModel(e), nil
}

func (pg *Pg) AccessPolicyUpdate(ctx context.Context, accessPolicy *models.AccessPolicy) error {
	return pg.WithTransaction(ctx, func(ctx context.Context) error {
		db := pg.GetConnection(ctx)

		e := entity.AccessPolicyFromModel(accessPolicy)
		e.UpdatedAt = clock.Now()

		r, err := db.NewUpdate().
			Model(e).
			Column("name", "subject_type", "subject_value", "filter_hostname", "logins", "source_ip", "action", "require_reauth", "reauth_period", "updated_at").
			Where("id = ?", accessPolicy.ID).
			Where("namespace_id = ?", accessPolicy.TenantID).
			Exec(ctx)
		if err != nil {
			return fromSQLError(err)
		}

		if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
			return store.ErrNoDocuments
		}

		// Sync the many-to-many tag relationships: drop the existing junction
		// entries and re-insert the current set so removed tags don't linger.
		if _, err := db.NewDelete().
			Model((*entity.AccessPolicyTag)(nil)).
			Where("access_policy_id = ?", e.ID).
			Exec(ctx); err != nil {
			return fromSQLError(err)
		}

		for _, tag := range e.Tags {
			apTag := entity.NewAccessPolicyTag(tag.ID, e.ID)
			apTag.CreatedAt = e.UpdatedAt

			if _, err := db.NewInsert().
				Model(apTag).
				On("CONFLICT (access_policy_id, tag_id) DO NOTHING").
				Exec(ctx); err != nil {
				return fromSQLError(err)
			}
		}

		return nil
	})
}

func (pg *Pg) AccessPolicyDelete(ctx context.Context, accessPolicy *models.AccessPolicy) error {
	db := pg.GetConnection(ctx)

	e := entity.AccessPolicyFromModel(accessPolicy)

	r, err := db.NewDelete().
		Model(e).
		Where("id = ?", accessPolicy.ID).
		Where("namespace_id = ?", accessPolicy.TenantID).
		Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

func AccessPolicyResolverToString(resolver store.AccessPolicyResolver) (string, error) {
	switch resolver {
	case store.AccessPolicyIDResolver:
		return "id", nil
	default:
		return "", store.ErrResolverNotFound
	}
}
