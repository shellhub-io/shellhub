package migrations

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/envs"
	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration117 = migrate.Migration{
	Version:     117,
	Description: "Set setup field in system collection for non-cloud environments",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   117,
			"action":    "Up",
		}).Info("Applying migration up")

		if envs.IsCloud() {
			return nil
		}

		usersCount, err := db.Collection("users").CountDocuments(ctx, bson.M{})
		if err != nil {
			return err
		}

		_, err = db.Collection("system").UpdateOne(ctx, bson.M{}, bson.M{"$set": bson.M{"setup": usersCount > 0}})

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   117,
			"action":    "Down",
		}).Info("Applying migration down")

		log.Info("Unable to undo setup field changes")

		return nil
	}),
}
