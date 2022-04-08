package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration44 = migrate.Migration{
	Version:     44,
	Description: "remove duplicated tags on public keys",
	Up: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   44,
			"action":    "Up",
		}).Info("Applying migration")

		_, err := db.Collection("public_keys").Aggregate(context.TODO(),
			mongo.Pipeline{
				{
					{"$match", bson.M{"filter.tags": bson.M{"$exists": true}}},
				},
				{
					{"$unwind", "$filter.tags"},
				},
				{
					{"$group", bson.M{
						"_id":  "$_id",
						"body": bson.M{"$push": "$$ROOT"},
						"tags": bson.M{
							"$addToSet": "$filter.tags",
						},
						"count": bson.M{
							"$sum": 1,
						},
					}},
				},
				{
					{"$replaceRoot", bson.D{{"newRoot", bson.M{"$mergeObjects": bson.A{bson.M{"$arrayElemAt": bson.A{"$body", 0}}, bson.M{"filter": bson.M{"tags": "$tags"}}, bson.M{"_id": "$_id"}}}}}},
				},
				{
					{"$merge", bson.M{"into": "public_keys", "whenMatched": "replace"}},
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
			"version":   44,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	},
}
