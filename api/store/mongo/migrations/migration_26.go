package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration26 = migrate.Migration{
	Version:     26,
	Description: "Create collection used to recover password and activate account",
	Up: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   26,
			"action":    "Up",
		}).Info("Applying migration")
		indexModel := mongo.IndexModel{
			Keys:    bson.D{{"created_at", 1}},
			Options: options.Index().SetName("ttl").SetExpireAfterSeconds(86400),
		}
		_, err := db.Collection("recovery_tokens").Indexes().CreateOne(context.TODO(), indexModel)
		if err != nil {
			return err
		}

		indexModel = mongo.IndexModel{
			Keys:    bson.D{{"token", 1}},
			Options: options.Index().SetName("token").SetUnique(false),
		}
		if _, err := db.Collection("recovery_tokens").Indexes().CreateOne(context.TODO(), indexModel); err != nil {
			return err
		}

		indexModel = mongo.IndexModel{
			Keys:    bson.D{{"user", 1}},
			Options: options.Index().SetName("user").SetUnique(false),
		}
		if _, err := db.Collection("recovery_tokens").Indexes().CreateOne(context.TODO(), indexModel); err != nil {
			return err
		}

		return nil
	},
	Down: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   26,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	},
}
