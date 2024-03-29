package migrations

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration18 = migrate.Migration{
	Version:     18,
	Description: "Set the max_devices value in the namespaces collection to 3 on enterprise",
	Up: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   18,
			"action":    "Up",
		}).Info("Applying migration")
		if envs.IsEnterprise() {
			_, err := db.Collection("namespaces").UpdateMany(context.TODO(), bson.M{}, bson.M{"$set": bson.M{"max_devices": 3}})

			return err
		}
		_, err := db.Collection("namespaces").UpdateMany(context.TODO(), bson.M{}, bson.M{"$set": bson.M{"max_devices": -1}})

		return err
	},
	Down: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   18,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	},
}
