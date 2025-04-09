package migrations

import (
	"context"

	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration80 = migrate.Migration{
	Version:     80,
	Description: "Remove the 'message' index from the 'recorded_sessions' collection.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithField("component", "migration").
			WithField("version", 80).
			WithField("action", " Up").
			Info("Applying migration")

		_, err := db.Collection("recorded_sessions").Indexes().DropOne(ctx, "message")

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithField("component", "migration").
			WithField("version", 80).
			WithField("action", "Down").
			Info("Applying migration")

		index := mongo.IndexModel{
			Keys:    bson.D{{Key: "message", Value: 1}},
			Options: options.Index().SetName("message").SetUnique(false),
		}

		_, err := db.Collection("recorded_sessions").Indexes().CreateOne(ctx, index)

		return err
	}),
}
