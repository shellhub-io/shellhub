package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration34 = migrate.Migration{
	Version:     34,
	Description: "create online index in devices collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   34,
			"action":    "Up",
		}).Info("Applying migration")

		indexModel := mongo.IndexModel{
			Keys: bson.D{{"online", 1}},
		}

		_, err := db.Collection("devices").Indexes().CreateOne(ctx, indexModel)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   34,
			"action":    "Down",
		}).Info("Applying migration")

		_, err := db.Collection("devices").Indexes().DropOne(ctx, "online")

		return err
	}),
}
