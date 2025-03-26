package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration15 = migrate.Migration{
	Version:     15,
	Description: "Set all names to lowercase in the namespaces",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   15,
			"action":    "Up",
		}).Info("Applying migration")
		_, err := db.Collection("namespaces").UpdateMany(ctx, bson.M{}, []bson.M{
			{
				"$set": bson.M{
					"name": bson.M{"$toLower": "$name"},
				},
			},
		})

		return err
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   15,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}
