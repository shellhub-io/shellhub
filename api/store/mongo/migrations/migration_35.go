package migrations

import (
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration35 = migrate.Migration{
	Version:     35,
	Description: "Rename the column authenticated to confirmed",
	Up: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   35,
			"action":    "Up",
		}).Info("Applying migration")

		return renameField(db, "users", "authenticated", "confirmed")
	},
	Down: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   35,
			"action":    "Down",
		}).Info("Applying migration")

		return renameField(db, "users", "confirmed", "authenticated")
	},
}
