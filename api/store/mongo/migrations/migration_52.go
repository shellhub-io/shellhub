package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration52 = migrate.Migration{
	Version:     52,
	Description: "add marketing field to users",
	Up: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   52,
			"action":    "Up",
		}).Info("Applying migration")

		_, err := db.Collection("users").Aggregate(context.Background(),
			mongo.Pipeline{
				{
					{"$match", bson.M{}},
				},
				{
					{"$set", bson.M{"email_marketing": true}},
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
			"version":   52,
			"action":    "Down",
		}).Info("Applying migration")

		_, err := db.Collection("users").Aggregate(context.Background(),
			mongo.Pipeline{
				{
					{"$match", bson.M{}},
				},
				{
					{"$set", bson.M{"email_marketing": false}},
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
