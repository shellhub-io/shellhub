package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration92 = migrate.Migration{
	Version:     MigrationVersion92,
	Description: "Adding seat and seats to sessions and sessions events",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   MigrationVersion92,
			"action":    "Up",
		}).Info("Applying migration")

		if _, err := db.
			Collection("sessions").
			UpdateMany(ctx, bson.M{
				"authenticated": true,
			}, bson.M{
				"$set": bson.M{
					"events.seats": bson.A{0},
				},
			}); err != nil {
			return err
		}

		if _, err := db.
			Collection("sessions").
			UpdateMany(ctx, bson.M{
				"authenticated": false,
			}, bson.M{
				"$set": bson.M{
					"events.seats": bson.A{},
				},
			}); err != nil {
			return err
		}

		if _, err := db.
			Collection("sessions_events").
			UpdateMany(ctx, bson.M{}, bson.M{
				"$set": bson.M{
					"seat": 0,
				},
			}); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   MigrationVersion92,
			"action":    "Down",
		}).Info("Reverting migration")

		filter := bson.M{}

		if _, err := db.
			Collection("sessions").
			UpdateMany(ctx, filter, bson.M{
				"$unset": bson.M{
					"events.seats": "",
				},
			}); err != nil {
			return err
		}

		if _, err := db.
			Collection("sessions_events").
			UpdateMany(ctx, filter, bson.M{
				"$unset": bson.M{
					"seat": "",
				},
			}); err != nil {
			return err
		}

		return nil
	}),
}
