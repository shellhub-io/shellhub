package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
)

// DeviceCreateTag pushes a new tag to the existing set of tags for a specific device.
func (s *Store) DeviceCreateTag(ctx context.Context, uid models.UID, tag string) error {
	t, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$push": bson.M{"tags": tag}})
	if err != nil {
		return FromMongoError(err)
	}

	if t.ModifiedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

// DeviceRemoveTag removes a specific tag from the set of tags of a device.
func (s *Store) DeviceRemoveTag(ctx context.Context, uid models.UID, tag string) error {
	t, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$pull": bson.M{"tags": tag}})
	if err != nil {
		return FromMongoError(err)
	}

	// ModifiedCount is 0 when no tag is updated
	if t.ModifiedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

// DeviceUpdateTags replaces the existing set of tags for a specific device with a new set.
func (s *Store) DeviceUpdateTag(ctx context.Context, uid models.UID, tags []string) error {
	t, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": bson.M{"tags": tags}})
	if err != nil {
		return FromMongoError(err)
	}

	if t.ModifiedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}
