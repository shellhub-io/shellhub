package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/server/api/store"
	"go.mongodb.org/mongo-driver/bson"
)

func (s *Store) PublicKeyPushTag(ctx context.Context, tenant, fingerprint, tag string) error {
	result, err := s.db.Collection("public_keys").UpdateOne(ctx, bson.M{"tenant_id": tenant, "fingerprint": fingerprint}, bson.M{"$addToSet": bson.M{"filter.tags": tag}})
	if err != nil {
		return err
	}

	if result.ModifiedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (s *Store) PublicKeyPullTag(ctx context.Context, tenant, fingerprint, tag string) error {
	result, err := s.db.Collection("public_keys").UpdateOne(ctx, bson.M{"tenant_id": tenant, "fingerprint": fingerprint}, bson.M{"$pull": bson.M{"filter.tags": tag}})
	if err != nil {
		return err
	}

	if result.ModifiedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (s *Store) PublicKeySetTags(ctx context.Context, tenant, fingerprint string, tags []string) (int64, int64, error) {
	res, err := s.db.Collection("public_keys").UpdateOne(ctx, bson.M{"tenant_id": tenant, "fingerprint": fingerprint}, bson.M{"$set": bson.M{"filter.tags": tags}})

	return res.MatchedCount, res.ModifiedCount, FromMongoError(err)
}

func (s *Store) PublicKeyBulkRenameTag(ctx context.Context, tenant, currentTag, newTag string) (int64, error) {
	res, err := s.db.Collection("public_keys").UpdateMany(ctx, bson.M{"tenant_id": tenant, "filter.tags": currentTag}, bson.M{"$set": bson.M{"filter.tags.$": newTag}})

	return res.ModifiedCount, FromMongoError(err)
}

func (s *Store) PublicKeyBulkDeleteTag(ctx context.Context, tenant, tag string) (int64, error) {
	res, err := s.db.Collection("public_keys").UpdateMany(ctx, bson.M{"tenant_id": tenant}, bson.M{"$pull": bson.M{"filter.tags": tag}})

	return res.ModifiedCount, FromMongoError(err)
}

func (s *Store) PublicKeyGetTags(ctx context.Context, tenant string) ([]string, int, error) {
	list, err := s.db.Collection("public_keys").Distinct(ctx, "filter.tags", bson.M{"tenant_id": tenant})

	tags := make([]string, len(list))
	for i, item := range list {
		tags[i] = item.(string) //nolint:forcetypeassert
	}

	return tags, len(tags), FromMongoError(err)
}
