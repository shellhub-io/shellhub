package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration46 = migrate.Migration{
	Version:     46,
	Description: "change public keys with empty username in favor of .* regexp",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   46,
			"action":    "Up",
		}).Info("Applying migration")

		_, err := db.Collection("public_keys").Aggregate(context.Background(),
			mongo.Pipeline{
				{
					{"$match", bson.M{"username": ""}},
				},
				{
					{"$set", bson.M{"username": ".*"}},
				},
				{
					{"$merge", bson.M{"into": "public_keys", "whenMatched": "replace"}},
				},
			},
		)
		if err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   46,
			"action":    "Down",
		}).Info("Applying migration")

		_, err := db.Collection("public_keys").Aggregate(context.Background(),
			mongo.Pipeline{
				{
					{"$match", bson.M{"username": ".*"}},
				},
				{
					{"$set", bson.M{"username": ""}},
				},
				{
					{"$merge", bson.M{"into": "public_keys", "whenMatched": "replace"}},
				},
			},
		)
		if err != nil {
			return err
		}

		return nil
	}),
}
