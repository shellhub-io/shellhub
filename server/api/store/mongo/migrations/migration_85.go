package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration85 = migrate.Migration{
	Version:     85,
	Description: "create index for tunnels address",
	Up: migrate.MigrationFunc(func(_ context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   85,
			"action":    "Up",
		}).Info("Applying migration up")
		name := "address"
		if _, err := db.Collection("tunnels").Indexes().CreateOne(context.Background(), mongo.IndexModel{
			Keys: bson.M{
				"address": 1,
			},
			Options: &options.IndexOptions{ //nolint:exhaustruct
				Name: &name,
			},
		}); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   85,
			"action":    "Down",
		}).Info("Applying migration down")
		if _, err := db.Collection("tunnels").Indexes().DropOne(context.Background(), "address"); err != nil {
			return err
		}

		return nil
	}),
}
