package migrations

import (
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration_4 = migrate.Migration{
	Version: 4,
	Up: func(db *mongo.Database) error {
		return renameField(db, "devices", "version", "info.version")
	},
	Down: func(db *mongo.Database) error {
		return renameField(db, "devices", "info.version", "version")
	},
}
