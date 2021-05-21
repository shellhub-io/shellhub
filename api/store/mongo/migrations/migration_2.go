package migrations

import (
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration_2 = migrate.Migration{
	Version: 2,
	Up: func(db *mongo.Database) error {
		logrus.Info("Applying migration 2 - Up")
		return renameField(db, "sessions", "device", "device_uid")
	},
	Down: func(db *mongo.Database) error {
		logrus.Info("Applying migration 2 - Down")
		return renameField(db, "sessions", "device_uid", "device")
	},
}
