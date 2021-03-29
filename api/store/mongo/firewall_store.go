package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *Store) CreateFirewallRule(ctx context.Context, rule *models.FirewallRule) error {
	if err := rule.Validate(); err != nil {
		return err
	}

	rule.ID = primitive.NewObjectID().Hex()

	if _, err := s.db.Collection("firewall_rules").InsertOne(ctx, &rule); err != nil {
		return err
	}

	return nil
}

func (s *Store) ListFirewallRules(ctx context.Context, pagination paginator.Query) ([]models.FirewallRule, int, error) {
	query := []bson.M{
		{
			"$sort": bson.M{
				"priority": 1,
			},
		},
	}

	// Only match for the respective tenant if requested
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	queryCount := append(query, bson.M{"$count": "count"})
	count, err := aggregateCount(ctx, s.db.Collection("firewall_rules"), queryCount)
	if err != nil {
		return nil, 0, err
	}

	query = append(query, buildPaginationQuery(pagination)...)

	rules := make([]models.FirewallRule, 0)
	cursor, err := s.db.Collection("firewall_rules").Aggregate(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		rule := new(models.FirewallRule)
		err = cursor.Decode(&rule)
		if err != nil {
			return rules, count, err
		}

		rules = append(rules, *rule)
	}

	return rules, count, err
}

func (s *Store) GetFirewallRule(ctx context.Context, id string) (*models.FirewallRule, error) {
	rule := new(models.FirewallRule)
	if err := s.db.Collection("firewall_rules").FindOne(ctx, bson.M{"_id": id}).Decode(&rule); err != nil {
		return nil, err
	}

	return rule, nil
}

func (s *Store) UpdateFirewallRule(ctx context.Context, id string, rule models.FirewallRuleUpdate) (*models.FirewallRule, error) {
	if err := rule.Validate(); err != nil {
		return nil, err
	}

	if _, err := s.db.Collection("firewall_rules").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": rule}); err != nil {
		return nil, err
	}

	r, err := s.GetFirewallRule(ctx, id)
	return r, err
}

func (s *Store) DeleteFirewallRule(ctx context.Context, id string) error {
	if _, err := s.db.Collection("firewall_rules").DeleteOne(ctx, bson.M{"_id": id}); err != nil {
		return err
	}

	return nil
}
