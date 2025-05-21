package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration3 = migrate.Migration{
	Version:     3,
	Description: "Rename the column attributes to info",
	Up: migrate.MigrationFunc(func(_ context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   3,
			"action":    "Up",
		}).Info("Applying migration")

		return renameField(db, "devices", "attributes", "info")
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   3,
			"action":    "Down",
		}).Info("Applying migration")

		return renameField(db, "devices", "info", "attributes")
	}),
}
