package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration49 = migrate.Migration{
	Version:     49,
	Description: "set the number of namespaces owned by each user",
	Up: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   49,
			"action":    "Up",
		}).Info("Applying migration up")

		_, err := db.Collection("users").Aggregate(context.Background(),
			mongo.Pipeline{
				{
					{"$match", bson.M{}},
				},
				{
					{
						"$set", bson.M{"tmp_id": bson.M{"$toString": "$_id"}}, // FIXME: I guess we could eliminate the "$_id" conversion from objectID to string.
					},
				},
				{
					{
						"$lookup", bson.M{
							"from":         "namespaces",
							"foreignField": "owner",
							"localField":   "tmp_id",
							"as":           "tmp",
						},
					},
				},
				{
					{
						"$set", bson.M{"namespaces": bson.M{"$size": "$tmp"}},
					},
				},
				{
					{
						"$unset", bson.A{"tmp_id", "tmp"},
					},
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
			"version":   49,
			"action":    "Down",
		}).Info("Applying migration down")

		_, err := db.Collection("users").Aggregate(context.Background(),
			mongo.Pipeline{
				{
					{"$match", bson.M{}},
				},
				{
					{"$unset", "namespaces"},
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
