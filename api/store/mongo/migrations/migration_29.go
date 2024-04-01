package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration29 = migrate.Migration{
	Version:     29,
	Description: "add last_login field to collection users",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   29,
			"action":    "Up",
		}).Info("Applying migration")
		if _, err := db.Collection("users").UpdateMany(context.TODO(), bson.M{}, bson.M{"$set": bson.M{"last_login": nil}}); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   29,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}
