package migrations

import (
	"context"

	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration81 = migrate.Migration{
	Version:     MigrationVersion81,
	Description: "Create a 'time' index in the 'recorded_sessions' collection.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithField("component", "migration").
			WithField("version", MigrationVersion81).
			WithField("action", " Up").
			Info("Applying migration")

		index := mongo.IndexModel{
			Keys:    bson.D{{Key: "time", Value: 1}},
			Options: options.Index().SetName("time").SetUnique(false),
		}

		_, err := db.Collection("recorded_sessions").Indexes().CreateOne(ctx, index)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithField("component", "migration").
			WithField("version", MigrationVersion81).
			WithField("action", "Down").
			Info("Applying migration")

		_, err := db.Collection("recorded_sessions").Indexes().DropOne(ctx, "time")

		return err
	}),
}
