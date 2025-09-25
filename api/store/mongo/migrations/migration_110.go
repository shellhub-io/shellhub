package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration110 = migrate.Migration{
	Version:     110,
	Description: "Remove all devices with status=removed",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   110,
			"action":    "Up",
		}).Info("Applying migration up")

		_, err := db.Collection("devices").DeleteMany(ctx, bson.M{"status": "removed"})

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   110,
			"action":    "Down",
		}).Warn("Nothing to do on down migration (cannot restore deleted devices)")

		return nil
	}),
}
