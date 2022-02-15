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
