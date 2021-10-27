package migrations

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration36 = migrate.Migration{
	Version:     36,
	Description: "update max_devices to 3",
	Up: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   36,
			"action":    "Up",
		}).Info("Applying migration")

		if envs.IsCloud() {
			if _, err := db.Collection("namespaces").UpdateMany(context.TODO(), bson.M{"billing": nil}, bson.M{"$set": bson.M{"max_devices": 3}}); err != nil {
				return err
			}
		}

		return nil
	},
	Down: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   36,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	},
}
