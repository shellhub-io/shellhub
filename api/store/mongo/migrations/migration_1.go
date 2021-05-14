package migrations

import (
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration_1 = migrate.Migration{
	Version: 1,
	Up: func(db *mongo.Database) error {
		return nil
	},
	Down: func(db *mongo.Database) error {
		return nil
	},
}
