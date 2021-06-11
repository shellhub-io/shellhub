package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration15 = migrate.Migration{
	Version: 15,
	Up: func(db *mongo.Database) error {
		logrus.Info("Applying migration 15 - Up")
		_, err := db.Collection("namespaces").UpdateMany(context.TODO(), bson.M{}, []bson.M{
			{
				"$set": bson.M{
					"name": bson.M{"$toLower": "$name"},
				},
			},
		})

		return err
	},
	Down: func(db *mongo.Database) error {
		logrus.Info("Applying migration 15 - Down")

		return nil
	},
}
