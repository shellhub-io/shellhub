package migrations

import (
	"context"

	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration_5 = migrate.Migration{
	Version: 5,
	Up: func(db *mongo.Database) error {
		mod := mongo.IndexModel{
			Keys:    bson.D{{"email", 1}},
			Options: options.Index().SetName("email").SetUnique(true),
		}
		_, err := db.Collection("users").Indexes().CreateOne(context.TODO(), mod)

		return err
	},
	Down: func(db *mongo.Database) error {
		_, err := db.Collection("users").Indexes().DropOne(context.TODO(), "email")

		return err
	},
}
