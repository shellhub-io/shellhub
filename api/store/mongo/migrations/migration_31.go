package migrations

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration31 = migrate.Migration{
	Version:     31,
	Description: "add last_login field to collection namespaces",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   31,
			"action":    "Up",
		}).Info("Applying migration")
		if _, err := db.Collection("namespaces").UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"created_at": clock.Now()}}); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   31,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}
