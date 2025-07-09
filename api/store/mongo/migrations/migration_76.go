package migrations

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration76 = migrate.Migration{
	Version:     MigrationVersion76,
	Description: "Remove user.namespace from users collection.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithField("component", "migration").
			WithField("version", MigrationVersion76).
			WithField("action", " Up").
			Info("Applying migration")

		filter := bson.M{"namespaces": bson.M{"$exists": true}}
		update := bson.M{"$unset": bson.M{"namespaces": ""}}

		_, err := db.Collection("users").UpdateMany(ctx, filter, update)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithField("component", "migration").
			WithField("version", MigrationVersion76).
			WithField("action", "Down").
			Info("Applying migration")

		filter := []bson.M{
			{
				"$match": bson.M{
					"namespaces": bson.M{
						"$exists": false,
					},
				},
			},
		}

		cursor, err := db.Collection("users").Aggregate(ctx, filter)
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)

		updateModels := make([]mongo.WriteModel, 0)
		for cursor.Next(ctx) {
			user := new(models.User)
			if err := cursor.Decode(user); err != nil {
				return err
			}

			cursor, err := db.Collection("namespaces").Find(ctx, bson.M{"members": bson.M{"$elemMatch": bson.M{"id": user.ID, "role": "owner"}}})
			if err != nil {
				return err
			}
			defer cursor.Close(ctx)

			namespaces := make([]models.Namespace, 0)
			if err := cursor.All(ctx, &namespaces); err != nil {
				continue
			}

			userID, _ := primitive.ObjectIDFromHex(user.ID)
			updateModel := mongo.
				NewUpdateOneModel().
				SetFilter(bson.M{"_id": userID}).
				SetUpdate(bson.M{"$set": bson.M{"namespaces": len(namespaces)}})

			updateModels = append(updateModels, updateModel)
		}

		if len(updateModels) == 0 {
			return nil
		}

		_, err = db.Collection("users").BulkWrite(ctx, updateModels)

		return err
	}),
}
