package migrations

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration65 = migrate.Migration{
	Version:     65,
	Description: "Adding the 'connected_at' attribute to the device if it does not already exist.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   65,
			"action":    "Up",
		}).Info("Applying migration")

		filter := bson.M{
			"connected_at":    bson.M{"$in": []interface{}{nil}},
			"disconnected_at": bson.M{"$in": []interface{}{nil}},
		}

		update := bson.M{
			"$set": bson.M{
				"connected_at":    time.Time{},
				"disconnected_at": time.Time{},
			},
		}

		_, err := db.Collection("devices").UpdateMany(ctx, filter, update)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   65,
			"action":    "Down",
		}).Info("Reverting migration")

		unset := bson.M{
			"$unset": bson.M{
				"connected_at": 1,
			},
		}

		_, err := db.Collection("devices").UpdateMany(ctx, bson.M{}, unset)

		return err
	}),
}
