package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration8 = migrate.Migration{
	Version:     8,
	Description: "Unset unique on recorded in the sessions collection",
	Up: func(db *mongo.Database) error {
		logrus.Info("Applying migration 8 - Up")
		mod := mongo.IndexModel{
			Keys:    bson.D{{"recorded", 1}},
			Options: options.Index().SetName("recorded").SetUnique(false),
		}
		if _, err := db.Collection("sessions").Indexes().CreateOne(context.TODO(), mod); err != nil {
			return err
		}
		_, err := db.Collection("sessions").UpdateMany(context.TODO(), bson.M{}, bson.M{"$set": bson.M{"recorded": false}})

		return err
	},
	Down: func(db *mongo.Database) error {
		logrus.Info("Applying migration 8 - Down")
		if _, err := db.Collection("sessions").UpdateMany(context.TODO(), bson.M{}, bson.M{"$unset": bson.M{"recorded": ""}}); err != nil {
			return err
		}
		_, err := db.Collection("sessions").Indexes().DropOne(context.TODO(), "recorded")

		return err
	},
}
