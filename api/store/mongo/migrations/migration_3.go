package migrations

import (
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration3 = migrate.Migration{
	Version:     3,
	Description: "Rename the column attributes to info",
	Up: func(db *mongo.Database) error {
		logrus.Info("Applying migration 3 - Up")

		return renameField(db, "devices", "attributes", "info")
	},
	Down: func(db *mongo.Database) error {
		logrus.Info("Applying migration 3 - Down")

		return renameField(db, "devices", "info", "attributes")
	},
}
