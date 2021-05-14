package migrations

import (
	"context"

	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration_15 = migrate.Migration{
	Version: 15,
	Up: func(db *mongo.Database) error {
		_, err := db.Collection("namespaces").UpdateMany(context.TODO(), bson.M{}, []bson.M{
			{
				"$set": bson.M{
					"name": bson.M{"$toLower": "$name"},
				},
			},
		})
		return err
	},
	Down: func(db *mongo.Database) error {
		return nil
	},
}
