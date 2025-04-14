package migrations

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration83 = migrate.Migration{
	Version:     83,
	Description: "Set the user's 'origin' attribute to 'manual' if it does not already exist.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   83,
			"action":    "Up",
		}).Info("Applying migration")

		filter := bson.M{
			"origin": bson.M{"$exists": false},
		}

		update := bson.M{
			"$set": bson.M{
				"origin": models.UserOriginLocal.String(),
			},
		}

		_, err := db.Collection("users").UpdateMany(ctx, filter, update)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   83,
			"action":    "Down",
		}).Info("Reverting migration")

		filter := bson.M{
			"origin": bson.M{"$exists": true},
		}

		update := bson.M{
			"$unset": bson.M{
				"origin": "",
			},
		}

		_, err := db.Collection("users").UpdateMany(ctx, filter, update)

		return err
	}),
}
