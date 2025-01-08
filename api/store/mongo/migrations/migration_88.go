package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration88 = migrate.Migration{
	Version:     88,
	Description: "Adding an 'authentication.saml' attributes to system collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   88,
			"action":    "Up",
		}).Info("Applying migration")

		filter := bson.M{
			"authentication.saml": bson.M{"$exists": false},
		}

		update := bson.M{
			"$set": bson.M{
				"authentication.saml": bson.M{
					"enabled": false,
					"idp": bson.M{
						"entity_id":    "",
						"signon_url":   "",
						"certificates": []string{},
					},
					"sp": bson.M{
						"sign_auth_requests": false,
						"certificate":        "",
						"private_key":        "",
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
			"version":   88,
			"action":    "Down",
		}).Info("Reverting migration")

		filter := bson.M{
			"authentication.saml": bson.M{"$exists": true},
		}

		update := bson.M{
			"$unset": bson.M{
				"authentication.saml": "",
			},
		}

		_, err := db.
			Collection("system").
			UpdateMany(ctx, filter, update)

		return err
	}),
}
