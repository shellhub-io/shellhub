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
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   50,
			"action":    "Up",
		}).Info("Applying migration up")

		var err error
		if envs.IsCloud() {
			_, err = db.Collection("users").Aggregate(context.Background(),
				mongo.Pipeline{
					{
						{"$match", bson.M{}},
					},
					{
						{"$set", bson.M{"tmp": bson.M{"$toString": "$_id"}}},
					},
					{
						{
							"$lookup", bson.M{
								"from": "namespaces",
								"let":  bson.M{"owner": "$tmp"},
								"pipeline": mongo.Pipeline{
									{
										{"$match", bson.M{
											"$expr": bson.M{
												"$and": bson.A{
													bson.M{"$eq": bson.A{"$owner", "$$owner"}},
													bson.M{"$eq": bson.A{"$billing.active", true}},
												},
											},
										}},
									},
								},
								"as": "list",
							},
						},
					},
					{
						{"$set", bson.M{"max_namespaces": bson.M{"$add": bson.A{bson.M{"$size": "$list"}, 1}}}},
					},
					{
						{"$unset", bson.A{"tmp", "list"}},
					},
					{
						{"$merge", bson.M{"into": "users", "whenMatched": "replace"}},
					},
				},
			)
		} else {
			_, err = db.Collection("users").Aggregate(context.Background(),
				mongo.Pipeline{
					{
						{"$match", bson.M{}},
					},
					{
						{"$set", bson.M{"max_namespaces": -1}},
					},
					{
						{"$merge", bson.M{"into": "users", "whenMatched": "replace"}},
					},
				},
			)
		}
		if err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
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
	}),
}
