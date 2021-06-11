package migrations

import (
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration4 = migrate.Migration{
	Version: 4,
	Up: func(db *mongo.Database) error {
		logrus.Info("Applying migration 4 - Up")

		return renameField(db, "devices", "version", "info.version")
	},
	Down: func(db *mongo.Database) error {
		logrus.Info("Applying migration 4 - Down")

		return renameField(db, "devices", "info.version", "version")
	},
}
