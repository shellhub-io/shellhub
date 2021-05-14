package migrations

import (
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration_3 = migrate.Migration{
	Version: 3,
	Up: func(db *mongo.Database) error {
		return renameField(db, "devices", "attributes", "info")
	},
	Down: func(db *mongo.Database) error {
		return renameField(db, "devices", "info", "attributes")
	},
}
