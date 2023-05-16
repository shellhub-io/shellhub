package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration57 = migrate.Migration{
	Version:     57,
	Description: "update billing state to status and its values",
	Up: func(database *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   57,
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
				"$addFields": bson.M{
					"billing.status": bson.M{
						"$switch": bson.M{
							"branches": []bson.M{
								{
									"case": bson.M{
										"$eq": []string{"$billing.state", "processed"},
									},
									"then": "active",
								},
								{
									"case": bson.M{
										"$eq": []string{"$billing.state", "past_due"},
									},
									"then": "past_due",
								},
								{
									"case": bson.M{
										"$eq": []string{"$billing.state", "pending"},
									},
									"then": "canceled",
								},
								{
									"case": bson.M{
										"$eq": []string{"$billing.state", "inactive"},
									},
									"then": "inactive",
								},
								{
									"case": bson.M{
										"$eq": []string{"$billing.state", "canceled"},
									},
									"then": "canceled",
								},
							},
							"default": "canceled",
						},
					},
				},
			},
			{
				"$unset": "billing.state",
			},
			{
				"$merge": bson.M{
					"into":        "namespaces",
					"whenMatched": "replace",
				},
			},
		}

		_, err := database.Collection("namespaces").Aggregate(context.Background(), pipeline)
		if err != nil {
			return err
		}

		return nil
	},
	Down: func(database *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   57,
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
				"$addFields": bson.M{
					"billing.state": bson.M{
						"$switch": bson.M{
							"branches": []bson.M{
								{
									"case": bson.M{
										"$eq": []string{"$billing.status", "active"},
									},
									"then": "processed",
								},
								{
									"case": bson.M{
										"$eq": []string{"$billing.status", "past_due"},
									},
									"then": "past_due",
								},
								{
									"case": bson.M{
										"$eq": []string{"$billing.state", "inactive"},
									},
									"then": "inactive",
								},
								{
									"case": bson.M{
										"$eq": []string{"$billing.state", "canceled"},
									},
									"then": "canceled",
								},
							},
							"default": "canceled",
						},
					},
				},
			},
			{
				"$unset": "billing.status",
			},
			{
				"$merge": bson.M{
					"into":        "namespaces",
					"whenMatched": "replace",
				},
			},
		}

		_, err := database.Collection("namespaces").Aggregate(context.Background(), pipeline)
		if err != nil {
			return err
		}

		return nil
	},
}
