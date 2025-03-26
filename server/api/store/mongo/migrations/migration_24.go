package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration24 = migrate.Migration{
	Version:     24,
	Description: "convert names and emails to lowercase",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   24,
			"action":    "Up",
		}).Info("Applying migration")
		if _, err := db.Collection("users").UpdateMany(ctx, bson.D{}, []bson.M{
			{
				"$set": bson.M{
					"username": bson.M{"$toLower": "$username"},
					"email":    bson.M{"$toLower": "$email"},
				},
			},
		}); err != nil {
			return err
		}

		_, err := db.Collection("namespaces").UpdateMany(ctx, bson.D{}, []bson.M{
			{
				"$set": bson.M{"name": bson.M{"$toLower": "$name"}},
			},
		})

		return err
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   24,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}
