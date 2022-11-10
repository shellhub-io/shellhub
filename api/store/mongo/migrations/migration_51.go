package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration51 = migrate.Migration{
	Version:     51,
	Description: "create index for name on devices",
	Up: func(database *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   51,
			"action":    "Up",
		}).Info("Applying migration up")
		Name := "name"

		if _, err := database.Collection("devices").Indexes().CreateOne(context.Background(), mongo.IndexModel{
			Keys: bson.M{
				Name: 1,
			},
			Options: &options.IndexOptions{ //nolint:exhaustruct
				Name: &Name,
			},
		}); err != nil {
			return err
		}

		return nil
	},
	Down: func(database *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   51,
			"action":    "Down",
		}).Info("Applying migration down")
		Name := "name"

		if _, err := database.Collection("devices").Indexes().DropOne(context.Background(), Name); err != nil {
			return err
		}

		return nil
	},
}
