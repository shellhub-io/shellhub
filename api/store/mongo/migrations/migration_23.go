package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration23 = migrate.Migration{
	Version:     23,
	Description: "change dot in namespace name and hostname to -",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   23,
			"action":    "Up",
		}).Info("Applying migration")
		if _, err := db.Collection("namespaces").UpdateMany(ctx, bson.D{}, []bson.M{
			{
				"$set": bson.M{
					"name": bson.M{
						"$replaceAll": bson.M{"input": "$name", "find": ".", "replacement": "-"},
					},
				},
			},
		}); err != nil {
			return err
		}

		if _, err := db.Collection("devices").UpdateMany(ctx, bson.D{}, []bson.M{
			{
				"$set": bson.M{
					"name": bson.M{
						"$replaceAll": bson.M{"input": "$name", "find": ".", "replacement": "-"},
					},
				},
			},
		}); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   23,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}
