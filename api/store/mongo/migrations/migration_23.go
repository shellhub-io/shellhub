package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration23 = migrate.Migration{
	Version:     23,
	Description: "change dot in namespace name and hostname to -",
	Up: func(db *mongo.Database) error {
		logrus.Info("Applying migration 23 - Up")
		if _, err := db.Collection("namespaces").UpdateMany(context.TODO(), bson.D{}, []bson.M{
			{
				"$set": bson.M{
					"name": bson.M{
						"$replaceAll": bson.M{"input": "$name", "find": ".", "replacement": "-"},
					},
				},
			},
		}); err != nil {
			return err
		}

		if _, err := db.Collection("devices").UpdateMany(context.TODO(), bson.D{}, []bson.M{
			{
				"$set": bson.M{
					"name": bson.M{
						"$replaceAll": bson.M{"input": "$name", "find": ".", "replacement": "-"},
					},
				},
			},
		}); err != nil {
			return err
		}

		return nil
	},
	Down: func(db *mongo.Database) error {
		logrus.Info("Applying migration 23 - Down")

		return nil
	},
}
