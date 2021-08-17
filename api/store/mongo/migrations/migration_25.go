package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration25 = migrate.Migration{
	Version:     25,
	Description: "remove devices with no namespaces related",
	Up: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   25,
			"action":    "Up",
		}).Info("Applying migration")
		query := []bson.M{
			{
				"$lookup": bson.M{
					"from":         "namespaces",
					"localField":   "tenant_id",
					"foreignField": "tenant_id",
					"as":           "namespace",
				},
			},
			{
				"$addFields": bson.M{
					"namespace": bson.M{"$anyElementTrue": []interface{}{"$namespace"}},
				},
			},

			{
				"$match": bson.M{
					"namespace": bson.M{"$eq": true},
				},
			},

			{
				"$unset": "namespace",
			},

			{
				"$out": "devices",
			},
		}

		_, err := db.Collection("devices").Aggregate(context.TODO(), query)

		return err
	},
	Down: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   25,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	},
}
