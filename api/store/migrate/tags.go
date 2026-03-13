package migrate

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type mongoTag struct {
	ID        primitive.ObjectID `bson:"_id"`
	TenantID  string             `bson:"tenant_id"`
	Name      string             `bson:"name"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

func convertTag(doc mongoTag) *entity.Tag {
	return &entity.Tag{
		ID:          ObjectIDToUUID(doc.ID.Hex()),
		NamespaceID: doc.TenantID,
		Name:        doc.Name,
		CreatedAt:   doc.CreatedAt,
		UpdatedAt:   doc.UpdatedAt,
	}
}

func (m *Migrator) migrateTags(ctx context.Context) error {
	validNS, err := m.loadValidNamespaces(ctx)
	if err != nil {
		return err
	}

	cursor, err := m.mongo.Collection("tags").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	batch := make([]*entity.Tag, 0, batchSize)
	total := 0
	skipped := 0

	for cursor.Next(ctx) {
		var doc mongoTag
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		if _, ok := validNS[doc.TenantID]; !ok {
			log.WithFields(log.Fields{
				"scope":     "core",
				"tag":       doc.Name,
				"namespace": doc.TenantID,
			}).Warn("Skipping tag with orphaned namespace")
			skipped++

			continue
		}

		batch = append(batch, convertTag(doc))
		if len(batch) >= batchSize {
			if _, err := m.pg.NewInsert().Model(&batch).Exec(ctx); err != nil {
				return err
			}
			total += len(batch)
			batch = batch[:0]
		}
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	if len(batch) > 0 {
		if _, err := m.pg.NewInsert().Model(&batch).Exec(ctx); err != nil {
			return err
		}
		total += len(batch)
	}

	if skipped > 0 {
		m.addOrphans("tags", skipped)
	}

	log.WithFields(log.Fields{
		"scope":   "core",
		"count":   total,
		"skipped": skipped,
	}).Info("Migrated tags")

	return nil
}
