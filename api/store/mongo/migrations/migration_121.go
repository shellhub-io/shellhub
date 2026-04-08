package migrations

import (
	"context"

	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration121 = migrate.Migration{
	Version:     121,
	Description: "Add disable_password and disable_public_key to namespace settings",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{"component": "migration", "version": 121, "action": "Up"}).Info("Applying migration")

		if _, err := db.Collection("namespaces").UpdateMany(ctx, bson.M{}, bson.M{
			"$set": bson.M{
				"settings.disable_password":   false,
				"settings.disable_public_key": false,
			},
		}); err != nil {
			log.WithError(err).Error("Failed to add disable_password and disable_public_key to namespace settings")

			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{"component": "migration", "version": 121, "action": "Down"}).Info("Reverting migration")

		if _, err := db.Collection("namespaces").UpdateMany(ctx, bson.M{}, bson.M{
			"$unset": bson.M{
				"settings.disable_password":   "",
				"settings.disable_public_key": "",
			},
		}); err != nil {
			log.WithError(err).Error("Failed to remove disable_password and disable_public_key from namespace settings")

			return err
		}

		return nil
	}),
}
