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

func (pg *Pg) TagCreate(ctx context.Context, tag *models.Tag) (string, error) {
	db := pg.getConnection(ctx)

	tag.CreatedAt = clock.Now()
	tag.UpdatedAt = clock.Now()

	e := entity.TagFromModel(tag)
	e.ID = uuid.Generate()

	var result entity.Tag
	err := db.NewInsert().
		Model(e).
		On("CONFLICT (namespace_id, name) DO UPDATE SET updated_at = EXCLUDED.updated_at").
		Returning("*").
		Scan(ctx, &result)
	if err != nil {
		return "", fromSQLError(err)
	}

	return result.ID, nil
}

func (pg *Pg) TagConflicts(ctx context.Context, tenantID string, target *models.TagConflicts) ([]string, bool, error) {
	db := pg.getConnection(ctx)

	if target.Name == "" {
		return nil, false, nil
	}

	tags := make([]entity.Tag, 0)
	query := db.NewSelect().
		Model(&tags).
		Column("name").
		Where("namespace_id = ?", tenantID).
		WhereGroup(" OR ", func(q *bun.SelectQuery) *bun.SelectQuery {
			if target.Name != "" {
				q = q.Where("name = ?", target.Name)
			}

			return q
		})

	if err := query.Scan(ctx); err != nil {
		return nil, false, fromSQLError(err)
	}

	seen := make(map[string]bool)
	for _, tag := range tags {
		if target.Name != "" && tag.Name == target.Name {
			seen["name"] = true
		}
	}

	conflicts := make([]string, 0, len(seen))
	for field := range seen {
		conflicts = append(conflicts, field)
	}

	return conflicts, len(conflicts) > 0, nil
}

func (pg *Pg) TagList(ctx context.Context, opts ...store.QueryOption) ([]models.Tag, int, error) {
	db := pg.getConnection(ctx)

	entities := make([]entity.Tag, 0)
	query := db.NewSelect().Model(&entities).Column("tag.*")
	if err := applyOptions(ctx, query, opts...); err != nil {
		return nil, 0, fromSQLError(err)
	}

	count, err := query.ScanAndCount(ctx)
	if err != nil {
		return nil, 0, fromSQLError(err)
	}

	tags := make([]models.Tag, len(entities))
	for i, e := range entities {
		tags[i] = *entity.TagToModel(&e)
	}

	return tags, count, nil
}

func (pg *Pg) TagResolve(ctx context.Context, resolver store.TagResolver, value string, opts ...store.QueryOption) (*models.Tag, error) {
	db := pg.getConnection(ctx)

	column, err := TagResolverToString(resolver)
	if err != nil {
		return nil, err
	}

	tag := new(entity.Tag)
	query := db.NewSelect().Model(tag).Column("tag.*").Relation("Namespace").Where("tag.? = ?", bun.Ident(column), value)

	if err := applyOptions(ctx, query, opts...); err != nil {
		return nil, fromSQLError(err)
	}

	if err := query.Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.TagToModel(tag), nil
}

func (pg *Pg) TagUpdate(ctx context.Context, tag *models.Tag) error {
	db := pg.getConnection(ctx)

	t := entity.TagFromModel(tag)
	t.UpdatedAt = clock.Now()

	r, err := db.NewUpdate().Model(t).WherePK().Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if count, err := r.RowsAffected(); err != nil || count == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

func (pg *Pg) TagPushToTarget(ctx context.Context, id string, target store.TagTarget, targetID string) error {
	db := pg.getConnection(ctx)

	tag := new(entity.Tag)
	if err := db.NewSelect().Model(tag).Where("id = ?", id).Scan(ctx); err != nil {
		return fromSQLError(err)
	}

	switch target {
	case store.TagTargetDevice:
		deviceTag := entity.NewDeviceTag(tag.ID, targetID)
		deviceTag.CreatedAt = clock.Now()

		if _, err := db.NewInsert().Model(deviceTag).On("CONFLICT (device_id, tag_id) DO NOTHING").Exec(ctx); err != nil {
			return fromSQLError(err)
		}
	case store.TagTargetPublicKey:
		publickeyTag := entity.NewPublicKeyTag(tag.ID, targetID)
		publickeyTag.CreatedAt = clock.Now()

		if _, err := db.NewInsert().Model(publickeyTag).On("CONFLICT (public_key_id, tag_id) DO NOTHING").Exec(ctx); err != nil {
			return fromSQLError(err)
		}
	}

	return nil
}

func (pg *Pg) TagPullFromTarget(ctx context.Context, id string, target store.TagTarget, targetIDs ...string) error {
	db := pg.getConnection(ctx)

	tag := new(entity.Tag)
	if err := db.NewSelect().Model(tag).Where("id = ?", id).Scan(ctx); err != nil {
		return fromSQLError(err)
	}

	switch target {
	case store.TagTargetDevice:
		query := db.NewDelete().Model((*entity.DeviceTag)(nil)).Where("tag_id = ?", id)
		if len(targetIDs) > 0 {
			query = query.Where("device_id IN (?)", bun.In(targetIDs))
		}

		if _, err := query.Exec(ctx); err != nil {
			return fromSQLError(err)
		}
	case store.TagTargetPublicKey:
		query := db.NewDelete().Model((*entity.PublicKeyTag)(nil)).Where("tag_id = ?", id)
		if len(targetIDs) > 0 {
			query = query.Where("public_key_id IN (?)", bun.In(targetIDs))
		}

		if _, err := query.Exec(ctx); err != nil {
			return fromSQLError(err)
		}
	}

	return nil
}

func (pg *Pg) TagDelete(ctx context.Context, tag *models.Tag) error {
	db := pg.getConnection(ctx)

	t := entity.TagFromModel(tag)

	r, err := db.NewDelete().Model(t).WherePK().Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return fromSQLError(err)
}

func TagResolverToString(resolver store.TagResolver) (string, error) {
	switch resolver {
	case store.TagIDResolver:
		return "id", nil
	case store.TagNameResolver:
		return "name", nil
	default:
		return "", store.ErrResolverNotFound
	}
}
