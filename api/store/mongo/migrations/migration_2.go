package migrations

import (
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration2 = migrate.Migration{
	Version:     2,
	Description: "Rename the column device to device_uid",
	Up: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   2,
			"action":    "Up",
		}).Info("Applying migration")

		return renameField(db, "sessions", "device", "device_uid")
	},
	Down: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   2,
			"action":    "Down",
		}).Info("Applying migration")

		return renameField(db, "sessions", "device_uid", "device")
	},
}
