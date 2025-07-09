package migrations

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration103 = migrate.Migration{
	Version:     MigrationVersion103,
	Description: "Convert devices_removed from removed_devices collection to devices with status removed",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{"component": "migration", "version": MigrationVersion103, "action": "Up"}).Info("Applying migration")

		cursor, err := db.Collection("removed_devices").Find(ctx, bson.M{})
		if err != nil {
			log.WithError(err).Error("Failed to find removed devices")

			return err
		}
		defer cursor.Close(ctx)

		processedCount := 0
		skippedCount := 0

		for cursor.Next(ctx) {
			var removedDevice struct {
				Device    map[string]any `bson:"device"`
				Timestamp any            `bson:"timestamp"`
			}

			if err := cursor.Decode(&removedDevice); err != nil {
				log.WithError(err).Error("Failed to decode removed device")

				continue
			}

			if removedDevice.Device == nil {
				log.Warn("Skipping removed device with nil device data")
				skippedCount++

				continue
			}

			existingDevice := db.Collection("devices").FindOne(ctx, bson.M{"uid": removedDevice.Device["uid"]})
			if existingDevice.Err() == nil {
				log.WithFields(log.Fields{"uid": removedDevice.Device["uid"]}).Info("Device already exists in devices collection, skipping")
				skippedCount++

				continue
			}

			deviceDoc := removedDevice.Device
			deviceDoc["status"] = string(models.DeviceStatusRemoved)
			deviceDoc["status_updated_at"] = removedDevice.Timestamp

			if _, err := db.Collection("devices").InsertOne(ctx, deviceDoc); err != nil {
				log.WithFields(log.Fields{"uid": removedDevice.Device["uid"], "error": err}).Error("Failed to insert device")

				continue
			}

			processedCount++
			log.WithFields(log.Fields{"uid": removedDevice.Device["uid"]}).Info("Successfully converted removed device to devices collection")
		}

		if err := cursor.Err(); err != nil {
			log.WithError(err).Error("Cursor error during migration")

			return err
		}

		log.WithFields(log.Fields{"processed_count": processedCount, "skipped_count": skippedCount}).
			Info("Migration completed successfully")

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{"component": "migration", "version": MigrationVersion103, "action": "Down"}).Info("Reverting migration")

		// NOTE: This rollback has a known limitation - we don't store the original device status
		// before it was changed to "removed". This means when reverting, devices in the
		// removed_devices collection will have status="removed" instead of their original
		// pre-deletion status (e.g., "accepted", "pending", etc.).
		// This is acceptable for rollback purposes as the main goal is data preservation.

		cursor, err := db.Collection("devices").Find(ctx, bson.M{"status": "removed"})
		if err != nil {
			log.WithError(err).Error("Failed to find devices with status removed")

			return err
		}
		defer cursor.Close(ctx)

		processedCount := 0
		skippedCount := 0

		for cursor.Next(ctx) {
			var device map[string]any

			if err := cursor.Decode(&device); err != nil {
				log.WithError(err).Error("Failed to decode device")

				continue
			}

			timestamp, exists := device["status_updated_at"]
			if !exists {
				log.WithFields(log.Fields{"uid": device["uid"]}).Warn("Device missing status_updated_at, skipping")
				skippedCount++

				continue
			}

			removedDevice := bson.M{
				"device":    device,
				"timestamp": timestamp,
			}

			if _, err := db.Collection("removed_devices").InsertOne(ctx, removedDevice); err != nil {
				log.WithFields(log.Fields{"uid": device["uid"], "error": err}).Error("Failed to insert removed device")

				continue
			}

			processedCount++
			log.WithFields(log.Fields{"uid": device["uid"]}).Info("Successfully reverted device to removed_devices collection")
		}

		if err := cursor.Err(); err != nil {
			log.WithError(err).Error("Cursor error during migration revert")

			return err
		}

		result, err := db.Collection("devices").DeleteMany(ctx, bson.M{"status": "removed"})
		if err != nil {
			log.WithError(err).Error("Failed to remove devices with status removed")

			return err
		}

		log.WithFields(log.Fields{"processed_count": processedCount, "skipped_count": skippedCount, "deleted_count": result.DeletedCount}).
			Info("Migration revert completed successfully")

		return nil
	}),
}
