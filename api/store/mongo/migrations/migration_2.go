package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration2 = migrate.Migration{
	Version:     2,
	Description: "Rename the column device to device_uid",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   2,
			"action":    "Up",
		}).Info("Applying migration")

		return renameField(db, "sessions", "device", "device_uid")
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   2,
			"action":    "Down",
		}).Info("Applying migration")

		return renameField(db, "sessions", "device_uid", "device")
	}),
}
