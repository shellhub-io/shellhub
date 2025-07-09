package migrations

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration75 = migrate.Migration{
	Version:     MigrationVersion75,
	Description: "Convert user.confirmed to user.status",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   MigrationVersion75,
			"action":    "Up",
		}).Info("Applying migration")

		pipeline := []bson.M{
			{
				"$match": bson.M{
					"confirmed": bson.M{
						"$exists": true,
					},
					"status": bson.M{
						"$exists": false,
					},
				},
			},
		}

		cursor, err := db.Collection("users").Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)

		updateModels := make([]mongo.WriteModel, 0)

		for cursor.Next(ctx) {
			user := make(map[string]interface{})
			if err := cursor.Decode(&user); err != nil {
				return err
			}

			updateModel := mongo.
				NewUpdateOneModel().
				SetFilter(bson.M{"_id": user["_id"]})

			if confirmed := user["confirmed"]; confirmed == true {
				updateModel.SetUpdate(bson.M{"$set": bson.M{"status": models.UserStatusConfirmed.String()}, "$unset": bson.M{"confirmed": ""}})
			} else {
				updateModel.SetUpdate(bson.M{"$set": bson.M{"status": models.UserStatusNotConfirmed.String()}, "$unset": bson.M{"confirmed": ""}})
			}

			updateModels = append(updateModels, updateModel)
		}

		if len(updateModels) == 0 {
			return nil
		}

		_, err = db.Collection("users").BulkWrite(ctx, updateModels)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   MigrationVersion75,
			"action":    "Down",
		}).Info("Reverting migration")

		pipeline := []bson.M{
			{
				"$match": bson.M{
					"confirmed": bson.M{
						"$exists": false,
					},
					"status": bson.M{
						"$exists": true,
					},
				},
			},
		}

		cursor, err := db.Collection("users").Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)

		updateModels := make([]mongo.WriteModel, 0)

		for cursor.Next(ctx) {
			user := make(map[string]interface{})
			if err := cursor.Decode(&user); err != nil {
				return err
			}

			updateModel := mongo.
				NewUpdateOneModel().
				SetFilter(bson.M{"_id": user["_id"]})

			if status := user["status"].(string); status == models.UserStatusConfirmed.String() {
				updateModel.SetUpdate(bson.M{"$set": bson.M{"confirmed": true}, "$unset": bson.M{"status": ""}})
			} else {
				updateModel.SetUpdate(bson.M{"$set": bson.M{"confirmed": false}, "$unset": bson.M{"status": ""}})
			}

			updateModels = append(updateModels, updateModel)
		}

		if len(updateModels) == 0 {
			return nil
		}

		_, err = db.Collection("users").BulkWrite(ctx, updateModels)

		return err
	}),
}
