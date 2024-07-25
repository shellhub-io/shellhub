package migrations

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration74 = migrate.Migration{
	Version:     74,
	Description: "Adding default message on announcement if is not set.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   74,
			"action":    "Up",
		}).Info("Applying migration")

		filter := bson.M{
			"settings.connection_announcement": "",
		}

		update := bson.M{
			"$set": bson.M{
				"settings.connection_announcement": models.DefaultAnnouncementMessage,
			},
		}

		_, err := db.
			Collection("namespaces").
			UpdateMany(ctx, filter, update)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   74,
			"action":    "Down",
		}).Info("Reverting migration")

		filter := bson.M{
			"settings.connection_announcement": models.DefaultAnnouncementMessage,
		}

		update := bson.M{
			"$set": bson.M{
				"settings.connection_announcement": "",
			},
		}

		_, err := db.
			Collection("namespaces").
			UpdateMany(ctx, filter, update)

		return err
	}),
}
