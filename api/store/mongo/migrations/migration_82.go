package migrations

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

var migration82 = migrate.Migration{
	Version:     82,
	Description: "Adding the 'namespaces.type' attribute to the namespaces if it does not already exist.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   82,
			"action":    "Up",
		}).Info("Applying migration")

		filter := bson.M{
			"type": bson.M{"$in": []interface{}{nil, ""}},
		}

		update := bson.M{
			"$set": bson.M{
				"type": models.TypeTeam,
			},
		}

		_, err := db.
			Collection("namespaces",
				options.Collection().SetWriteConcern(writeconcern.Majority()),
			).
			UpdateMany(ctx, filter, update)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   82,
			"action":    "Down",
		}).Info("Reverting migration")

		filter := bson.M{
			"type": bson.M{"$in": []interface{}{nil, ""}},
		}

		update := bson.M{
			"$unset": bson.M{
				"type": models.TypeTeam,
			},
		}

		_, err := db.
			Collection("namespaces",
				options.Collection().SetWriteConcern(writeconcern.Majority()),
			).
			UpdateMany(ctx, filter, update)

		return err
	}),
}
