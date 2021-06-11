package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration_26 = migrate.Migration{
	Version:     26,
	Description: "Create collection used to recover password and activate account",
	Up: func(db *mongo.Database) error {
		logrus.Info("Applying migration 26 - Up")
		index_model := mongo.IndexModel{
			Keys:    bson.D{{"created_at", 1}},
			Options: options.Index().SetName("ttl").SetExpireAfterSeconds(86400),
		}
		_, err := db.Collection("recovery_tokens").Indexes().CreateOne(context.TODO(), index_model)
		if err != nil {
			return err
		}

		index_model = mongo.IndexModel{
			Keys:    bson.D{{"token", 1}},
			Options: options.Index().SetName("token").SetUnique(false),
		}
		if _, err := db.Collection("recovery_tokens").Indexes().CreateOne(context.TODO(), index_model); err != nil {
			return err
		}

		index_model = mongo.IndexModel{
			Keys:    bson.D{{"user", 1}},
			Options: options.Index().SetName("user").SetUnique(false),
		}
		if _, err := db.Collection("recovery_tokens").Indexes().CreateOne(context.TODO(), index_model); err != nil {
			return err
		}

		return nil
	},
	Down: func(db *mongo.Database) error {
		logrus.Info("Applying migration 26 - Down")
		return nil
	},
}
