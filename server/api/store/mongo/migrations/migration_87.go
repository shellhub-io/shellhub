package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration87 = migrate.Migration{
	Version:     87,
	Description: "Adding an 'authentication' attributes to system collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   87,
			"action":    "Up",
		}).Info("Applying migration")

		if count, _ := db.Collection("system").CountDocuments(ctx, bson.M{}); count == 0 {
			if _, err := db.Collection("system").InsertOne(ctx, bson.M{"setup": true}); err != nil {
				return err
			}
		}

		filter := bson.M{
			"authentication": bson.M{"$exists": false},
		}

		update := bson.M{
			"$set": bson.M{
				"authentication": bson.M{
					"local": bson.M{
						"enabled": true,
					},
				},
			},
		}

		_, err := db.
			Collection("system").
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
			"authentication": bson.M{"$exists": true},
		}

		update := bson.M{
			"$unset": bson.M{
				"authentication": "",
			},
		}

		_, err := db.
			Collection("system").
			UpdateMany(ctx, filter, update)

		return err
	}),
}
