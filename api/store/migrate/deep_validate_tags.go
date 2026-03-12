package migrate

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/uptrace/bun"
	"go.mongodb.org/mongo-driver/bson"
)

func (m *Migrator) deepValidateTags(ctx context.Context, r *ValidationReport) error {
	validNS, err := m.loadValidNamespaces(ctx)
	if err != nil {
		return err
	}

	cursor, err := m.mongo.Collection("tags").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx) //nolint:errcheck

	batch := make([]mongoTag, 0, batchSize)

	for cursor.Next(ctx) {
		var doc mongoTag
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		// Skip tags whose namespace was not migrated (orphaned).
		if _, ok := validNS[doc.TenantID]; !ok {
			continue
		}

		batch = append(batch, doc)
		if len(batch) >= batchSize {
			if err := m.compareTagBatch(ctx, r, batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	if len(batch) > 0 {
		return m.compareTagBatch(ctx, r, batch)
	}

	return nil
}

func (m *Migrator) compareTagBatch(ctx context.Context, r *ValidationReport, batch []mongoTag) error {
	ids := make([]string, len(batch))
	expected := make(map[string]*entity.Tag, len(batch))
	for i, doc := range batch {
		e := convertTag(doc)
		ids[i] = e.ID
		expected[e.ID] = e
	}

	var actual []entity.Tag
	if err := m.pg.NewSelect().Model(&actual).Where("id IN (?)", bun.List(ids)).Scan(ctx); err != nil {
		return err
	}

	pgMap := make(map[string]*entity.Tag, len(actual))
	for i := range actual {
		pgMap[actual[i].ID] = &actual[i]
	}

	r.AddCompared("tags", int64(len(batch)))

	for _, id := range ids {
		exp := expected[id]
		act, ok := pgMap[id]
		if !ok {
			r.AddMissing("tags", id)

			continue
		}

		t := "tags"
		r.CheckField(t, id, "NamespaceID", exp.NamespaceID, act.NamespaceID)
		r.CheckField(t, id, "Name", exp.Name, act.Name)
		r.CheckTime(t, id, "CreatedAt", exp.CreatedAt, act.CreatedAt)
		r.CheckTime(t, id, "UpdatedAt", exp.UpdatedAt, act.UpdatedAt)
	}

	return nil
}
