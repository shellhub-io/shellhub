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
	cursor, err := m.mongo.Collection("tags").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	batch := make([]*entity.Tag, 0, batchSize)
	total := 0

	for cursor.Next(ctx) {
		var doc mongoTag
		if err := cursor.Decode(&doc); err != nil {
			return err
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

	log.WithField("count", total).Info("Migrated tags")

	return nil
}
