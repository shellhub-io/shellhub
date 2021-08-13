package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration32 = migrate.Migration{
	Version:     32,
	Description: "add authenticated field to collection users",
	Up: func(db *mongo.Database) error {
		logrus.Info("Applying migration 32 - Up")
		if _, err := db.Collection("users").UpdateMany(context.TODO(), bson.M{}, bson.M{"$set": bson.M{"authenticated": true}}); err != nil {
			return err
		}

		return nil
	},
	Down: func(db *mongo.Database) error {
		logrus.Info("Applying migration 32 - Down")

		return nil
	},
}
