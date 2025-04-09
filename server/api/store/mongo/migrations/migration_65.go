package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration65 = migrate.Migration{
	Version:     65,
	Description: "Adding the 'recovery_email' attribute to the user if it does not already exist.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   65,
			"action":    "Up",
		}).Info("Applying migration")

		filter := bson.M{
			"recovery_email": bson.M{"$in": []interface{}{nil, ""}},
		}

		update := bson.M{
			"$set": bson.M{
				"recovery_email": "",
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
			"version":   65,
			"action":    "Down",
		}).Info("Reverting migration")

		filter := bson.M{
			"_id": bson.M{"$ne": nil},
		}

		update := bson.M{
			"$unset": bson.M{
				"recovery_email": "",
			},
		}

		_, err := db.
			Collection("users").
			UpdateMany(ctx, filter, update)

		return err
	}),
}
