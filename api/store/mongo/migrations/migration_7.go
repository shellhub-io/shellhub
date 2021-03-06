package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration7 = migrate.Migration{
	Version:     7,
	Description: "Unset unique on uid and message in the recoded_sessions collection",
	Up: func(db *mongo.Database) error {
		logrus.Info("Applying migration 7 - Up")
		mod := mongo.IndexModel{
			Keys:    bson.D{{"uid", 1}},
			Options: options.Index().SetName("uid").SetUnique(false),
		}
		if _, err := db.Collection("recorded_sessions").Indexes().CreateOne(context.TODO(), mod); err != nil {
			return err
		}

		mod = mongo.IndexModel{
			Keys:    bson.D{{"message", 1}},
			Options: options.Index().SetName("message").SetUnique(false),
		}
		_, err := db.Collection("recorded_sessions").Indexes().CreateOne(context.TODO(), mod)

		return err
	},
	Down: func(db *mongo.Database) error {
		logrus.Info("Applying migration 7 - Down")
		if _, err := db.Collection("recorded_sessions").UpdateMany(context.TODO(), bson.M{}, bson.M{"$unset": bson.M{"uid": ""}}); err != nil {
			return err
		}
		if _, err := db.Collection("recorded_sessions").UpdateMany(context.TODO(), bson.M{}, bson.M{"$unset": bson.M{"message": ""}}); err != nil {
			return err
		}
		if _, err := db.Collection("recorded_sessions").Indexes().DropOne(context.TODO(), "uid"); err != nil {
			return err
		}
		_, err := db.Collection("recorded_sessions").Indexes().DropOne(context.TODO(), "message")

		return err
	},
}
