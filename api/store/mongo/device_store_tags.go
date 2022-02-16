package mongo

import (
	"context"
	"fmt"

	"github.com/emirpasic/gods/sets/hashset"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
)

func (s *Store) DeviceCreateTag(ctx context.Context, uid models.UID, tag string) error {
	_, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$push": bson.M{"tags": tag}})

	return err
}

func (s *Store) DeviceRemoveTag(ctx context.Context, uid models.UID, tag string) error {
	_, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$pull": bson.M{"tags": tag}})

	return err
}

func (s *Store) DeviceRenameTag(ctx context.Context, tenantID string, currentTagName string, newTagName string) error {
	// Create a session to run the transaction.
	session, err := s.db.Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	// Rename all devices tags inside a transaction.
	_, err = session.WithTransaction(ctx, func(sessCtx mongodriver.SessionContext) (interface{}, error) {
		_, err := s.db.Collection("devices").UpdateMany(ctx, bson.M{"tags": currentTagName, "tenant_id": tenantID}, bson.M{"$set": bson.M{"tags.$": newTagName}})
		if err != nil {
			return nil, fromMongoError(err)
		}

		if err := s.PublicKeyRenameTag(ctx, tenantID, currentTagName, newTagName); err != store.ErrNoDocuments {
			return nil, err
		}

		return nil, nil
	})

	return err
}

func (s *Store) DeviceUpdateTag(ctx context.Context, uid models.UID, tags []string) error {
	_, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": bson.M{"tags": tags}})

	return err
}

func (s *Store) DeviceGetTags(ctx context.Context, tenantID string) ([]string, int, error) {
	tagsDevice, err := s.db.Collection("devices").Distinct(ctx, "tags", bson.M{"tenant_id": tenantID})
	if err != nil {
		return nil, 0, err
	}

	tagsKey, err := s.db.Collection("public_keys").Distinct(ctx, "filter.tags", bson.M{"tenant_id": tenantID})
	if err != nil {
		return nil, 0, err
	}

	tagsSet := hashset.New()
	tagsSet.Add(tagsDevice...)
	tagsSet.Add(tagsKey...)

	tags := make([]string, tagsSet.Size())
	for i, v := range tags {
		tags[i] = fmt.Sprint(v)
	}

	return tags, len(tags), err
}

func (s *Store) DeviceDeleteTags(ctx context.Context, tenantID string, tagName string) error {
	_, err := s.db.Collection("devices").UpdateMany(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$pull": bson.M{"tags": tagName}})

	if err := s.PublicKeyDeleteTag(ctx, tenantID, tagName); err != store.ErrNoDocuments {
		return err
	}

	return err
}
