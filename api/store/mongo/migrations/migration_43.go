package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration43 = migrate.Migration{
	Version:     43,
	Description: "add tags field to firewall_rules collection",
	Up: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   43,
			"action":    "Up",
		}).Info("Applying migration")

		_, err := db.Collection("firewall_rules").Aggregate(context.TODO(),
			mongo.Pipeline{
				{
					{"$match", bson.M{}},
				},
				{
					{"$set", bson.M{"filter.hostname": "$hostname"}},
				},
				{
					{"$unset", "hostname"},
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
	},
	Down: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   43,
			"action":    "Down",
		}).Info("Applying migration")

		_, err := db.Collection("firewall_rules").Aggregate(context.TODO(),
			mongo.Pipeline{
				{
					{"$match", bson.M{}},
				},
				{
					{"$set", bson.M{"hostname": "$filter.hostname"}},
				},
				{
					{"$unset", "filter"},
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
	},
}
