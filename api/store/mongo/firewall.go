package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mongo/queries"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Store) FirewallRuleList(ctx context.Context, paginator query.Paginator) ([]models.FirewallRule, int, error) {
	query := []bson.M{
		{
			"$sort": bson.M{
				"priority": 1,
			},
		},
	}

	// Only match for the respective tenant if requested
	if tenant := gateway.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	queryCount := query
	queryCount = append(queryCount, bson.M{"$count": "count"})
	count, err := AggregateCount(ctx, s.db.Collection("firewall_rules"), queryCount)
	if err != nil {
		return nil, 0, FromMongoError(err)
	}

	query = append(query, queries.FromPaginator(&paginator)...)

	rules := make([]models.FirewallRule, 0)
	cursor, err := s.db.Collection("firewall_rules").Aggregate(ctx, query)
	if err != nil {
		return nil, 0, FromMongoError(err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		rule := new(models.FirewallRule)
		err = cursor.Decode(&rule)
		if err != nil {
			return rules, count, FromMongoError(err)
		}

		rules = append(rules, *rule)
	}

	return rules, count, FromMongoError(err)
}

func (s *Store) FirewallRuleCreate(ctx context.Context, rule *models.FirewallRule) error {
	if err := rule.Validate(); err != nil {
		return FromMongoError(err)
	}

	if _, err := s.db.Collection("firewall_rules").InsertOne(ctx, &rule); err != nil {
		return FromMongoError(err)
	}

	return nil
}

func (s *Store) FirewallRuleGet(ctx context.Context, id string) (*models.FirewallRule, error) {
	rule := new(models.FirewallRule)
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, FromMongoError(err)
	}

	if err := s.db.Collection("firewall_rules").FindOne(ctx, bson.M{"_id": objID}).Decode(&rule); err != nil {
		return nil, FromMongoError(err)
	}

	return rule, nil
}

func (s *Store) FirewallRuleUpdate(ctx context.Context, id string, rule models.FirewallRuleUpdate) (*models.FirewallRule, error) {
	if err := rule.Validate(); err != nil {
		return nil, FromMongoError(err)
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, FromMongoError(err)
	}

	updateOpts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	result := s.db.Collection("firewall_rules").FindOneAndUpdate(ctx, bson.M{"_id": objID}, bson.M{"$set": rule}, updateOpts)

	if result.Err() != nil {
		return nil, FromMongoError(result.Err())
	}

	firewallRule := new(models.FirewallRule)
	if err := result.Decode(&firewallRule); err != nil {
		return nil, FromMongoError(err)
	}

	return firewallRule, nil
}

func (s *Store) FirewallRuleDelete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return FromMongoError(err)
	}

	fRule, err := s.db.Collection("firewall_rules").DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return FromMongoError(err)
	}

	if fRule.DeletedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}
