package mongo

import (
	"context"
	"fmt"

	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
)

func (s *Store) DeviceCreateTag(ctx context.Context, uid models.UID, tag string) error {
	_, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$push": bson.M{"tags": tag}})

	return err
}

func (s *Store) DeviceDeleteTag(ctx context.Context, uid models.UID, tag string) error {
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
		if _, err := s.db.Collection("devices").UpdateMany(ctx,
			bson.M{"tags": currentTagName, "tenant_id": tenantID},
			bson.M{"$set": bson.M{"tags.$": newTagName}}); err != nil {
			return nil, fromMongoError(err)
		}

		return nil, nil
	})

	return err
}

func (s *Store) DeviceListTag(ctx context.Context) ([]string, int, error) {
	tagList, err := s.db.Collection("devices").Distinct(ctx, "tags", bson.M{})

	tags := make([]string, len(tagList))
	for i, v := range tagList {
		tags[i] = fmt.Sprint(v)
	}

	return tags, len(tags), err
}

func (s *Store) DeviceUpdateTag(ctx context.Context, uid models.UID, tags []string) error {
	_, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": bson.M{"tags": tags}})

	return err
}

func (s *Store) DeviceGetTags(ctx context.Context, tenantID string) ([]string, int, error) {
	tagList, err := s.db.Collection("devices").Distinct(ctx, "tags", bson.M{"tenant_id": tenantID})

	tags := make([]string, len(tagList))
	for i, v := range tagList {
		tags[i] = fmt.Sprint(v)
	}

	return tags, len(tags), err
}

func (s *Store) DeviceDeleteAllTags(ctx context.Context, tenantID string, tagName string) error {
	_, err := s.db.Collection("devices").UpdateMany(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$pull": bson.M{"tags": tagName}})

	return err
}
