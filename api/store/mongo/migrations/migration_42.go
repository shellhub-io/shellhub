package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration42 = migrate.Migration{
	Version:     42,
	Description: "change hostname to filter hostname",
	Up: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   42,
			"action":    "Up",
		}).Info("Applying migration")

		_, err := db.Collection("public_keys").Aggregate(context.TODO(),
			mongo.Pipeline{
				{
					{"$match", bson.M{}},
				},
				{
					{"$set", bson.M{"filter.hostname": "$hostname"}},
				},
				{
					{"$unset", "hostname"},
				},
				{
					{"$merge", bson.M{"into": "public_keys"}},
				},
			},
		)
		if err != nil {
			return err
		}

		return nil
	},
	Down: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   42,
			"action":    "Down",
		}).Info("Applying migration")

		_, err := db.Collection("public_keys").Aggregate(context.TODO(),
			mongo.Pipeline{
				{
					{"$match", bson.M{}},
				},
				{
					{"$set", bson.M{"hostname": "$hostname"}},
				},
				{
					{"$unset", "filter"},
				},
				{
					{"$merge", bson.M{"into": "public_keys"}},
				},
			},
		)
		if err != nil {
			return err
		}

		return nil
	},
}
