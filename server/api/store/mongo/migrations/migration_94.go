package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration94 = migrate.Migration{
	Version:     94,
	Description: "Adding 'disconnected_at' attribute to 'devices'",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   94,
			"action":    "Up",
		}).Info("Applying migration")

		session, err := db.Client().StartSession()
		if err != nil {
			return err
		}
		defer session.EndSession(ctx)

		fn := func(_ mongo.SessionContext) (any, error) {
			pipeline := []bson.M{
				{
					"$match": bson.M{
						"uid": bson.M{
							"$exists": true,
						},
					},
				},
				{
					"$lookup": bson.M{
						"from":         "connected_devices",
						"localField":   "uid",
						"foreignField": "uid",
						"as":           "connected_devices_data",
					},
				},
			}

			cursor, err := db.Collection("devices").Aggregate(ctx, pipeline)
			if err != nil {
				return nil, err
			}
			defer cursor.Close(ctx)

			for cursor.Next(ctx) {
				device := make(map[string]any)
				if err := cursor.Decode(&device); err != nil {
					return nil, err
				}

				update := bson.M{"$set": bson.M{"disconnected_at": device["last_seen"]}}
				if connectedDevicesData, ok := device["connected_devices_data"].(bson.A); ok && len(connectedDevicesData) > 0 {
					update = bson.M{"$set": bson.M{"disconnected_at": nil}}
				}

				if _, err := db.Collection("devices").UpdateOne(ctx, bson.M{"_id": device["_id"]}, update); err != nil {
					return nil, err
				}
			}

			if err := db.Collection("connected_devices").Drop(ctx); err != nil {
				return nil, err
			}

			return nil, nil
		}

		_, err = session.WithTransaction(ctx, fn)

		return err
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   94,
			"action":    "Down",
		}).Info("Cannot down migration")

		return nil
	}),
}
