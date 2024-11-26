package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration83 = migrate.Migration{
	Version:     83,
	Description: "Creating Tag collection.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   83,
			"action":    "Up",
		}).Info("Applying migration")

		if err := db.CreateCollection(ctx, "tags"); err != nil {
			return err
		}

		index := mongo.IndexModel{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetName("name").SetUnique(false),
		}

		_, err := db.Collection("tags").Indexes().CreateOne(ctx, index)
		if err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   83,
			"action":    "Down",
		}).Info("Reverting migration")

		_, err := db.Collection("tags").Indexes().DropOne(ctx, "names")
		if err != nil {
			return err
		}

		return db.Collection("tags").Drop(ctx)
	}),
}
