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
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   51,
			"action":    "Up",
		}).Info("Applying migration up")
		Name := "name"

		if _, err := db.Collection("devices").Indexes().CreateOne(context.Background(), mongo.IndexModel{
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
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   51,
			"action":    "Down",
		}).Info("Applying migration down")
		Name := "name"

		if _, err := db.Collection("devices").Indexes().DropOne(context.Background(), Name); err != nil {
			return err
		}

		return nil
	}),
}
