package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration41 = migrate.Migration{
	Version:     MigrationVersion41,
	Description: "update online index from devices collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   MigrationVersion41,
			"action":    "Up",
		}).Info("Applying migration")

		if _, err := db.Collection("connected_devices").Indexes().DropOne(ctx, "last_seen"); err != nil {
			return err
		}

		mod := mongo.IndexModel{
			Keys:    bson.D{{"last_seen", 1}},
			Options: options.Index().SetName("last_seen").SetExpireAfterSeconds(MigrationTTL120),
		}
		if _, err := db.Collection("connected_devices").Indexes().CreateOne(ctx, mod); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   MigrationVersion41,
			"action":    "Down",
		}).Info("Applying migration")

		if _, err := db.Collection("connected_devices").Indexes().DropOne(ctx, "last_seen"); err != nil {
			return err
		}

		mod := mongo.IndexModel{
			Keys:    bson.D{{"last_seen", 1}},
			Options: options.Index().SetName("last_seen").SetExpireAfterSeconds(MigrationTTL60),
		}
		if _, err := db.Collection("connected_devices").Indexes().CreateOne(ctx, mod); err != nil {
			return err
		}

		return nil
	}),
}
