package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration16 = migrate.Migration{
	Version:     16,
	Description: "Set the fingerprint as unique on public_keys collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   16,
			"action":    "Up",
		}).Info("Applying migration")
		mod := mongo.IndexModel{
			Keys:    bson.D{{"fingerprint", 1}},
			Options: options.Index().SetName("fingerprint").SetUnique(true),
		}
		_, err := db.Collection("public_keys").Indexes().CreateOne(ctx, mod)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   16,
			"action":    "Down",
		}).Info("Applying migration")
		_, err := db.Collection("public_keys").Indexes().DropOne(ctx, "fingerprint")

		return err
	}),
}
