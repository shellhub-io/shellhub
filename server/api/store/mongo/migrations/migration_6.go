package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration6 = migrate.Migration{
	Version:     6,
	Description: "Unset unique on status in the devices collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   6,
			"action":    "Up",
		}).Info("Applying migration")
		mod := mongo.IndexModel{
			Keys:    bson.D{{"status", 1}},
			Options: options.Index().SetName("status").SetUnique(false),
		}
		if _, err := db.Collection("devices").Indexes().CreateOne(ctx, mod); err != nil {
			return err
		}
		_, err := db.Collection("devices").UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"status": "accepted"}})

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   6,
			"action":    "Down",
		}).Info("Applying migration")
		if _, err := db.Collection("devices").UpdateMany(ctx, bson.M{}, bson.M{"$unset": bson.M{"status": ""}}); err != nil {
			return err
		}
		_, err := db.Collection("status").Indexes().DropOne(ctx, "status")

		return err
	}),
}
