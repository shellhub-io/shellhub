package migrations

import (
	"context"

	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration121 = migrate.Migration{
	Version:     121,
	Description: "Increase active_sessions TTL from 30s to 60s",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{"component": "migration", "version": 121, "action": "Up"}).Info("Applying migration")

		if _, err := db.Collection("active_sessions").Indexes().DropOne(ctx, "last_seen"); err != nil {
			log.WithError(err).Error("Failed to drop old last_seen TTL index")

			return err
		}

		expireAfter := int32(60)
		if _, err := db.Collection("active_sessions").Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{{Key: "last_seen", Value: 1}},
			Options: options.Index().
				SetName("last_seen").
				SetExpireAfterSeconds(expireAfter),
		}); err != nil {
			log.WithError(err).Error("Failed to recreate last_seen TTL index")

			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{"component": "migration", "version": 121, "action": "Down"}).Info("Reverting migration")

		if _, err := db.Collection("active_sessions").Indexes().DropOne(ctx, "last_seen"); err != nil {
			log.WithError(err).Error("Failed to drop last_seen TTL index")

			return err
		}

		expireAfter := int32(30)
		if _, err := db.Collection("active_sessions").Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{{Key: "last_seen", Value: 1}},
			Options: options.Index().
				SetName("last_seen").
				SetExpireAfterSeconds(expireAfter),
		}); err != nil {
			log.WithError(err).Error("Failed to recreate last_seen TTL index with original value")

			return err
		}

		return nil
	}),
}
