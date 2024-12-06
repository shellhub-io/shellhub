package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration86 = migrate.Migration{
	Version:     86,
	Description: "Set the user's 'origin' attribute to 'manual' if it's 'local'.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   86,
			"action":    "Up",
		}).Info("Applying migration")

		filter := bson.M{
			"origin": "local",
		}

		update := bson.M{
			"$set": bson.M{
				"origin": "manual",
			},
		}

		_, err := db.Collection("users").UpdateMany(ctx, filter, update)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   86,
			"action":    "Down",
		}).Info("Reverting migration")

		filter := bson.M{
			"origin": "manual",
		}

		update := bson.M{
			"$set": bson.M{
				"origin": "local",
			},
		}

		_, err := db.Collection("users").UpdateMany(ctx, filter, update)

		return err
	}),
}
