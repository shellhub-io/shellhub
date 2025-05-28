package migrations

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration91 = migrate.Migration{
	Version:     91,
	Description: "Add sessions_events collections",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   91,
			"action":    "Up",
		}).Info("Applying migration")

		if err := db.CreateCollection(ctx, "sessions_events"); err != nil {
			return err
		}

		sessionIndex := mongo.IndexModel{
			Keys: bson.M{
				"session": 1,
			},
		}

		if _, err := db.Collection("sessions_events").Indexes().CreateOne(ctx, sessionIndex); err != nil {
			return err
		}

		cursor, err := db.Collection("sessions").Find(ctx, bson.M{"events.items": bson.M{"$exists": true}})
		if err != nil {
			return err
		}

		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var session struct {
				UID    string `bson:"uid"`
				Events struct {
					Items []models.SessionEvent `bson:"items"`
				} `bson:"events"`
			}

			if err := cursor.Decode(&session); err != nil {
				return err
			}

			for _, event := range session.Events.Items {
				event.Session = session.UID
				if _, err := db.Collection("sessions_events").InsertOne(ctx, event); err != nil {
					return err
				}
			}

			if _, err := db.Collection("sessions").UpdateOne(ctx, bson.M{
				"uid": session.UID,
			}, bson.M{
				"$unset": bson.M{
					"events.items": "",
				},
			}); err != nil {
				return err
			}
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   91,
			"action":    "Down",
		}).Info("Reverting migration")
		cursor, err := db.Collection("sessions_events").Find(ctx, bson.M{})
		if err != nil {
			return err
		}

		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var event models.SessionEvent
			if err := cursor.Decode(&event); err != nil {
				return err
			}
			sessionID := event.Session

			event.Session = ""
			update := bson.M{"$push": bson.M{"events.items": event}}
			if _, err := db.Collection("sessions").UpdateOne(ctx, bson.M{"uid": sessionID}, update); err != nil {
				return err
			}
		}

		if err := db.Collection("sessions_events").Drop(ctx); err != nil {
			return err
		}

		return nil
	}),
}
