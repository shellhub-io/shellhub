package migrations

import (
	"context"

	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration106 = migrate.Migration{
	Version:     106,
	Description: "Add performance indexes to devices collection for cleanup and status filtering",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{"component": "migration", "version": 106, "action": "Up"}).Info("Applying migration")

		index := mongo.IndexModel{
			Keys:    bson.D{{Key: "status", Value: 1}, {Key: "status_updated_at", Value: 1}},
			Options: options.Index().SetName("idx_status_status_updated_at"),
		}

		indexName, err := db.Collection("devices").Indexes().CreateOne(ctx, index)
		if err != nil {
			log.WithError(err).Error("Failed to create indexes on devices collection")

			return err
		}

		log.WithFields(log.Fields{"collection": "devices", "index": indexName}).
			Info("Successfully created indexes on devices collection")

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{"component": "migration", "version": 106, "action": "Down"}).Info("Reverting migration")

		if _, err := db.Collection("devices").Indexes().DropOne(ctx, "idx_status_status_updated_at"); err != nil {
			log.WithFields(log.Fields{"index": "idx_status_status_updated_at", "error": err}).Error("Failed to drop index (may not exist)")

			return err
		}

		log.WithField("index", "idx_status_status_updated_at").Info("Successfully dropped index")

		return nil
	}),
}
