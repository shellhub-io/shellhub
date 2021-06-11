package migrations

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration18 = migrate.Migration{
	Version: 18,
	Up: func(db *mongo.Database) error {
		logrus.Info("Applying migration 18 - Up")
		if envs.IsEnterprise() {
			_, err := db.Collection("namespaces").UpdateMany(context.TODO(), bson.M{}, bson.M{"$set": bson.M{"max_devices": 3}})

			return err
		}
		_, err := db.Collection("namespaces").UpdateMany(context.TODO(), bson.M{}, bson.M{"$set": bson.M{"max_devices": -1}})

		return err
	},
	Down: func(db *mongo.Database) error {
		logrus.Info("Applying migration 18 - Down")

		return nil
	},
}
