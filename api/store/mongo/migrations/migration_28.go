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
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   28,
			"action":    "Up",
		}).Info("Applying migration")
		if _, err := db.Collection("users").UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"created_at": clock.Now()}}); err != nil {
			return err
		}

		if _, err := db.Collection("devices").UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"created_at": clock.Now()}}); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   28,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}
