package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *Store) FirewallRulePushTag(ctx context.Context, id, tag string) error {
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

func (s *Store) FirewallRulePullTag(ctx context.Context, id, tag string) error {
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

func (s *Store) FirewallRuleSetTags(ctx context.Context, id string, tags []string) error {
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

func (s *Store) FirewallRuleBulkRenameTag(ctx context.Context, tenant, currentTag, newTag string) (int64, error) {
	res, err := s.db.Collection("firewall_rules").UpdateMany(ctx, bson.M{"tenant_id": tenant, "filter.tags": currentTag}, bson.M{"$set": bson.M{"filter.tags.$": newTag}})

	return res.ModifiedCount, FromMongoError(err)
}

func (s *Store) FirewallRuleBulkDeleteTag(ctx context.Context, tenant, tag string) (int64, error) {
	res, err := s.db.Collection("firewall_rules").UpdateMany(ctx, bson.M{"tenant_id": tenant}, bson.M{"$pull": bson.M{"filter.tags": tag}})

	return res.ModifiedCount, FromMongoError(err)
}

func (s *Store) FirewallRuleGetTags(ctx context.Context, tenant string) ([]string, int, error) {
	list, err := s.db.Collection("firewall_rules").Distinct(ctx, "filter.tags", bson.M{"tenant_id": tenant})

	tags := make([]string, len(list))
	for i, item := range list {
		tags[i] = item.(string) //nolint:forcetypeassert
	}

	return tags, len(tags), FromMongoError(err)
}
