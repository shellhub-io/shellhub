package migrations

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration21 = migrate.Migration{
	Version:     21,
	Description: "Remove all sessions, recorded_sessions for the devices",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   21,
			"action":    "Up",
		}).Info("Applying migration")
		cursor, err := db.Collection("sessions").Find(ctx, bson.D{})
		if err != nil {
			return err
		}

		for cursor.Next(ctx) {
			session := new(models.Session)
			err = cursor.Decode(&session)
			if err != nil {
				return err
			}

			device := new(models.Device)
			if err := db.Collection("devices").FindOne(ctx, bson.M{"uid": session.DeviceUID}).Decode(&device); err != nil {
				if err != mongo.ErrNoDocuments {
					return err
				}

				if _, err := db.Collection("sessions").DeleteMany(ctx, bson.M{"device_uid": session.DeviceUID}); err != nil {
					return err
				}
			}
		}

		if err := cursor.Err(); err != nil {
			return err
		}

		cursor.Close(ctx)

		cursor, err = db.Collection("recorded_sessions").Find(ctx, bson.D{})
		if err != nil {
			return err
		}

		for cursor.Next(ctx) {
			record := new(models.RecordedSession)
			err = cursor.Decode(&record)
			if err != nil {
				return err
			}

			namespace := new(models.Namespace)
			if err := db.Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": record.TenantID}).Decode(&namespace); err != nil {
				if err != mongo.ErrNoDocuments {
					return err
				}

				if _, err := db.Collection("recorded_sessions").DeleteMany(ctx, bson.M{"tenant_id": record.TenantID}); err != nil {
					return err
				}
			}
			session := new(models.Session)
			if err := db.Collection("sessions").FindOne(ctx, bson.M{"uid": record.UID}).Decode(&session); err != nil {
				if err != mongo.ErrNoDocuments {
					return err
				}

				if _, err := db.Collection("recorded_sessions").DeleteMany(ctx, bson.M{"uid": record.UID}); err != nil {
					return err
				}
			}

		}
		if err := cursor.Err(); err != nil {
			return err
		}

		cursor.Close(ctx)

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   21,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}
