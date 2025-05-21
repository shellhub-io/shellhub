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
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   8,
			"action":    "Up",
		}).Info("Applying migration")
		mod := mongo.IndexModel{
			Keys:    bson.D{{"recorded", 1}},
			Options: options.Index().SetName("recorded").SetUnique(false),
		}
		if _, err := db.Collection("sessions").Indexes().CreateOne(ctx, mod); err != nil {
			return err
		}
		_, err := db.Collection("sessions").UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"recorded": false}})

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   8,
			"action":    "Down",
		}).Info("Applying migration")
		if _, err := db.Collection("sessions").UpdateMany(ctx, bson.M{}, bson.M{"$unset": bson.M{"recorded": ""}}); err != nil {
			return err
		}
		_, err := db.Collection("sessions").Indexes().DropOne(ctx, "recorded")

		return err
	}),
}
