package migrations

import (
	"context"

	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration77 = migrate.Migration{
	Version:     77,
	Description: "Recreate the unique index on the 'username' field in the 'users' collection with a partial filter for documents where the 'username' field is a string.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithField("component", "migration").
			WithField("version", 77).
			WithField("action", " Up").
			Info("Applying migration")

		_, _ = db.Collection("users").Indexes().DropOne(ctx, "username")

		indexModel := mongo.IndexModel{
			Keys:    bson.M{"username": 1},
			Options: options.Index().SetName("username").SetUnique(true).SetPartialFilterExpression(bson.M{"username": bson.M{"$type": "string"}}),
		}

		_, err := db.Collection("users").Indexes().CreateOne(ctx, indexModel)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithField("component", "migration").
			WithField("version", 77).
			WithField("action", "Down").
			Info("Reverting migration")

		_, err := db.Collection("users").Indexes().DropOne(ctx, "username")

		return err
	}),
}
