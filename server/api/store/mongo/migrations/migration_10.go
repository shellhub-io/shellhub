package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration10 = migrate.Migration{
	Version:     10,
	Description: "Unset unique on session_record in the users collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   10,
			"action":    "Up",
		}).Info("Applying migration")
		mod := mongo.IndexModel{
			Keys:    bson.D{{"session_record", 1}},
			Options: options.Index().SetName("session_record").SetUnique(false),
		}
		if _, err := db.Collection("users").Indexes().CreateOne(ctx, mod); err != nil {
			return err
		}
		_, err := db.Collection("users").UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"session_record": true}})

		return err
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   10,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}
