package migrations

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/envs"
	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration79 = migrate.Migration{
	Version:     79,
	Description: "create and populate the system collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithField("component", "migration").
			WithField("version", 79).
			WithField("action", " Up").
			Info("Applying migration")

		if err := db.CreateCollection(ctx, "system"); err != nil {
			return err
		}

		if envs.IsCommunity() {
			users, err := db.Collection("users").CountDocuments(ctx, bson.M{})
			if err != nil {
				return err
			}

			if _, err := db.Collection("system").InsertOne(ctx, bson.M{
				"setup": users > 0,
			}); err != nil {
				return err
			}
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithField("component", "migration").
			WithField("version", 79).
			WithField("action", "Down").
			Info("Reverting migration")

		return db.Collection("system").Drop(ctx)
	}),
}
