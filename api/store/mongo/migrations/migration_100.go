package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration100 = migrate.Migration{
	Version:     MigrationVersion100,
	Description: "Remove direct-tcpip events",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   MigrationVersion100,
			"action":    "Up",
		}).Info("Applying migration: 100")

		pipeline := mongo.Pipeline{
			{{Key: "$match", Value: bson.M{"type": "direct-tcpip"}}},
			{{Key: "$group", Value: bson.M{
				"_id":   "$session",
				"seats": bson.M{"$addToSet": "$seat"},
			}}},
		}

		cursor, err := db.Collection("sessions_events").Aggregate(ctx, pipeline)
		if err != nil {
			logrus.WithError(err).Error("Failed to aggregate direct-tcpip seats")

			return err
		}

		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var result struct {
				ID    string `bson:"_id"`
				Seats []int  `bson:"seats"`
			}

			if err := cursor.Decode(&result); err != nil {
				logrus.WithError(err).Error("Failed to decode aggregation result")

				return err
			}

			_, err := db.Collection("sessions").UpdateOne(
				ctx,
				bson.M{"uid": result.ID},
				bson.M{"$pullAll": bson.M{
					"events.seats": result.Seats,
				}},
			)
			if err != nil {
				logrus.WithError(err).WithField("session_uid", result.ID).Error("Failed to remove seats from session")

				return err
			}
		}

		if err := cursor.Err(); err != nil {
			logrus.WithError(err).Error("Cursor error during seat removal")

			return err
		}

		_, err = db.Collection("sessions").UpdateMany(
			ctx,
			bson.M{"events.types": "direct-tcpip"},
			bson.M{"$pull": bson.M{
				"events.types": "direct-tcpip",
			}},
		)
		if err != nil {
			logrus.WithError(err).Error("Failed to remove direct-tcpip from events.types in sessions")

			return err
		}

		_, err = db.Collection("sessions_events").DeleteMany(
			ctx,
			bson.M{"type": "direct-tcpip"},
		)
		if err != nil {
			logrus.WithError(err).Error("Failed to remove direct-tcpip events from sessions_events")

			return err
		}

		_, err = db.Collection("sessions").UpdateMany(
			ctx,
			bson.M{
				"recorded":     true,
				"events.types": bson.M{"$size": 0},
			},
			bson.M{"$set": bson.M{"recorded": false}},
		)
		if err != nil {
			logrus.WithError(err).Error("Failed to update recorded flag for sessions with empty events.types")

			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   MigrationVersion100,
			"action":    "Down",
		}).Info("Reverting migration: 100")

		_, err := db.Collection("sessions").UpdateMany(
			ctx,
			bson.M{
				"recorded":     false,
				"events.types": bson.M{"$size": 0},
			},
			bson.M{"$set": bson.M{"recorded": true}},
		)
		if err != nil {
			logrus.WithError(err).Error("Failed to revert recorded flag changes")

			return err
		}

		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   MigrationVersion100,
		}).Warn("Cannot restore deleted direct-tcpip events and seats - data loss is permanent")

		return nil
	}),
}
