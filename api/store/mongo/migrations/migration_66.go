package migrations

import (
	"context"

	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration66 = migrate.Migration{
	Version:     66,
	Description: "Replace the user's MFA attributes.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.
			WithFields(log.Fields{
				"component": "migration",
				"version":   66,
				"action":    "Up",
			}).
			Info("Applying migration")

		filter := bson.M{
			"_id": bson.M{
				"$ne": nil,
			},
		}

		rename := bson.M{
			"$rename": bson.M{
				"status_mfa": "mfa.enabled",
				"secret":     "mfa.secret",
				"codes":      "mfa.recovery_codes",
			},
		}

		if _, err := db.Collection("users").UpdateMany(ctx, filter, rename); err != nil {
			return err
		}

		unset := bson.M{
			"$unset": bson.M{
				"status_mfa": "",
				"secret":     "",
				"codes":      "",
			},
		}

		_, err := db.Collection("users").UpdateMany(ctx, filter, unset)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.
			WithFields(log.Fields{
				"component": "migration",
				"version":   66,
				"action":    "Up",
			}).
			Info("Applying migration")

		log.Info("Unable to undo the MFA object")

		return nil
	}),
}
