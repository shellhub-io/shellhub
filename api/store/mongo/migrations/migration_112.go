package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration112 = migrate.Migration{
	Version:     112,
	Description: "Remove events subdocument from sessions collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   112,
			"action":    "Up",
		}).Info("Applying migration up")

		if _, err := db.Collection("sessions").UpdateMany(ctx, bson.M{}, bson.M{"$unset": bson.M{"events": ""}}); err != nil { // nolint:revive
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   112,
			"action":    "Down",
		}).Info("Reverting migration down")

		pipeline := []bson.M{
			{
				"$lookup": bson.M{
					"from": "sessions_events",
					"let":  bson.M{"sessionUID": "$uid"},
					"pipeline": []bson.M{
						{
							"$match": bson.M{
								"$expr": bson.M{"$eq": []string{"$session", "$$sessionUID"}},
							},
						},
						{
							"$group": bson.M{
								"_id":   nil,
								"types": bson.M{"$addToSet": "$type"},
								"seats": bson.M{"$addToSet": "$seat"},
							},
						},
					},
					"as": "eventData",
				},
			},
			{
				"$set": bson.M{
					"events": bson.M{
						"$cond": bson.M{
							"if": bson.M{"$gt": []any{bson.M{"$size": "$eventData"}, 0}},
							"then": bson.M{
								"types": bson.M{"$arrayElemAt": []any{"$eventData.types", 0}},
								"seats": bson.M{"$arrayElemAt": []any{"$eventData.seats", 0}},
							},
							"else": bson.M{
								"types": []string{},
								"seats": []int{},
							},
						},
					},
				},
			},
			{
				"$unset": "eventData",
			},
			{
				"$merge": bson.M{
					"into":        "sessions",
					"whenMatched": "replace",
				},
			},
		}

		_, err := db.Collection("sessions").Aggregate(ctx, pipeline)

		return err
	}),
}
