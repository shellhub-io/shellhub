package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration61 = migrate.Migration{
	Version:     61,
	Description: "delete devices with empty name",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   61,
			"action":    "Up",
		}).Info("Applying migration up")
		if _, err := db.Collection("devices").DeleteMany(context.Background(), bson.M{"$or": bson.A{
			bson.M{"name": ""},
			bson.M{"name": bson.M{"$exists": false}},
		}}); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		// This migration is not reversible.
		return nil
	}),
}
