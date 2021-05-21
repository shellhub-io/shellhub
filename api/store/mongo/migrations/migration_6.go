package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration_6 = migrate.Migration{
	Version: 6,
	Up: func(db *mongo.Database) error {
		logrus.Info("Applying migration 6 - Up")
		mod := mongo.IndexModel{
			Keys:    bson.D{{"status", 1}},
			Options: options.Index().SetName("status").SetUnique(false),
		}
		if _, err := db.Collection("devices").Indexes().CreateOne(context.TODO(), mod); err != nil {
			return err
		}
		_, err := db.Collection("devices").UpdateMany(context.TODO(), bson.M{}, bson.M{"$set": bson.M{"status": "accepted"}})
		return err
	},
	Down: func(db *mongo.Database) error {
		logrus.Info("Applying migration 6 - Down")
		if _, err := db.Collection("devices").UpdateMany(context.TODO(), bson.M{}, bson.M{"$unset": bson.M{"status": ""}}); err != nil {
			return err
		}
		_, err := db.Collection("status").Indexes().DropOne(context.TODO(), "status")
		return err
	},
}
