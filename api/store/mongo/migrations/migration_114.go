package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration114 = migrate.Migration{
	Version:     114,
	Description: "Add admin field to users collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   114,
			"action":    "Up",
		}).Info("Applying migration up: Adding super_admin field to all users")

		_, err := db.Collection("users").UpdateMany(
			ctx,
			bson.M{},
			bson.M{"$set": bson.M{"admin": false}},
		)

		return err
	}),

	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   114,
			"action":    "Down",
		}).Info("Reverting migration down: Removing super_admin field from all users")

		_, err := db.Collection("users").UpdateMany(
			ctx,
			bson.M{},
			bson.M{"$unset": bson.M{"admin": ""}},
		)

		return err
	}),
}
