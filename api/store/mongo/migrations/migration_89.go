package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration89 = migrate.Migration{
	Version:     MigrationVersion89,
	Description: "Adding an external ID attribute to users collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   MigrationVersion89,
			"action":    "Up",
		}).Info("Applying migration")

		filter := bson.M{
			"external_id": bson.M{"$exists": false},
		}

		update := bson.M{
			"$set": bson.M{
				"external_id": "",
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
			"version":   MigrationVersion89,
			"action":    "Down",
		}).Info("Reverting migration")

		filter := bson.M{
			"external_id": bson.M{"$exists": true},
		}

		update := bson.M{
			"$unset": bson.M{
				"external_id": "",
			},
		}

		_, err := db.
			Collection("users").
			UpdateMany(ctx, filter, update)

		return err
	}),
}
