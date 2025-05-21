package migrations

import (
	"context"
	"strings"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration9 = migrate.Migration{
	Version:     9,
	Description: "Set all devices names to lowercase in the devices collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   9,
			"action":    "Up",
		}).Info("Applying migration")
		cursor, err := db.Collection("devices").Find(ctx, bson.D{})
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			device := new(models.Device)
			err := cursor.Decode(&device)
			if err != nil {
				return err
			}

			device.Name = strings.ToLower(device.Name)
			if _, err = db.Collection("devices").UpdateOne(ctx, bson.M{"uid": device.UID}, bson.M{"$set": bson.M{"name": strings.ToLower(device.Name)}}); err != nil {
				return err
			}
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   9,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}
