package migrations

import (
	"context"

	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration105 = migrate.Migration{
	Version:     105,
	Description: "Drop removed_devices collection as it's no longer needed",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{"component": "migration", "version": 105, "action": "Up"}).Info("Applying migration")

		collections, err := db.ListCollectionNames(ctx, bson.M{"name": "removed_devices"})
		if err != nil {
			log.WithError(err).Error("Failed to list collections")

			return err
		}

		if len(collections) == 0 {
			log.Info("Collection removed_devices does not exist, skipping drop")

			return nil
		}

		if err := db.Collection("removed_devices").Drop(ctx); err != nil {
			log.WithError(err).Error("Failed to drop removed_devices collection")

			return err
		}

		log.WithFields(log.Fields{"collection": "removed_devices"}).Info("Successfully dropped removed_devices collection")

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{"component": "migration", "version": 105, "action": "Down"}).Info("Cannot revert migration")

		return nil
	}),
}
