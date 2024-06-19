package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration70 = migrate.Migration{
	Version:     70,
	Description: "Adding the 'preferences' attribute to the user if it does not already exist.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   70,
			"action":    "Up",
		}).Info("Applying migration")

		filter := bson.M{
			"preferences": bson.M{"$exists": false},
		}

		update := bson.M{
			"$set": bson.M{
				"preferences": bson.M{},
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
			"version":   70,
			"action":    "Down",
		}).Info("Reverting migration")

		filter := bson.M{
			"preferences": bson.M{"$exists": true},
		}

		update := bson.M{
			"$unset": bson.M{
				"preferences": "",
			},
		}

		_, err := db.
			Collection("users").
			UpdateMany(ctx, filter, update)

		return err
	}),
}
