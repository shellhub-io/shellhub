package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration_12 = migrate.Migration{
	Version: 12,
	Up: func(db *mongo.Database) error {
		logrus.Info("Applying migration 12 - Up")
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
		logrus.Info("Applying migration 12 - Down")
		_, err := db.Collection("namespaces").Indexes().DropOne(context.TODO(), "tenant_id")
		return err
	},
}
