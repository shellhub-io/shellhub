package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration39 = migrate.Migration{
	Version:     39,
	Description: "remove online index from devices collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   39,
			"action":    "Up",
		}).Info("Applying migration")

		if _, err := db.Collection("devices").Indexes().DropOne(ctx, "online_1"); err != nil {
			return err
		}

		_, err := db.Collection("devices").UpdateMany(ctx, bson.M{}, bson.M{"$unset": bson.M{"online": nil}})

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   39,
			"action":    "Down",
		}).Info("Applying migration")

		indexModel := mongo.IndexModel{
			Keys: bson.D{{"online", 1}},
		}

		_, err := db.Collection("devices").Indexes().CreateOne(ctx, indexModel)

		return err
	}),
}
