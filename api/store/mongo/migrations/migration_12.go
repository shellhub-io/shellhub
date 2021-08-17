package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration12 = migrate.Migration{
	Version:     12,
	Description: "Set the tenant_id as unique in the namespaces collection",
	Up: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   12,
			"action":    "Up",
		}).Info("Applying migration")
		mod := mongo.IndexModel{
			Keys:    bson.D{{"tenant_id", 1}},
			Options: options.Index().SetName("tenant_id").SetUnique(true),
		}
		if _, err := db.Collection("namespaces").Indexes().CreateOne(context.TODO(), mod); err != nil {
			return err
		}
		mod = mongo.IndexModel{
			Keys:    bson.D{{"name", 1}},
			Options: options.Index().SetName("name").SetUnique(true),
		}
		_, err := db.Collection("namespaces").Indexes().CreateOne(context.TODO(), mod)

		return err
	},
	Down: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   12,
			"action":    "Down",
		}).Info("Applying migration")
		_, err := db.Collection("namespaces").Indexes().DropOne(context.TODO(), "tenant_id")

		return err
	},
}
