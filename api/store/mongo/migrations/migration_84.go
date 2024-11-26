package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration84 = migrate.Migration{
	Version:     84,
	Description: "create index for sessions' type",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   84,
			"action":    "Up",
		}).Info("Applying migration up")
		name := "events.types"
		if _, err := db.Collection("sessions").Indexes().CreateOne(context.Background(), mongo.IndexModel{
			Keys: bson.M{
				"events.types": 1,
			},
			Options: &options.IndexOptions{ //nolint:exhaustruct
				Name: &name,
			},
		}); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   84,
			"action":    "Down",
		}).Info("Applying migration down")
		if _, err := db.Collection("sessions").Indexes().DropOne(context.Background(), "events.types"); err != nil {
			return err
		}

		return nil
	}),
}
