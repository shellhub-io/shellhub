package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration64 = migrate.Migration{
	Version:     64,
	Description: "Adding the 'settings.connection_announcement' attribute to the namespace if it does not already exist.",
	Up: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   64,
			"action":    "Up",
		}).Info("Applying migration")

		filter := bson.M{
			"settings.connection_announcement": bson.M{"$in": []interface{}{nil, ""}},
		}

		update := bson.M{
			"$set": bson.M{
				"settings.connection_announcement": "",
			},
		}

		_, err := db.
			Collection("namespaces").
			UpdateMany(context.TODO(), filter, update)

		return err
	},
	Down: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   64,
			"action":    "Down",
		}).Info("Reverting migration")

		filter := bson.M{
			"settings.connection_announcement": bson.M{"$in": []interface{}{nil, ""}},
		}

		update := bson.M{
			"$unset": bson.M{
				"settings.connection_announcement": "",
			},
		}

		_, err := db.
			Collection("namespaces").
			UpdateMany(context.TODO(), filter, update)

		return err
	},
}
