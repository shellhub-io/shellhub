package migrations

import (
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration1 = migrate.Migration{
	Version:     1,
	Description: "Create the database for the system",
	Up: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   1,
			"action":    "Up",
		}).Info("Applying migration")

		return nil
	},
	Down: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   1,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	},
}
