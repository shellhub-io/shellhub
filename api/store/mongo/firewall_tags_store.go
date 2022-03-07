package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FirewallRuleAddTag adds a tag to the tag's list in models.FirewallRule.
//
// The tag needs to exist on a models.Device. If it is not, the tag addition to
// models.FirewallRule will fail.
func (s *Store) FirewallRuleAddTag(ctx context.Context, id, tag string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fromMongoError(err)
	}

	result, err := s.db.Collection("firewall_rules").UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$addToSet": bson.M{"filter.tags": tag}})
	if err != nil {
		return err
	}

	if result.ModifiedCount <= 0 {
		return store.ErrNoDocuments
	}

	return nil
}

// FirewallRuleRemoveTag removes a tag from the tag's list in models.FirewallRule.
//
// The tag needs to exist on a models.Device. If it is not, the tag deletion from
// models.FirewallRule will fail.
func (s *Store) FirewallRuleRemoveTag(ctx context.Context, id, tag string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fromMongoError(err)
	}

	result, err := s.db.Collection("firewall_rules").UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$pull": bson.M{"filter.tags": tag}})
	if err != nil {
		return err
	}

	if result.ModifiedCount <= 0 {
		return store.ErrNoDocuments
	}

	return nil
}

// FirewallRuleUpdateTags update with a new set the tag's list in models.FirewallRule.
//
// All tags need to exist on a models.Device. If it is not true, the tags' update
// to models.FirewallRule will fail.
func (s *Store) FirewallRuleUpdateTags(ctx context.Context, id string, tags []string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fromMongoError(err)
	}

	// If all tags exist in device, set the tags to tag's field in models.PublicKey.
	result, err := s.db.Collection("firewall_rules").UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"filter.tags": tags}})
	if err != nil {
		return err
	}

	if result.ModifiedCount <= 0 {
		return store.ErrNoDocuments
	}

	return nil
}

// FirewallRuleRenameTag renames a tag to a new name in models.FirewallRule.
func (s *Store) FirewallRuleRenameTag(ctx context.Context, tenant, tagCurrent, tagNew string) error {
	result, err := s.db.Collection("firewall_rules").UpdateMany(ctx, bson.M{"tenant_id": tenant, "filter.tags": tagCurrent}, bson.M{"$set": bson.M{"filter.tags.$": tagNew}})
	if err != nil {
		return err
	}

	if result.ModifiedCount <= 0 {
		return store.ErrNoDocuments
	}

	return nil
}

// FirewallRuleDeleteTag removes a tag from all models.FirewallRule.
func (s *Store) FirewallRuleDeleteTag(ctx context.Context, tenant, tag string) error {
	result, err := s.db.Collection("firewall_rules").UpdateMany(ctx, bson.M{"tenant_id": tenant}, bson.M{"$pull": bson.M{"filter.tags": tag}})
	if err != nil {
		return err
	}

	if result.ModifiedCount <= 0 {
		return store.ErrNoDocuments
	}

	return nil
}

// FirewallRuleGetTags gets all tags from all models.FirewallRule.
func (s *Store) FirewallRuleGetTags(ctx context.Context, tenant string) ([]string, int, error) {
	list, err := s.db.Collection("firewall_rules").Distinct(ctx, "filter.tags", bson.M{"tenant_id": tenant})

	tags := make([]string, len(list))
	for i, item := range list {
		tags[i] = item.(string)
	}

	return tags, len(tags), err
}
