package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration35 = migrate.Migration{
	Version:     MigrationVersion35,
	Description: "Rename the column authenticated to confirmed",
	Up: migrate.MigrationFunc(func(_ context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   MigrationVersion35,
			"action":    "Up",
		}).Info("Applying migration")

		return renameField(db, "users", "authenticated", "confirmed")
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   MigrationVersion35,
			"action":    "Down",
		}).Info("Applying migration")

		return renameField(db, "users", "confirmed", "authenticated")
	}),
}
