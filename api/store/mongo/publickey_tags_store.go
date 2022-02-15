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

	if result.ModifiedCount <= 0 {
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

	if result.ModifiedCount <= 0 {
		return store.ErrNoDocuments
	}

	return nil
}

// PublicKeyUpdateTags update with a new set the tag's list in models.PublicKey.
//
// To update models.PublicKey with a new set, all tags need to exist on a models.Device. If it is not true, the update
// action will fail.
func (s *Store) PublicKeyUpdateTags(ctx context.Context, tenant, fingerprint string, tags []string) error {
	// If all tags exist in device, set the tags to tag's field in models.PublicKey.
	result, err := s.db.Collection("public_keys").UpdateOne(ctx, bson.M{"tenant_id": tenant, "fingerprint": fingerprint}, bson.M{"$set": bson.M{"filter.tags": tags}})
	if err != nil {
		return err
	}

	if result.ModifiedCount <= 0 {
		return store.ErrNoDocuments
	}

	return nil
}

// PublicKeyRenameTag renames a tag to a new name.
func (s *Store) PublicKeyRenameTag(ctx context.Context, tenant, old, neo string) error {
	result, err := s.db.Collection("public_keys").UpdateMany(ctx, bson.M{"tenant_id": tenant, "filter.tags": old}, bson.M{"$set": bson.M{"filter.tags.$": neo}})
	if err != nil {
		return err
	}

	if result.ModifiedCount <= 0 {
		return store.ErrNoDocuments
	}

	return nil
}
