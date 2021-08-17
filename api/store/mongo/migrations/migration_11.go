package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration11 = migrate.Migration{
	Version:     11,
	Description: "Create a ttl for the private_keys collection",
	Up: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   11,
			"action":    "Up",
		}).Info("Applying migration")
		mod := mongo.IndexModel{
			Keys:    bson.D{{"created_at", 1}},
			Options: options.Index().SetName("ttl").SetExpireAfterSeconds(60),
		}
		_, err := db.Collection("private_keys").Indexes().CreateOne(context.TODO(), mod)
		if err != nil {
			return err
		}

		return nil
	},
	Down: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   11,
			"action":    "Down",
		}).Info("Applying migration")
		_, err := db.Collection("private_keys").Indexes().DropOne(context.TODO(), "ttl")

		return err
	},
}
