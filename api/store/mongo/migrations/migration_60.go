package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration60 = migrate.Migration{
	Version:     60,
	Description: "create index for tenant_id on active_sessions",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   60,
			"action":    "Up",
		}).Info("Applying migration up")
		indexName := "tenant_id"
		if _, err := db.Collection("active_sessions").Indexes().CreateOne(context.Background(), mongo.IndexModel{
			Keys: bson.M{
				"tenant_id": 1,
			},
			Options: &options.IndexOptions{ //nolint:exhaustruct
				Name: &indexName,
			},
		}); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   60,
			"action":    "Down",
		}).Info("Applying migration down")
		if _, err := db.Collection("active_sessions").Indexes().DropOne(context.Background(), "tenant_id"); err != nil {
			return err
		}

		return nil
	}),
}
