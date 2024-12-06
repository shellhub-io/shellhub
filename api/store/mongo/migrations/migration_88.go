package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration88 = migrate.Migration{
	Version:     88,
	Description: "Adding an 'auth_methods' attributes to user collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   88,
			"action":    "Up",
		}).Info("Applying migration")

		filter := bson.M{
			"preferences.auth_methods": bson.M{"$exists": false},
		}

		update := bson.M{
			"$set": bson.M{
				"preferences.auth_methods": []string{"manual"},
			},
		}

		_, err := db.
			Collection("users").
			UpdateMany(ctx, filter, update)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   88,
			"action":    "Down",
		}).Info("Reverting migration")

		filter := bson.M{
			"prefenrences.auth_methods": bson.M{"$exists": true},
		}

		update := bson.M{
			"$unset": bson.M{
				"preferences.auth_methods": "",
			},
		}

		_, err := db.
			Collection("users").
			UpdateMany(ctx, filter, update)

		return err
	}),
}
