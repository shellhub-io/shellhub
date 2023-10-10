package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration59 = migrate.Migration{
	Version:     59,
	Description: "Converts all 'name' field values in the 'users' collection to lowercase.",
	Up: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   59,
			"action":    "Up",
		}).Info("Starting migration Up action.")

		_, err := db.Collection("users").UpdateMany(
			context.TODO(),
			bson.M{},
			[]bson.M{
				{
					"$set": bson.M{
						"username": bson.M{
							"$toLower": "$username",
						},
						"email": bson.M{
							"$toLower": "$email",
						},
					},
				},
			},
		)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"component": "migration",
				"version":   59,
				"action":    "Up",
				"error":     err.Error(),
			}).Error("Failed to execute Up migration.")

			return err
		}

		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   59,
			"action":    "Up",
		}).Info("Completed migration Up action successfully.")

		return nil
	},
	Down: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   59,
			"action":    "Down",
		}).Info("Starting migration Down action.")

		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   59,
			"action":    "Down",
		}).Info("Completed migration Down action successfully.")

		return nil
	},
}
