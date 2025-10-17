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
	tag.CreatedAt = clock.Now()
	tag.UpdatedAt = clock.Now()

	e := entity.TagFromModel(tag)
	e.ID = uuid.Generate()

	var result entity.Tag
	err := pg.driver.NewInsert().
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
	query := pg.driver.NewSelect().Model((*entity.Tag)(nil)).Column("name").Where("namespace_id = ?", tenantID)

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
	entities := make([]entity.Tag, 0)
	query := pg.driver.NewSelect().Model(&entities).Column("tag.*")
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
	tag := new(entity.Tag)
	query := pg.driver.NewSelect().Model(tag).Column("tag.*").Relation("Namespace")
	switch resolver {
	case store.TagIDResolver:
		query = query.Where("tag.id = ?", value)
	case store.TagNameResolver:
		query = query.Where("tag.name = ?", value)
	}

	if err := applyOptions(ctx, query, opts...); err != nil {
		return nil, fromSqlError(err)
	}

	if err := query.Scan(ctx); err != nil {
		return nil, fromSqlError(err)
	}

	return entity.TagToModel(tag), nil
}

func (pg *Pg) TagUpdate(ctx context.Context, id string, changes *models.TagChanges) error {
	updateMap := map[string]interface{}{
		"updated_at": clock.Now(),
	}

	if changes.Name != "" {
		updateMap["name"] = changes.Name
	}

	res, err := pg.driver.NewUpdate().
		Model((*entity.Tag)(nil)).
		Set("name = ?", changes.Name).
		Set("updated_at = ?", clock.Now()).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fromSqlError(err)
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

func (pg *Pg) TagPushToTarget(ctx context.Context, id string, target store.TagTarget, targetID string) error {
	tag := new(entity.Tag)
	if err := pg.driver.NewSelect().Model(tag).Where("id = ?", id).Scan(ctx); err != nil {
		return fromSqlError(err)
	}

	switch target {
	case store.TagTargetDevice:
		deviceTag := entity.NewDeviceTag(tag.ID, targetID)
		deviceTag.CreatedAt = clock.Now()

		if _, err := pg.driver.NewInsert().Model(deviceTag).On("CONFLICT (device_id, tag_id) DO NOTHING").Exec(ctx); err != nil {
			return fromSqlError(err)
		}
	case store.TagTargetPublicKey:
		publickeyTag := entity.NewPublicKeyTag(tag.ID, targetID)
		publickeyTag.CreatedAt = clock.Now()

		if _, err := pg.driver.NewInsert().Model(publickeyTag).On("CONFLICT (public_key_id, tag_id) DO NOTHING").Exec(ctx); err != nil {
			return fromSqlError(err)
		}
	}

	return nil
}

func (pg *Pg) TagPullFromTarget(ctx context.Context, id string, target store.TagTarget, targetIDs ...string) error {
	tag := new(entity.Tag)
	if err := pg.driver.NewSelect().Model(tag).Where("id = ?", id).Scan(ctx); err != nil {
		return fromSqlError(err)
	}

	switch target {
	case store.TagTargetDevice:
		query := pg.driver.NewDelete().Model((*entity.DeviceTag)(nil)).Where("tag_id = ?", id)
		if len(targetIDs) > 0 {
			query = query.Where("device_id IN (?)", bun.In(targetIDs))
		}

		if _, err := query.Exec(ctx); err != nil {
			return fromSqlError(err)
		}
	case store.TagTargetPublicKey:
		query := pg.driver.NewDelete().Model((*entity.PublicKeyTag)(nil)).Where("tag_id = ?", id)
		if len(targetIDs) > 0 {
			query = query.Where("public_key_id IN (?)", bun.In(targetIDs))
		}

		if _, err := query.Exec(ctx); err != nil {
			return fromSqlError(err)
		}
	}

	return nil
}

func (pg *Pg) TagDelete(ctx context.Context, id string) error {
	// Cascade will delete the relationships
	res, err := pg.driver.NewDelete().Model((*entity.Tag)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fromSqlError(err)
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		return store.ErrNoDocuments
	}

	return nil
}
