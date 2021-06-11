package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration24 = migrate.Migration{
	Version:     24,
	Description: "convert names and emails to lowercase",
	Up: func(db *mongo.Database) error {
		logrus.Info("Applying migration 24 - Up")
		if _, err := db.Collection("users").UpdateMany(context.TODO(), bson.D{}, []bson.M{
			{
				"$set": bson.M{"username": bson.M{"$toLower": "$username"},
					"email": bson.M{"$toLower": "$email"}},
			},
		}); err != nil {
			return err
		}

		_, err := db.Collection("namespaces").UpdateMany(context.TODO(), bson.D{}, []bson.M{
			{
				"$set": bson.M{"name": bson.M{"$toLower": "$name"}},
			},
		})

		return err
	},
	Down: func(db *mongo.Database) error {
		logrus.Info("Applying migration 24 - Down")

		return nil
	},
}
