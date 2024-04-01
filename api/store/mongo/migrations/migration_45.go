package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration45 = migrate.Migration{
	Version:     45,
	Description: "remove duplicated tags on firewall rules",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   45,
			"action":    "Up",
		}).Info("Applying migration")

		_, err := db.Collection("firewall_rules").Aggregate(ctx,
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
					{"$merge", bson.M{"into": "firewall_rules", "whenMatched": "replace"}},
				},
			},
		)
		if err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   45,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}
