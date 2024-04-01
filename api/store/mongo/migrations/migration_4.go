package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration4 = migrate.Migration{
	Version:     4,
	Description: "Rename the column version to info.version",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   4,
			"action":    "Up",
		}).Info("Applying migration")

		return renameField(db, "devices", "version", "info.version")
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   4,
			"action":    "Down",
		}).Info("Applying migration")

		return renameField(db, "devices", "info.version", "version")
	}),
}
