package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration33 = migrate.Migration{
	Version:     33,
	Description: "insert access_type on members list",
	Up: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   33,
			"action":    "Up",
		}).Info("Applying migration")

		query := []bson.M{
			{

				"$unwind": bson.M{"path": "$members", "preserveNullAndEmptyArrays": "false"},
			},

			{

				"$set": bson.M{
					"members": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"eq": "owner"},
							"then": bson.M{"id": "$members", "access_type": "owner"},
							"else": bson.M{"id": "$members", "access_type": "observer"},
						},
					},
				},
				"$group": bson.M{
					"_id": "tenant_id",
					"members": bson.M{
						"$addToSet": "$members",
					},
				},
			},
		}
		_, err := db.Collection("namespaces").Aggregate(context.TODO(), query)

		return err
	},

	Down: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   33,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	},
}
