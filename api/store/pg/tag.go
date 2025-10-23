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
	db := pg.getExecutor(ctx)

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
		return "", fromSqlError(err)
	}

	return result.ID, nil
}

func (pg *Pg) TagConflicts(ctx context.Context, tenantID string, target *models.TagConflicts) ([]string, bool, error) {
	db := pg.getExecutor(ctx)

	query := db.NewSelect().Model((*entity.Tag)(nil)).Column("name").Where("namespace_id = ?", tenantID)
	if target.Name != "" {
		query = query.Where("name = ?", target.Name)
	}

	tags := make([]map[string]any, 0)
	if err := query.Scan(ctx, &tags); err != nil {
		return nil, false, fromSqlError(err)
	}

	conflicts := make([]string, 0)
	for _, tag := range tags {
		if tag["name"] == target.Name {
			conflicts = append(conflicts, "name")
		}
	}

	return conflicts, len(conflicts) > 0, nil
}

func (pg *Pg) TagList(ctx context.Context, opts ...store.QueryOption) ([]models.Tag, int, error) {
	db := pg.getExecutor(ctx)

	entities := make([]entity.Tag, 0)
	query := db.NewSelect().Model(&entities).Column("tag.*")
	if err := applyOptions(ctx, query, opts...); err != nil {
		return nil, 0, fromSqlError(err)
	}

	count, err := query.ScanAndCount(ctx)
	if err != nil {
		return nil, 0, fromSqlError(err)
	}

	tags := make([]models.Tag, len(entities))
	for i, e := range entities {
		tags[i] = *entity.TagToModel(&e)
	}

	return tags, count, nil
}

func (pg *Pg) TagResolve(ctx context.Context, resolver store.TagResolver, value string, opts ...store.QueryOption) (*models.Tag, error) {
	db := pg.getExecutor(ctx)

	column, err := TagResolverToString(resolver)
	if err != nil {
		return nil, err
	}

	tag := new(entity.Tag)
	query := db.NewSelect().Model(tag).Column("tag.*").Relation("Namespace").Where("tag.? = ?", bun.Ident(column), value)

	if err := applyOptions(ctx, query, opts...); err != nil {
		return nil, fromSqlError(err)
	}

	if err := query.Scan(ctx); err != nil {
		return nil, fromSqlError(err)
	}

	return entity.TagToModel(tag), nil
}

func (pg *Pg) TagUpdate(ctx context.Context, tag *models.Tag) error {
	db := pg.getExecutor(ctx)

	t := entity.TagFromModel(tag)
	t.UpdatedAt = clock.Now()

	r, err := db.NewUpdate().Model(t).WherePK().Exec(ctx)
	if err != nil {
		return fromSqlError(err)
	}

	if count, err := r.RowsAffected(); err != nil || count == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

func (pg *Pg) TagPushToTarget(ctx context.Context, id string, target store.TagTarget, targetID string) error {
	db := pg.getExecutor(ctx)

	tag := new(entity.Tag)
	if err := db.NewSelect().Model(tag).Where("id = ?", id).Scan(ctx); err != nil {
		return fromSqlError(err)
	}

	switch target {
	case store.TagTargetDevice:
		deviceTag := entity.NewDeviceTag(tag.ID, targetID)
		deviceTag.CreatedAt = clock.Now()

		if _, err := db.NewInsert().Model(deviceTag).On("CONFLICT (device_id, tag_id) DO NOTHING").Exec(ctx); err != nil {
			return fromSqlError(err)
		}
	case store.TagTargetPublicKey:
		publickeyTag := entity.NewPublicKeyTag(tag.ID, targetID)
		publickeyTag.CreatedAt = clock.Now()

		if _, err := db.NewInsert().Model(publickeyTag).On("CONFLICT (public_key_id, tag_id) DO NOTHING").Exec(ctx); err != nil {
			return fromSqlError(err)
		}
	}

	return nil
}

func (pg *Pg) TagPullFromTarget(ctx context.Context, id string, target store.TagTarget, targetIDs ...string) error {
	db := pg.getExecutor(ctx)

	tag := new(entity.Tag)
	if err := db.NewSelect().Model(tag).Where("id = ?", id).Scan(ctx); err != nil {
		return fromSqlError(err)
	}

	switch target {
	case store.TagTargetDevice:
		query := db.NewDelete().Model((*entity.DeviceTag)(nil)).Where("tag_id = ?", id)
		if len(targetIDs) > 0 {
			query = query.Where("device_id IN (?)", bun.In(targetIDs))
		}

		if _, err := query.Exec(ctx); err != nil {
			return fromSqlError(err)
		}
	case store.TagTargetPublicKey:
		query := db.NewDelete().Model((*entity.PublicKeyTag)(nil)).Where("tag_id = ?", id)
		if len(targetIDs) > 0 {
			query = query.Where("public_key_id IN (?)", bun.In(targetIDs))
		}

		if _, err := query.Exec(ctx); err != nil {
			return fromSqlError(err)
		}
	}

	return nil
}

func (pg *Pg) TagDelete(ctx context.Context, tag *models.Tag) error {
	db := pg.getExecutor(ctx)

	t := entity.TagFromModel(tag)

	r, err := db.NewDelete().Model(t).WherePK().Exec(ctx)
	if err != nil {
		return fromSqlError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return fromSqlError(err)
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
