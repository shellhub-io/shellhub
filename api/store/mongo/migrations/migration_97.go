package migrations

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration97 = migrate.Migration{
	Version:     MigrationVersion97,
	Description: "Set namespace type to team when type is empty",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   MigrationVersion97,
			"action":    "Up",
		}).Info("Applying migration")

		if _, err := db.
			Collection("namespaces").
			UpdateMany(ctx, bson.M{
				"type": "",
			}, bson.M{
				"$set": bson.M{
					"type": models.TypePersonal,
				},
			}); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   MigrationVersion97,
			"action":    "Down",
		}).Info("Cannot undo migration")

		return nil
	}),
}
