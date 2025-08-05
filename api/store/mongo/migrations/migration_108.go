package migrations

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration108 = migrate.Migration{
	Version:     108,
	Description: "Migrate session events to session seats structure",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{"component": "migration", "version": 108, "action": "Up"}).Info("Applying migration")

		cursor, err := db.Collection("sessions").Find(ctx, bson.M{
			"seats":  nil,
			"events": bson.M{"$exists": true},
		})
		if err != nil {
			log.WithError(err).Error("Failed to fetch sessions")

			return err
		}

		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var session struct {
				UID   string               `bson:"uid"`
				Seats []models.SessionSeat `bson:"seats"`
			}

			if err := cursor.Decode(&session); err != nil {
				log.WithError(err).Error("Failed to decode session")

				return err
			}

			eventsCursor, err := db.Collection("sessions_events").Find(ctx, bson.M{"session": session.UID})
			if err != nil {
				log.WithError(err).WithField("session", session.UID).Error("Failed to fetch events for session")

				return err
			}

			eventsBySeat := make(map[int][]string)

			for eventsCursor.Next(ctx) {
				var event struct {
					Type string `bson:"type"`
					Seat int    `bson:"seat"`
				}

				if err := eventsCursor.Decode(&event); err != nil {
					log.WithError(err).Error("Failed to decode event")

					return err
				}

				if _, ok := eventsBySeat[event.Seat]; !ok {
					eventsBySeat[event.Seat] = []string{}
				}

				eventsBySeat[event.Seat] = append(eventsBySeat[event.Seat], event.Type)
			}

			eventsCursor.Close(ctx)

			var seats []models.SessionSeat
			var seatIDs []int

			eventTypes := make(map[string]bool)

			for seatID, events := range eventsBySeat {
				seats = append(seats, models.SessionSeat{
					ID:     seatID,
					Events: events,
				})

				seatIDs = append(seatIDs, seatID)

				for _, eventType := range events {
					eventTypes[eventType] = true
				}
			}

			var types []string
			for eventType := range eventTypes {
				types = append(types, eventType)
			}

			_, err = db.Collection("sessions").UpdateOne(ctx,
				bson.M{"uid": session.UID},
				bson.M{
					"$set": bson.M{
						"seats": seats,
						"events": models.SessionEvents{
							Types: types,
							Seats: seatIDs,
						},
					},
				})
			if err != nil {
				log.WithError(err).WithField("session", session.UID).Error("Failed to update session")

				return err
			}
		}

		log.WithFields(log.Fields{"component": "migration", "version": 108, "action": "Up"}).Info("Migration completed successfully")

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{"component": "migration", "version": 108, "action": "Down"}).Info("Reverting migration")

		if _, err := db.Collection("sessions").UpdateMany(
			ctx,
			bson.M{},
			bson.M{
				"$unset": bson.M{
					"seats": "",
				},
			},
		); err != nil {
			log.WithError(err).Error("Failed to revert events migration")

			return err
		}

		log.WithFields(log.Fields{"component": "migration", "version": 108, "action": "Down"}).Info("Migration reverted successfully")

		return nil
	}),
}
