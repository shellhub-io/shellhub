package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration53 = migrate.Migration{
	Version:     53,
	Description: "create index to announcement ID",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   53,
			"action":    "Up",
		}).Info("Applying migration")
		field := "uuid"
		collection := "announcements"
		unique := true

		if _, err := db.Collection(collection).Indexes().CreateOne(context.Background(), mongo.IndexModel{
			Keys: bson.M{
				field: 1,
			},
			Options: &options.IndexOptions{ //nolint:exhaustruct
				Name:   &field,
				Unique: &unique,
			},
		}); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   53,
			"action":    "Down",
		}).Info("Applying migration")
		index := "uuid"
		collection := "announcements"

		if _, err := db.Collection(collection).Indexes().DropOne(context.Background(), index); err != nil {
			return err
		}

		return nil
	}),
}
