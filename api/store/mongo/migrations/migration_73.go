package migrations

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration73 = migrate.Migration{
	Version:     73,
	Description: "Adding the 'members.$.added_at' attribute to the namespace if it does not already exist. The value is the Go time.Time zeroer",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   73,
			"action":    "Up",
		}).Info("Applying migration")

		pipeline := []bson.M{
			{
				"$match": bson.M{
					"tenant_id": bson.M{
						"$exists": true,
					},
				},
			},
		}

		cursor, err := db.Collection("namespaces").Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)

		updateModels := make([]mongo.WriteModel, 0)

		for cursor.Next(ctx) {
			namespace := new(models.Namespace)
			if err := cursor.Decode(namespace); err != nil {
				return err
			}

			for _, m := range namespace.Members {
				if m.AddedAt == (time.Time{}) {
					updateModel := mongo.
						NewUpdateOneModel().
						SetFilter(bson.M{"tenant_id": namespace.TenantID, "members": bson.M{"$elemMatch": bson.M{"id": m.ID}}}).
						// We update the added_at field to the same value as in the if statement
						// because when the attribute is null in MongoDB, it will be converted
						// to the zero value of time.Time.
						SetUpdate(bson.M{"$set": bson.M{"members.$.added_at": time.Time{}}})

					updateModels = append(updateModels, updateModel)
				}
			}
		}

		if len(updateModels) == 0 {
			return nil
		}

		_, err = db.Collection("namespaces").BulkWrite(ctx, updateModels)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   73,
			"action":    "Down",
		}).Info("Reverting migration")

		pipeline := []bson.M{
			{
				"$match": bson.M{
					"tenant_id": bson.M{
						"$exists": true,
					},
				},
			},
		}

		cursor, err := db.Collection("namespaces").Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)

		updateModels := make([]mongo.WriteModel, 0)

		for cursor.Next(ctx) {
			namespace := new(models.Namespace)
			if err := cursor.Decode(namespace); err != nil {
				return err
			}

			for _, m := range namespace.Members {
				updateModel := mongo.
					NewUpdateOneModel().
					SetFilter(bson.M{"tenant_id": namespace.TenantID, "members": bson.M{"$elemMatch": bson.M{"id": m.ID}}}).
					SetUpdate(bson.M{"$unset": bson.M{"members.$.added_at": ""}})

				updateModels = append(updateModels, updateModel)
			}
		}

		if len(updateModels) == 0 {
			return nil
		}

		_, err = db.Collection("namespaces").BulkWrite(ctx, updateModels)

		return err
	}),
}
