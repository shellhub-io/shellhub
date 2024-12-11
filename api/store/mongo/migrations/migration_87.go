package migrations

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration87 = migrate.Migration{
	Version:     87,
	Description: "Adding an 'auth_methods' attributes to user collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   87,
			"action":    "Up",
		}).Info("Applying migration")

		filter := bson.M{
			"preferences.auth_methods": bson.M{"$exists": false},
		}

		update := bson.M{
			"$set": bson.M{
				"preferences.auth_methods": []string{models.UserAuthMethodLocal.String()},
			},
		}

		_, err := db.
			Collection("users").
			UpdateMany(ctx, filter, update)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   87,
			"action":    "Down",
		}).Info("Reverting migration")

		filter := bson.M{
			"preferences.auth_methods": bson.M{"$exists": true},
		}

		update := bson.M{
			"$unset": bson.M{
				"preferences.auth_methods": "",
			},
		}

		_, err := db.
			Collection("users").
			UpdateMany(ctx, filter, update)

		return err
	}),
}
