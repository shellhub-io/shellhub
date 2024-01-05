package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"go.mongodb.org/mongo-driver/bson"
)

// PublicKeyAddTag adds a tag to the tag's list in models.PublicKey.
//
// To add a tag to a models.PublicKey, that tag needs to exist on a models.Device. If it is not, the tag addition to
// PublicKey will fail.
func (s *Store) PublicKeyAddTag(ctx context.Context, tenant, fingerprint, tag string) error {
	result, err := s.db.Collection("public_keys").UpdateOne(ctx, bson.M{"tenant_id": tenant, "fingerprint": fingerprint}, bson.M{"$addToSet": bson.M{"filter.tags": tag}})
	if err != nil {
		return err
	}

	if result.ModifiedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

// PublicKeyRemoveTag removes a tag to the tag's list in models.PublicKey.
//
// To remove a tag from a models.PublicKey, that tag needs to exist on a models.Device. If it is not, the tag deletion from
// PublicKey will fail.
func (s *Store) PublicKeyRemoveTag(ctx context.Context, tenant, fingerprint, tag string) error {
	result, err := s.db.Collection("public_keys").UpdateOne(ctx, bson.M{"tenant_id": tenant, "fingerprint": fingerprint}, bson.M{"$pull": bson.M{"filter.tags": tag}})
	if err != nil {
		return err
	}

	if result.ModifiedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

// PublicKeyUpdateTags sets the tags for a public key with the specified fingerprint and tenant.
// It returns the number of matching documents, the number of modified documents, and any encountered errors.
//
// All tags need to exist on a device. If it is not true, the update action will fail.
func (s *Store) PublicKeyUpdateTags(ctx context.Context, tenant, fingerprint string, tags []string) (int64, int64, error) {
	// If all tags exist in device, set the tags to tag's field in models.PublicKey.
	res, err := s.db.Collection("public_keys").UpdateOne(ctx, bson.M{"tenant_id": tenant, "fingerprint": fingerprint}, bson.M{"$set": bson.M{"filter.tags": tags}})

	return res.MatchedCount, res.ModifiedCount, FromMongoError(err)
}

func (s *Store) PublicKeyRenameTag(ctx context.Context, tenant, currentTags, newTags string) (int64, error) {
	res, err := s.db.Collection("public_keys").UpdateMany(ctx, bson.M{"tenant_id": tenant, "filter.tags": currentTags}, bson.M{"$set": bson.M{"filter.tags.$": newTags}})

	return res.ModifiedCount, FromMongoError(err)
}

func (s *Store) PublicKeyDeleteTag(ctx context.Context, tenant, tag string) (int64, error) {
	res, err := s.db.Collection("public_keys").UpdateMany(ctx, bson.M{"tenant_id": tenant}, bson.M{"$pull": bson.M{"filter.tags": tag}})

	return res.ModifiedCount, FromMongoError(err)
}

// PublicKeyGetTags gets all tags from public keys.
func (s *Store) PublicKeyGetTags(ctx context.Context, tenant string) ([]string, int, error) {
	list, err := s.db.Collection("public_keys").Distinct(ctx, "filter.tags", bson.M{"tenant_id": tenant})

	tags := make([]string, len(list))
	for i, item := range list {
		tags[i] = item.(string) //nolint:forcetypeassert
	}

	return tags, len(tags), err
}
