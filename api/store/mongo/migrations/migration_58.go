package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration58 = migrate.Migration{
	Version:     58,
	Description: "",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   58,
			"action":    "Up",
		}).Info("Applying migration up")

		pipeline := []bson.M{
			{
				"$match": bson.M{
					"billing": bson.M{
						"$ne": nil,
					},
				},
			},
			{
				"$set": bson.M{
					"billing.current_period_end": bson.M{
						"$convert": bson.M{
							"input": "$billing.current_period_end",
							"to":    "long",
						},
					},
				},
			},
			{
				"$merge": bson.M{
					"into":        "namespaces",
					"whenMatched": "replace",
				},
			},
		}

		_, err := db.Collection("namespaces").Aggregate(context.Background(), pipeline)
		if err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   58,
			"action":    "Down",
		}).Info("Applying migration down")

		pipeline := []bson.M{
			{
				"$match": bson.M{
					"billing": bson.M{
						"$ne": nil,
					},
				},
			},
			{
				"$set": bson.M{
					"billing.current_period_end": bson.M{
						"$convert": bson.M{
							"input": "$billing.current_period_end",
							"to":    "date",
						},
					},
				},
			},
			{
				"$merge": bson.M{
					"into":        "namespaces",
					"whenMatched": "replace",
				},
			},
		}

		_, err := db.Collection("namespaces").Aggregate(context.Background(), pipeline)
		if err != nil {
			return err
		}

		return nil
	}),
}
