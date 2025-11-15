package migrations

import (
	"context"

	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration116 = migrate.Migration{
	Version:     116,
	Description: "Create unique index on email field in user_invitations collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   116,
			"action":    "Up",
		}).Info("Applying migration up")

		indexModel := mongo.IndexModel{Keys: bson.M{"email": 1}, Options: options.Index().SetName("email").SetUnique(true)}
		if _, err := db.Collection("user_invitations").Indexes().CreateOne(ctx, indexModel); err != nil {
			log.WithError(err).Error("Failed to create email index on user_invitations collection")

			return err
		}

		log.Info("Successfully created unique email index on user_invitations collection")

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   116,
			"action":    "Down",
		}).Info("Applying migration down")

		if _, err := db.Collection("user_invitations").Indexes().DropOne(ctx, "email"); err != nil {
			log.WithError(err).Error("Failed to drop email index from user_invitations collection")

			return err
		}

		log.Info("Successfully dropped email index from user_invitations collection")

		return nil
	}),
}
