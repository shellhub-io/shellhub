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
	Version: 21,
	Up: func(db *mongo.Database) error {
		logrus.Info("Applying migration 21 - Up")
		cursor, err := db.Collection("sessions").Find(context.TODO(), bson.D{})
		if err != nil {
			return err
		}

		for cursor.Next(context.TODO()) {
			session := new(models.Session)
			err = cursor.Decode(&session)
			if err != nil {
				return err
			}

			device := new(models.Device)
			if err := db.Collection("devices").FindOne(context.TODO(), bson.M{"uid": session.DeviceUID}).Decode(&device); err != nil {
				if err != mongo.ErrNoDocuments {
					return err
				}

				if _, err := db.Collection("sessions").DeleteMany(context.TODO(), bson.M{"device_uid": session.DeviceUID}); err != nil {
					return err
				}
			}
		}

		if err := cursor.Err(); err != nil {
			return err
		}

		cursor.Close(context.TODO())

		cursor, err = db.Collection("recorded_sessions").Find(context.TODO(), bson.D{})
		if err != nil {
			return err
		}

		for cursor.Next(context.TODO()) {
			record := new(models.RecordedSession)
			err = cursor.Decode(&record)
			if err != nil {
				return err
			}

			namespace := new(models.Namespace)
			if err := db.Collection("namespaces").FindOne(context.TODO(), bson.M{"tenant_id": record.TenantID}).Decode(&namespace); err != nil {
				if err != mongo.ErrNoDocuments {
					return err
				}

				if _, err := db.Collection("recorded_sessions").DeleteMany(context.TODO(), bson.M{"tenant_id": record.TenantID}); err != nil {
					return err
				}
			}
			session := new(models.Session)
			if err := db.Collection("sessions").FindOne(context.TODO(), bson.M{"uid": record.UID}).Decode(&session); err != nil {
				if err != mongo.ErrNoDocuments {
					return err
				}

				if _, err := db.Collection("recorded_sessions").DeleteMany(context.TODO(), bson.M{"uid": record.UID}); err != nil {
					return err
				}
			}

		}
		if err := cursor.Err(); err != nil {
			return err
		}

		cursor.Close(context.TODO())

		return nil
	},
	Down: func(db *mongo.Database) error {
		logrus.Info("Applying migration 21 - Down")

		return nil
	},
}
