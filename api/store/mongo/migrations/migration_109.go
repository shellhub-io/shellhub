package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration109 = migrate.Migration{
	Version:     109,
	Description: "Add indexes to tags collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   109,
			"action":    "Up",
		}).Info("Applying migration up")

		_, err := db.Collection("tags").Indexes().CreateMany(ctx, []mongo.IndexModel{
			{
				Keys:    bson.D{{Key: "tenant_id", Value: 1}, {Key: "name", Value: 1}},
				Options: options.Index().SetUnique(true).SetName("idx_tenant_id_name_unique"),
			},
			{
				Keys:    bson.D{{Key: "tenant_id", Value: 1}},
				Options: options.Index().SetName("idx_tenant_id"),
			},
		})

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   109,
			"action":    "Down",
		}).Info("Reverting migration down")

		if _, err := db.Collection("tags").Indexes().DropOne(ctx, "idx_tenant_id_name_unique"); err != nil {
			return err
		}

		if _, err := db.Collection("tags").Indexes().DropOne(ctx, "idx_tenant_id"); err != nil {
			return err
		}

		return nil
	}),
}
