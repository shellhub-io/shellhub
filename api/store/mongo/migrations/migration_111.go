package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration111 = migrate.Migration{
	Version:     111,
	Description: "Add removed_at field with null value to all devices",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   111,
			"action":    "Up",
		}).Info("Applying migration up")

		_, err := db.Collection("devices").UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"removed_at": nil}})

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   111,
			"action":    "Down",
		}).Info("Reverting migration down")

		_, err := db.Collection("devices").UpdateMany(ctx, bson.M{}, bson.M{"$unset": bson.M{"removed_at": ""}})

		return err
	}),
}
