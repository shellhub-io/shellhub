package mongo

import (
	"github.com/shellhub-io/shellhub/api/store/mongo/migrations"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/mongo"
)

func ApplyMigrations(db *mongo.Database) error {
	return migrate.NewMigrate(db, migrations.GenerateMigrations()...).Up(migrate.AllAvailable)
}
