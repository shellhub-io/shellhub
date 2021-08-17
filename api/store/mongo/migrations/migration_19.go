package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration19 = migrate.Migration{
	Version:     19,
	Description: "Remove all fingerprint associated with a public keys collection",
	Up: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   19,
			"action":    "Up",
		}).Info("Applying migration")
		_, err := db.Collection("public_keys").Indexes().DropOne(context.TODO(), "fingerprint")

		return err
	},
	Down: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   19,
			"action":    "Down",
		}).Info("Applying migration")
		mod := mongo.IndexModel{
			Keys:    bson.D{{"fingerprint", 1}},
			Options: options.Index().SetName("fingerprint").SetUnique(true),
		}
		_, err := db.Collection("public_keys").Indexes().CreateOne(context.TODO(), mod)

		return err
	},
}
