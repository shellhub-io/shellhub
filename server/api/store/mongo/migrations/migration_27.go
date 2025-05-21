package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration27 = migrate.Migration{
	Version:     27,
	Description: "Set closed field in the sessions",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   27,
			"action":    "Up",
		}).Info("Applying migration")
		_, err := db.Collection("sessions").UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"closed": true}})

		return err
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   27,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}
