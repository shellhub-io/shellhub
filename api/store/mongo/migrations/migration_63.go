package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration63 = migrate.Migration{
	Version:     63,
	Description: "add MFA fields to collection users",
	Up: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   63,
			"action":    "Up",
		}).Info("Applying migration")

		update := bson.M{
			"$set": bson.M{
				"status_mfa": false,
				"secret":     "",
				"codes":      []string{},
			},
		}

		if _, err := db.Collection("users").UpdateMany(context.TODO(), bson.M{}, update); err != nil {
			return err
		}

		return nil
	},
	Down: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   63,
			"action":    "Down",
		}).Info("Reverting migration")

		update := bson.M{
			"$unset": bson.M{
				"status_mfa": "",
				"secret":     "",
				"codes":      "",
			},
		}

		if _, err := db.Collection("users").UpdateMany(context.TODO(), bson.M{}, update); err != nil {
			return err
		}

		return nil
	},
}
