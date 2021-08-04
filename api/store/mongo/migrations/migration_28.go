package migrations

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration28 = migrate.Migration{
	Version:     28,
	Description: "add timestamps fields to collections users and devices",
	Up: func(db *mongo.Database) error {
		logrus.Info("Applying migration 28 - Up")
		if _, err := db.Collection("users").UpdateMany(context.TODO(), bson.M{}, bson.M{"$set": bson.M{"created_at": clock.Now()}}); err != nil {
			return err
		}

		if _, err := db.Collection("devices").UpdateMany(context.TODO(), bson.M{}, bson.M{"$set": bson.M{"created_at": clock.Now()}}); err != nil {
			return err
		}

		return nil
	},
	Down: func(db *mongo.Database) error {
		logrus.Info("Applying migration 28 - Down")

		return nil
	},
}
