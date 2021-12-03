package migrations

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration38 = migrate.Migration{
	Version:     38,
	Description: "Set last_login to created_at, when created_at is a zero value",
	Up: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   38,
			"action":    "Up",
		}).Info("Applying migration")
		zeroTime := time.Time{}.UTC()
		_, err := db.Collection("users").Aggregate(context.TODO(),
			mongo.Pipeline{
				{
					{"$match", bson.D{
						{"$and", bson.A{
							bson.M{"$or": bson.A{
								bson.M{"created_at": bson.M{"$eq": zeroTime}},
								bson.M{"created_at": bson.M{"$eq": nil}},
							}},
							bson.M{"last_login": bson.M{"$ne": zeroTime}},
						}},
					}},
				},
				{
					{"$replaceRoot", bson.D{{"newRoot", bson.M{"$mergeObjects": bson.A{"$$ROOT", bson.M{"created_at": "$last_login"}}}}}},
				},
				{
					{"$merge", bson.M{"into": "users"}},
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
			"version":   38,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	},
}
