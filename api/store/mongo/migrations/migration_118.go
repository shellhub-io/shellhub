package migrations

import (
	"context"

	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration118 = migrate.Migration{
	Version:     118,
	Description: "Migrate member invitations from namespaces members array to membership_invitations collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   118,
			"action":    "Up",
		}).Info("Applying migration up")

		session, err := db.Client().StartSession()
		if err != nil {
			return err
		}
		defer session.EndSession(ctx)

		_, err = session.WithTransaction(ctx, func(sCtx mongo.SessionContext) (any, error) {
			cursor, err := db.Collection("namespaces").Find(sCtx, bson.M{})
			if err != nil {
				log.WithError(err).Error("Failed to find namespaces")

				return nil, err
			}

			defer cursor.Close(sCtx)

			invitations := make([]any, 0)
			namespacesToUpdate := make([]bson.M, 0)

			for cursor.Next(sCtx) {
				namespace := make(bson.M)
				if err := cursor.Decode(&namespace); err != nil {
					log.WithError(err).Error("Failed to decode namespace document")

					return nil, err
				}

				if members, ok := namespace["members"].(bson.A); ok {
					updatedMembers := make(bson.A, 0)
					for _, m := range members {
						if member, ok := m.(bson.M); ok {
							if member["role"] != "owner" {
								invitations = append(
									invitations,
									bson.M{
										"tenant_id":         namespace["tenant_id"],
										"user_id":           member["id"],
										"invited_by":        namespace["owner"],
										"role":              member["role"],
										"status":            member["status"],
										"created_at":        member["added_at"],
										"updated_at":        member["added_at"],
										"status_updated_at": member["added_at"],
										"expires_at":        member["expires_at"],
										"invitations":       1,
									},
								)
							}

							if member["status"] == "accepted" {
								member := bson.M{"id": member["id"], "added_at": member["added_at"], "role": member["role"]}
								updatedMembers = append(updatedMembers, member)
							}
						}
					}

					namespace["members"] = updatedMembers
					namespacesToUpdate = append(namespacesToUpdate, namespace)
				}
			}

			if err := cursor.Err(); err != nil {
				log.WithError(err).Error("Cursor error while iterating namespaces")

				return nil, err
			}

			if len(invitations) > 0 {
				if _, err = db.Collection("membership_invitations").InsertMany(sCtx, invitations); err != nil {
					log.WithError(err).Error("Failed to insert membership invitations")

					return nil, err
				}

				log.WithField("count", len(invitations)).Info("Successfully migrated member invitations to membership_invitations collection")
			} else {
				log.Info("No member invitations found to migrate")
			}

			for _, ns := range namespacesToUpdate {
				nsID := ns["_id"]
				if _, err = db.Collection("namespaces").ReplaceOne(sCtx, bson.M{"_id": nsID}, ns); err != nil {
					log.WithError(err).Error("Failed to update namespace")

					return nil, err
				}
			}

			if len(namespacesToUpdate) > 0 {
				log.WithField("count", len(namespacesToUpdate)).Info("Successfully updated namespaces with cleaned members")
			}

			return nil, nil
		})

		return err
	}),

	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   118,
			"action":    "Down",
		}).Warning("Migration down is not implemented - this migration cannot be reversed safely")

		return nil
	}),
}
