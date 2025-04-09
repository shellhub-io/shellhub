package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration93 = migrate.Migration{
	Version:     93,
	Description: "remove public_url and public_url_address from device collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   93,
			"action":    "Up",
		}).Info("Applying migration")

		_, err := db.Collection("devices").UpdateMany(ctx, bson.M{}, bson.M{"$unset": bson.M{"public_url": "", "public_url_address": ""}})

		return err
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   93,
			"action":    "Down",
		}).Info("Cannot undo migration")

		return nil
	}),
}
