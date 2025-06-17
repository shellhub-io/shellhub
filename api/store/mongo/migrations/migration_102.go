package migrations

import (
	"context"

	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration102 = migrate.Migration{
	Version:     102,
	Description: "Remove legacy devices_count field from namespaces in favor of status-specific counters",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{"component": "migration", "version": 102, "action": "Up"}).Info("Applying migration")

		r, err := db.Collection("namespaces").UpdateMany(ctx, bson.M{}, bson.M{"$unset": bson.M{"devices_count": ""}})
		if err != nil {
			log.WithError(err).Error("Failed to remove device count fields")

			return err
		}

		log.WithFields(log.Fields{"modified_count": r.ModifiedCount}).Info("Removed device count fields from namespaces")

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{"component": "migration", "version": 102, "action": "Down"}).Info("Cannot revert migration")

		return nil
	}),
}
