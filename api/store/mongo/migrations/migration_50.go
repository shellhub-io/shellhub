package migrations

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration50 = migrate.Migration{
	Version:     50,
	Description: "set max number of namespaces per user",
	Up: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   50,
			"action":    "Up",
		}).Info("Applying migration up")

		var maxNamespacesWanted int
		if envs.IsCloud() {
			maxNamespacesWanted = 1
		} else {
			maxNamespacesWanted = -1
		}

		_, err := db.Collection("users").Aggregate(context.Background(),
			mongo.Pipeline{
				{
					{"$match", bson.M{}},
				},
				{
					{"$set", bson.M{"max_namespaces": maxNamespacesWanted}},
				},
				{
					{"$merge", bson.M{"into": "users", "whenMatched": "replace"}},
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
			"version":   50,
			"action":    "Down",
		}).Info("Applying migration down")

		_, err := db.Collection("users").Aggregate(context.Background(),
			mongo.Pipeline{
				{
					{"$match", bson.M{}},
				},
				{
					{"$unset", "max_namespaces"},
				},
				{
					{"$merge", bson.M{"into": "users", "whenMatched": "replace"}},
				},
			},
		)
		if err != nil {
			return err
		}

		return nil
	},
}
