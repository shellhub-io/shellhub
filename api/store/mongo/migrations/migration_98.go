package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration98 = migrate.Migration{
	Version:     MigrationVersion98,
	Description: "Convert the username's to nil when it's a blank string",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   MigrationVersion98,
			"action":    "Up",
		}).Info("Applying migration")

		_, err := db.Collection("users").UpdateMany(ctx, bson.M{"username": ""}, bson.M{"$set": bson.M{"username": nil}})

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   MigrationVersion98,
			"action":    "Down",
		}).Info("Cannot undo migration")

		_, err := db.Collection("users").UpdateMany(ctx, bson.M{"username": nil}, bson.M{"$set": bson.M{"username": ""}})

		return err
	}),
}
