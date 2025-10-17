package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration112 = migrate.Migration{
	Version:     112,
	Description: "Add super_admin field to users collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   112,
			"action":    "Up",
		}).Info("Applying migration up: Adding super_admin field to all users")

		// Add super_admin=false to all existing users
		_, err := db.Collection("users").UpdateMany(
			ctx,
			bson.M{},
			bson.M{"$set": bson.M{"super_admin": false}},
		)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   112,
			"action":    "Down",
		}).Info("Reverting migration down: Removing super_admin field from all users")

		// Remove the super_admin field
		_, err := db.Collection("users").UpdateMany(
			ctx,
			bson.M{},
			bson.M{"$unset": bson.M{"super_admin": ""}},
		)

		return err
	}),
}
