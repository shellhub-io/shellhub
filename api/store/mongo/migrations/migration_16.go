package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration16 = migrate.Migration{
	Version:     16,
	Description: "Set the fingerprint as unique on public_keys collection",
	Up: func(db *mongo.Database) error {
		logrus.Info("Applying migration 16 - Up")
		mod := mongo.IndexModel{
			Keys:    bson.D{{"fingerprint", 1}},
			Options: options.Index().SetName("fingerprint").SetUnique(true),
		}
		_, err := db.Collection("public_keys").Indexes().CreateOne(context.TODO(), mod)

		return err
	},

	Down: func(db *mongo.Database) error {
		logrus.Info("Applying migration 16 - Down")
		_, err := db.Collection("public_keys").Indexes().DropOne(context.TODO(), "fingerprint")

		return err
	},
}
