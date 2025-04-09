package migrations

import (
	"context"

	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration68 = migrate.Migration{
	Version:     68,
	Description: "Rename `api_keys.user_id` to `api_keys.created_by`.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.
			WithFields(log.Fields{
				"component": "migration",
				"version":   68,
				"action":    "Up",
			}).
			Info("Applying migration")

		filter := bson.M{
			"user_id": bson.M{"$nin": []interface{}{nil, ""}},
		}

		rename := bson.M{
			"$rename": bson.M{
				"user_id": "created_by",
			},
		}

		if _, err := db.Collection("api_keys").UpdateMany(ctx, filter, rename); err != nil {
			return err
		}

		unset := bson.M{
			"$unset": bson.M{
				"user_id": "",
			},
		}

		_, err := db.Collection("api_keys").UpdateMany(ctx, filter, unset)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.
			WithFields(log.Fields{
				"component": "migration",
				"version":   68,
				"action":    "Down",
			}).
			Info("Applying migration")

		filter := bson.M{
			"created_by": bson.M{"$nin": []interface{}{nil, ""}},
		}

		rename := bson.M{
			"$rename": bson.M{
				"created_by": "user_id",
			},
		}

		if _, err := db.Collection("api_keys").UpdateMany(ctx, filter, rename); err != nil {
			return err
		}

		unset := bson.M{
			"$unset": bson.M{
				"created_by": "",
			},
		}

		_, err := db.Collection("api_keys").UpdateMany(ctx, filter, unset)

		return err
	}),
}
