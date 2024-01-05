package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
)

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

func (s *Store) DeviceRemoveTag(ctx context.Context, uid models.UID, tag string) error {
	t, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$pull": bson.M{"tags": tag}})
	if err != nil {
		return FromMongoError(err)
	}

	if t.ModifiedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (s *Store) DeviceUpdateTag(ctx context.Context, uid models.UID, tags []string) (int64, int64, error) {
	tag, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": bson.M{"tags": tags}})

	return tag.MatchedCount, tag.ModifiedCount, FromMongoError(err)
}

func (s *Store) DeviceRenameTag(ctx context.Context, tenant, currentTags, newTags string) (int64, error) {
	res, err := s.db.Collection("devices").UpdateMany(ctx, bson.M{"tenant_id": tenant, "tags": currentTags}, bson.M{"$set": bson.M{"tags.$": newTags}})

	return res.ModifiedCount, FromMongoError(err)
}

func (s *Store) DeviceDeleteTag(ctx context.Context, tenant, tag string) (int64, error) {
	res, err := s.db.Collection("devices").UpdateMany(ctx, bson.M{"tenant_id": tenant}, bson.M{"$pull": bson.M{"tags": tag}})

	return res.ModifiedCount, FromMongoError(err)
}
