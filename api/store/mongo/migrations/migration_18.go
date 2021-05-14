package migrations

import (
	"context"
	"os"

	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration_18 = migrate.Migration{
	Version: 18,
	Up: func(db *mongo.Database) error {
		if os.Getenv("SHELLHUB_ENTERPRISE") == "true" {
			_, err := db.Collection("namespaces").UpdateMany(context.TODO(), bson.M{}, bson.M{"$set": bson.M{"max_devices": 3}})
			return err
		}
		_, err := db.Collection("namespaces").UpdateMany(context.TODO(), bson.M{}, bson.M{"$set": bson.M{"max_devices": -1}})

		return err
	},
	Down: func(db *mongo.Database) error {
		return nil
	},
}
