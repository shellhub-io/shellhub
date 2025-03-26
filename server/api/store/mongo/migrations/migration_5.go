package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration5 = migrate.Migration{
	Version:     5,
	Description: "Set the email as unique on users collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   5,
			"action":    "Up",
		}).Info("Applying migration")
		mod := mongo.IndexModel{
			Keys:    bson.D{{"email", 1}},
			Options: options.Index().SetName("email").SetUnique(true),
		}
		_, err := db.Collection("users").Indexes().CreateOne(ctx, mod)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   5,
			"action":    "Down",
		}).Info("Applying migration")
		_, err := db.Collection("users").Indexes().DropOne(ctx, "email")

		return err
	}),
}
