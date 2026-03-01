package migrate

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

type mongoPublicKey struct {
	Fingerprint string        `bson:"fingerprint"`
	TenantID    string        `bson:"tenant_id"`
	Data        []byte        `bson:"data"`
	CreatedAt   time.Time     `bson:"created_at"`
	Name        string        `bson:"name"`
	Username    string        `bson:"username"`
	Filter      mongoPKFilter `bson:"filter"`
}

type mongoPKFilter struct {
	Hostname string   `bson:"hostname"`
	TagIDs   []string `bson:"tag_ids"`
}

func convertPublicKey(doc mongoPublicKey) *entity.PublicKey {
	return &entity.PublicKey{
		Fingerprint:    doc.Fingerprint,
		NamespaceID:    doc.TenantID,
		CreatedAt:      doc.CreatedAt,
		UpdatedAt:      time.Time{},
		Name:           doc.Name,
		Username:       doc.Username,
		Data:           doc.Data,
		FilterHostname: doc.Filter.Hostname,
	}
}

func convertAPIKey(doc mongoAPIKey) *entity.APIKey {
	return &entity.APIKey{
		KeyDigest:   doc.ID,
		NamespaceID: doc.TenantID,
		Name:        doc.Name,
		Role:        doc.Role,
		UserID:      ObjectIDToUUID(doc.CreatedBy),
		CreatedAt:   doc.CreatedAt,
		UpdatedAt:   doc.UpdatedAt,
		ExpiresIn:   doc.ExpiresIn,
	}
}

func (m *Migrator) migratePublicKeys(ctx context.Context) error {
	cursor, err := m.mongo.Collection("public_keys").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	batch := make([]*entity.PublicKey, 0, batchSize)
	total := 0

	for cursor.Next(ctx) {
		var doc mongoPublicKey
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		batch = append(batch, convertPublicKey(doc))
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

	log.WithField("count", total).Info("Migrated public_keys")

	return nil
}

func (m *Migrator) migratePublicKeyTags(ctx context.Context) error {
	cursor, err := m.mongo.Collection("public_keys").Find(ctx, bson.M{"filter.tag_ids": bson.M{"$exists": true, "$ne": bson.A{}}})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	batch := make([]*entity.PublicKeyTag, 0, batchSize)
	total := 0

	for cursor.Next(ctx) {
		var doc mongoPublicKey
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		for _, tagID := range doc.Filter.TagIDs {
			e := &entity.PublicKeyTag{
				PublicKeyFingerprint: doc.Fingerprint,
				TagID:                ObjectIDToUUID(tagID),
				CreatedAt:            doc.CreatedAt,
			}
			batch = append(batch, e)

			if len(batch) >= batchSize {
				if _, err := m.pg.NewInsert().Model(&batch).Exec(ctx); err != nil {
					return err
				}
				total += len(batch)
				batch = batch[:0]
			}
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

	log.WithField("count", total).Info("Migrated public_key_tags")

	return nil
}

type mongoAPIKey struct {
	ID        string    `bson:"_id"`
	Name      string    `bson:"name"`
	TenantID  string    `bson:"tenant_id"`
	Role      string    `bson:"role"`
	CreatedBy string    `bson:"created_by"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
	ExpiresIn int64     `bson:"expires_in"`
}

func (m *Migrator) migrateAPIKeys(ctx context.Context) error {
	cursor, err := m.mongo.Collection("api_keys").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	batch := make([]*entity.APIKey, 0, batchSize)
	total := 0

	for cursor.Next(ctx) {
		var doc mongoAPIKey
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		batch = append(batch, convertAPIKey(doc))
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

	log.WithField("count", total).Info("Migrated api_keys")

	return nil
}
