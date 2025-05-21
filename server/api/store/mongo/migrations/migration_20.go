package migrations

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration20 = migrate.Migration{
	Version:     20,
	Description: "Change the model on db for firewall_rules collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   20,
			"action":    "Up",
		}).Info("Applying migration")
		cursor, err := db.Collection("firewall_rules").Find(ctx, bson.D{})
		if err != nil {
			return err
		}

		type firewallRule struct {
			ID                        primitive.ObjectID `json:"id" bson:"_id"`
			TenantID                  string             `json:"tenant_id" bson:"tenant_id"`
			models.FirewallRuleFields `bson:",inline"`
		}

		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			firewall := new(models.FirewallRule)
			err := cursor.Decode(&firewall)
			if err != nil {
				return err
			}
			objID, err := primitive.ObjectIDFromHex(firewall.ID)
			replacedRule := firewallRule{
				TenantID:           firewall.TenantID,
				ID:                 objID,
				FirewallRuleFields: firewall.FirewallRuleFields,
			}

			if err == nil {
				if errDelete := db.Collection("firewall_rules").FindOneAndDelete(ctx, bson.M{"_id": firewall.ID}); errDelete.Err() != nil {
					continue
				}

				if _, err := db.Collection("firewall_rules").InsertOne(ctx, replacedRule); err != nil {
					return err
				}
			}
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   20,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}
