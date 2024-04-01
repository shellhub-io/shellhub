package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration33 = migrate.Migration{
	Version:     33,
	Description: "add tags field to collection devices",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   33,
			"action":    "Up",
		}).Info("Applying migration")
		mod := mongo.IndexModel{
			Keys:    bson.D{{"tags", 1}},
			Options: options.Index().SetName("tags").SetUnique(false),
		}
		_, err := db.Collection("devices").Indexes().CreateOne(ctx, mod)
		if err != nil {
			return err
		}

		if _, err := db.Collection("devices").UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"tags": []string{}}}); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   33,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}
