package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration30 = migrate.Migration{
	Version:     30,
	Description: "add remote_addr field to collection devices",
	Up: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   30,
			"action":    "Up",
		}).Info("Applying migration")
		if _, err := db.Collection("devices").UpdateMany(context.TODO(), bson.M{}, bson.M{"$set": bson.M{"remote_addr": ""}}); err != nil {
			return err
		}

		return nil
	},
	Down: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   30,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	},
}
