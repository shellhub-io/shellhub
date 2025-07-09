package migrations

import (
	"context"

	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration104 = migrate.Migration{
	Version:     MigrationVersion104,
	Description: "Add devices_removed_count field to namespaces based on existing devices with status removed",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{"component": "migration", "version": MigrationVersion104, "action": "Up"}).Info("Applying migration")

		initDoc := bson.M{
			"$set": bson.M{
				"devices_removed_count": 0,
			},
		}

		if _, err := db.Collection("namespaces").UpdateMany(ctx, bson.M{}, initDoc); err != nil {
			log.WithError(err).Error("Failed to initialize namespace devices_removed_count")

			return err
		}

		pipeline := []bson.M{
			{
				"$match": bson.M{
					"status": "removed",
				},
			},
			{
				"$group": bson.M{
					"_id":   "$tenant_id",
					"count": bson.M{"$sum": 1},
				},
			},
		}

		cursor, err := db.Collection("devices").Aggregate(ctx, pipeline)
		if err != nil {
			log.WithError(err).Error("Failed to aggregate removed devices count")

			return err
		}
		defer cursor.Close(ctx)

		updatedNamespaces := 0

		for cursor.Next(ctx) {
			var result struct {
				ID    string `bson:"_id"`
				Count int    `bson:"count"`
			}

			if err := cursor.Decode(&result); err != nil {
				log.WithError(err).Error("Failed to decode aggregation result")

				continue
			}

			updateDoc := bson.M{
				"$set": bson.M{
					"devices_removed_count": result.Count,
				},
			}

			if _, err := db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": result.ID}, updateDoc); err != nil {
				log.WithFields(log.Fields{"tenant_id": result.ID, "error": err}).Error("Failed to update namespace devices_removed_count")

				continue
			}

			updatedNamespaces++
			log.WithFields(log.Fields{"tenant_id": result.ID, "removed_count": result.Count}).Info("Updated namespace devices_removed_count")
		}

		if err := cursor.Err(); err != nil {
			log.WithError(err).Error("Cursor error during migration")

			return err
		}

		log.WithFields(log.Fields{"updated_namespaces": updatedNamespaces}).Info("Migration completed successfully")

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{"component": "migration", "version": MigrationVersion104, "action": "Down"}).Info("Reverting migration")

		updateDoc := bson.M{
			"$unset": bson.M{
				"devices_removed_count": "",
			},
		}

		result, err := db.Collection("namespaces").UpdateMany(ctx, bson.M{}, updateDoc)
		if err != nil {
			log.WithError(err).Error("Failed to remove devices_removed_count field")

			return err
		}

		log.WithFields(log.Fields{"modified_count": result.ModifiedCount}).Info("Removed devices_removed_count field from namespaces")

		return nil
	}),
}
