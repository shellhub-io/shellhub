package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *Store) FirewallRuleAddTag(ctx context.Context, id, tag string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return FromMongoError(err)
	}

	result, err := s.db.Collection("firewall_rules").UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$addToSet": bson.M{"filter.tags": tag}})
	if err != nil {
		return err
	}

	if result.ModifiedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (s *Store) FirewallRuleRemoveTag(ctx context.Context, id, tag string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return FromMongoError(err)
	}

	result, err := s.db.Collection("firewall_rules").UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$pull": bson.M{"filter.tags": tag}})
	if err != nil {
		return err
	}

	if result.ModifiedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (s *Store) FirewallRuleUpdateTags(ctx context.Context, id string, tags []string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return FromMongoError(err)
	}

	result, err := s.db.Collection("firewall_rules").UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"filter.tags": tags}})
	if err != nil {
		return err
	}

	if result.ModifiedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (s *Store) FirewallRuleRenameTag(ctx context.Context, tenant, currentTags, newTags string) (int64, error) {
	res, err := s.db.Collection("firewall_rules").UpdateMany(ctx, bson.M{"tenant_id": tenant, "filter.tags": currentTags}, bson.M{"$set": bson.M{"filter.tags.$": newTags}})

	return res.ModifiedCount, FromMongoError(err)
}

func (s *Store) FirewallRuleDeleteTag(ctx context.Context, tenant, tag string) (int64, error) {
	res, err := s.db.Collection("firewall_rules").UpdateMany(ctx, bson.M{"tenant_id": tenant}, bson.M{"$pull": bson.M{"filter.tags": tag}})

	return res.ModifiedCount, FromMongoError(err)
}

func (s *Store) FirewallRuleGetTags(ctx context.Context, tenant string) ([]string, int, error) {
	list, err := s.db.Collection("firewall_rules").Distinct(ctx, "filter.tags", bson.M{"tenant_id": tenant})

	tags := make([]string, len(list))
	for i, item := range list {
		tags[i] = item.(string) //nolint:forcetypeassert
	}

	return tags, len(tags), err
}
