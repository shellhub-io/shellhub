package migrations

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/hash"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration67 = migrate.Migration{
	Version:     67,
	Description: "Hash the user's MFA recovery code before storing it as a plain string.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.
			WithFields(log.Fields{
				"component": "migration",
				"version":   67,
				"action":    "Up",
			}).
			Info("Applying migration")

		pipeline := []bson.M{
			{
				"$match": bson.M{
					"mfa.enabled": true,
					"mfa.recovery_codes.0": bson.M{
						"$not": bson.M{
							"$regex": "^\\$",
						},
					},
				},
			},
		}

		cursor, err := db.Collection("users").Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)

		updateModels := make([]mongo.WriteModel, 0)

		for cursor.Next(ctx) {
			user := new(models.User)
			if err := cursor.Decode(user); err != nil {
				return err
			}

			recoveryCodes := make([]string, 0)
			for _, c := range user.MFA.RecoveryCodes {
				hash, err := hash.Do(c)
				if err != nil {
					return err
				}

				recoveryCodes = append(recoveryCodes, hash)

			}

			filter := bson.M{"username": user.Username}
			update := bson.M{"$set": bson.M{"mfa.recovery_codes": recoveryCodes}}

			updateModels = append(updateModels, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(false))
		}

		if len(updateModels) > 0 {
			if _, err := db.Collection("users").BulkWrite(ctx, updateModels); err != nil {
				return err
			}
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.
			WithFields(log.Fields{
				"component": "migration",
				"version":   67,
				"action":    "Down",
			}).
			Info("Applying migration")

		log.Info("Unable to undo the recovery code hash")

		return nil
	}),
}
