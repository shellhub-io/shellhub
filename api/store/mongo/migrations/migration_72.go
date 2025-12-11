package migrations

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Member struct as it was when migration 72 was created (with Status field)
type memberForMigration72 struct {
	ID      string          `json:"id,omitempty" bson:"id,omitempty"`
	AddedAt time.Time       `json:"added_at" bson:"added_at"`
	Email   string          `json:"email" bson:"email,omitempty" validate:"email"`
	Role    authorizer.Role `json:"role" bson:"role" validate:"required,oneof=administrator operator observer"`
	Status  string          `json:"status" bson:"status"`
}

// Namespace struct for migration 72 with the old Member type
type namespaceForMigration72 struct {
	models.Namespace `bson:",inline"`
	Members          []memberForMigration72 `json:"members" bson:"members"`
}

var migration72 = migrate.Migration{
	Version:     72,
	Description: "Adding the 'members.$.status' attribute to the namespace if it does not already exist.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   72,
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
			namespace := new(namespaceForMigration72)
			if err := cursor.Decode(namespace); err != nil {
				return err
			}

			for _, m := range namespace.Members {
				if m.Status == "" {
					updateModel := mongo.
						NewUpdateOneModel().
						SetFilter(bson.M{"tenant_id": namespace.TenantID, "members": bson.M{"$elemMatch": bson.M{"id": m.ID}}}).
						SetUpdate(bson.M{"$set": bson.M{"members.$.status": "accepted"}})

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
			"version":   72,
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
			namespace := new(namespaceForMigration72)
			if err := cursor.Decode(namespace); err != nil {
				return err
			}

			for _, m := range namespace.Members {
				if m.Status != "" {
					updateModel := mongo.
						NewUpdateOneModel().
						SetFilter(bson.M{"tenant_id": namespace.TenantID, "members": bson.M{"$elemMatch": bson.M{"id": m.ID}}}).
						SetUpdate(bson.M{"$unset": bson.M{"members.$.status": ""}})

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
}
