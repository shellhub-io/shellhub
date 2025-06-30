package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration99 = migrate.Migration{
	Version:     99,
	Description: "Update session.recorded to false when session has no events",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   99,
			"action":    "Up",
		}).Info("Applying migration")

		_, err := db.Collection("sessions").
			UpdateMany(
				ctx,
				bson.M{
					"recorded":     true,
					"events.types": bson.M{"$size": 0},
				},
				bson.M{"$set": bson.M{"recorded": false}},
			)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   99,
			"action":    "Down",
		}).Info("Reverting migration")

		// NOTE: This migration shouldn't be reverted.

		return nil
	}),
}
