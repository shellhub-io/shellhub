package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration32 = migrate.Migration{
	Version:     32,
	Description: "add authenticated field to collection users",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   32,
			"action":    "Up",
		}).Info("Applying migration")
		if _, err := db.Collection("users").UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"authenticated": true}}); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   32,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}
