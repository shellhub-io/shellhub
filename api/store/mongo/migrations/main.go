package migrations

import (
	"context"

	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GenerateMigrations() []migrate.Migration {
	return []migrate.Migration{
		migration_1,
		migration_2,
		migration_3,
		migration_4,
		migration_5,
		migration_6,
		migration_7,
		migration_8,
		migration_9,
		migration_10,
		migration_11,
		migration_12,
		migration_13,
		migration_14,
		migration_15,
		migration_16,
		migration_17,
		migration_18,
		migration_19,
		migration_20,
		migration_21,
		migration_22,
		migration_23,
	}
}

func renameField(db *mongo.Database, coll, from, to string) error {
	_, err := db.Collection(coll).UpdateMany(context.Background(), bson.M{}, bson.M{"$rename": bson.M{from: to}})
	return err
}
