package migrations

import (
	"context"

	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration120 = migrate.Migration{
	Version:     120,
	Description: "Reconcile namespace device counter caches and convert to int64",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{"component": "migration", "version": 120, "action": "Up"}).Info("Applying migration")

		initDoc := bson.M{
			"$set": bson.M{
				"devices_accepted_count": int64(0),
				"devices_pending_count":  int64(0),
				"devices_rejected_count": int64(0),
				"devices_removed_count":  int64(0),
			},
		}

		if _, err := db.Collection("namespaces").UpdateMany(ctx, bson.M{}, initDoc); err != nil {
			log.WithError(err).Error("Failed to initialize namespace device counts")

			return err
		}

		pipeline := []bson.M{
			{
				"$group": bson.M{
					"_id": bson.M{
						"tenant_id": "$tenant_id",
						"status":    "$status",
					},
					"count": bson.M{"$sum": 1},
				},
			},
			{
				"$group": bson.M{
					"_id": "$_id.tenant_id",
					"counts": bson.M{
						"$push": bson.M{
							"status": "$_id.status",
							"count":  "$count",
						},
					},
				},
			},
		}

		cursor, err := db.Collection("devices").Aggregate(ctx, pipeline)
		if err != nil {
			log.WithError(err).Error("Failed to aggregate device counts")

			return err
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var result struct {
				ID     string `bson:"_id"`
				Counts []struct {
					Status string `bson:"status"`
					Count  int64  `bson:"count"`
				} `bson:"counts"`
			}

			if err := cursor.Decode(&result); err != nil {
				log.WithError(err).Error("Failed to decode aggregation result")

				continue
			}

			updateDoc := bson.M{
				"devices_accepted_count": int64(0),
				"devices_pending_count":  int64(0),
				"devices_rejected_count": int64(0),
				"devices_removed_count":  int64(0),
			}

			for _, count := range result.Counts {
				switch count.Status {
				case "accepted":
					updateDoc["devices_accepted_count"] = count.Count
				case "pending":
					updateDoc["devices_pending_count"] = count.Count
				case "rejected":
					updateDoc["devices_rejected_count"] = count.Count
				case "removed":
					updateDoc["devices_removed_count"] = count.Count
				}
			}

			if _, err := db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": result.ID}, bson.M{"$set": updateDoc}); err != nil {
				log.WithFields(log.Fields{"tenant_id": result.ID, "error": err}).Error("Failed to update namespace device counts")

				continue
			}

			log.WithFields(log.Fields{"tenant_id": result.ID, "counts": result.Counts}).Info("Reconciled namespace device counts")
		}

		if err := cursor.Err(); err != nil {
			log.WithError(err).Error("Cursor error during migration")

			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{"component": "migration", "version": 120, "action": "Down"}).Info("Reverting migration")

		return nil
	}),
}
