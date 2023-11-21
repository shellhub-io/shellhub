package migrations

import (
	"context"

	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration62 = migrate.Migration{
	Version:     62,
	Description: "create index for tenant_id on recorded_sessions",
	Up: func(database *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   62,
			"action":    "Up",
		}).Info("Applying migration up")

		indexName := "tenant_id"
		_, err := database.Collection("recorded_sessions").Indexes().CreateOne(context.Background(), mongo.IndexModel{
			Keys: bson.M{
				"tenant_id": 1,
			},
			Options: &options.IndexOptions{ //nolint:exhaustruct
				Name: &indexName,
			},
		})
		if err != nil {
			log.WithFields(log.Fields{
				"component": "migration",
				"version":   62,
				"action":    "Up",
			}).WithError(err).Info("Error while trying to apply migration 62")

			return err
		}

		log.WithFields(log.Fields{
			"component": "migration",
			"version":   62,
			"action":    "Up",
		}).Info("Succeeds to to apply migration 62")

		return nil
	},
	Down: func(database *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   62,
			"action":    "Down",
		}).Info("Applying migration down")
		if _, err := database.Collection("recorded_sessions").Indexes().DropOne(context.Background(), "tenant_id"); err != nil {
			return err
		}

		return nil
	},
}
