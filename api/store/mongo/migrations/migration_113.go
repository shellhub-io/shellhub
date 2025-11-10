package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration113 = migrate.Migration{
	Version:     113,
	Description: "Add tls structure with enabled, verify, and domain fields to web_endpoints collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   113,
			"action":    "Up",
		}).Info("Applying migration up")

		_, err := db.Collection("tunnels").UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"tls": bson.M{"enabled": false, "verify": false, "domain": ""}}})

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   113,
			"action":    "Down",
		}).Info("Reverting migration down")

		_, err := db.Collection("tunnels").UpdateMany(ctx, bson.M{}, bson.M{"$unset": bson.M{"tls": ""}})

		return err
	}),
}
