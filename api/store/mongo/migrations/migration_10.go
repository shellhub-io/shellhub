package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration10 = migrate.Migration{
	Version:     10,
	Description: "Unset unique on session_record in the users collection",
	Up: func(db *mongo.Database) error {
		logrus.Info("Applying migration 10 - Up")
		mod := mongo.IndexModel{
			Keys:    bson.D{{"session_record", 1}},
			Options: options.Index().SetName("session_record").SetUnique(false),
		}
		if _, err := db.Collection("users").Indexes().CreateOne(context.TODO(), mod); err != nil {
			return err
		}
		_, err := db.Collection("users").UpdateMany(context.TODO(), bson.M{}, bson.M{"$set": bson.M{"session_record": true}})

		return err
	},
	Down: func(db *mongo.Database) error {
		logrus.Info("Applying migration 10 - Down")

		return nil
	},
}
