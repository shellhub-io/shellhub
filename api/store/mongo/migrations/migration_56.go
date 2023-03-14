package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration56 = migrate.Migration{
	Version:     56,
	Description: "create index for public url address on devices",
	Up: func(database *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   56,
			"action":    "Up",
		}).Info("Applying migration up")
		field := "public_url_address"
		unique := true

		if _, err := database.Collection("devices").Indexes().CreateOne(context.Background(), mongo.IndexModel{
			Keys: bson.M{
				field: 1,
			},
			Options: &options.IndexOptions{ //nolint:exhaustruct
				Unique: &unique,
				Name:   &field,
			},
		}); err != nil {
			return err
		}

		return nil
	},
	Down: func(database *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   56,
			"action":    "Down",
		}).Info("Applying migration down")
		field := "public_url_address"

		if _, err := database.Collection("devices").Indexes().DropOne(context.Background(), field); err != nil {
			return err
		}

		return nil
	},
}
