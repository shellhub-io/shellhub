package migrations

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// invertFirewallRulePriority inverts the priority of the firewall rules.
//
// The priority of the firewall rules is inverted to follow a common pattern in the industry.
//
// If any error occurs, the migration is aborted.
func invertFirewallRulePriority(db *mongo.Database) error {
	ctx := context.Background()

	type Properties struct {
		ID       string
		Priority int
	}

	options := new(options.FindOptions)
	options.SetSort(bson.D{{"priority", 1}}) // Sort by priority in ascending order.

	namespaces, err := db.Collection("namespaces").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer namespaces.Close(ctx)

	for namespaces.Next(ctx) {
		var properties []Properties

		var namespace models.Namespace
		if err := namespaces.Decode(&namespace); err != nil {
			return err
		}

		rules, err := db.Collection("firewall_rules").Find(ctx, bson.M{"tenant_id": namespace.TenantID}, options)
		if err != nil {
			return err
		}
		defer rules.Close(ctx)

		for rules.Next(ctx) {
			rule := new(models.FirewallRule)
			if err := rules.Decode(rule); err != nil {
				return err
			}

			properties = append(properties, Properties{
				ID:       rule.ID,
				Priority: rule.Priority,
			})
		}

		for index := 0; index <= len(properties)-1; index++ {
			id, _ := primitive.ObjectIDFromHex(properties[index].ID)

			_, err := db.Collection("firewall_rules").UpdateMany(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"priority": properties[len(properties)-1-index].Priority}})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

var migration48 = migrate.Migration{
	Version:     48,
	Description: "invert Firewall priority",
	Up: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   48,
			"action":    "Up",
		}).Info("Applying migration up")

		return invertFirewallRulePriority(db)
	},
	Down: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   48,
			"action":    "Down",
		}).Info("Applying migration down")

		return invertFirewallRulePriority(db)
	},
}
