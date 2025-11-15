package migrations

import (
	"context"

	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration115 = migrate.Migration{
	Version:     115,
	Description: "Migrate user invitations from users collection to user_invitations collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   115,
			"action":    "Up",
		}).Info("Applying migration up")

		session, err := db.Client().StartSession()
		if err != nil {
			return err
		}
		defer session.EndSession(ctx)

		_, err = session.WithTransaction(ctx, func(sCtx mongo.SessionContext) (any, error) {
			cursor, err := db.Collection("users").Find(sCtx, bson.M{"status": "invited"})
			if err != nil {
				log.WithError(err).Error("Failed to find invited users")

				return nil, err
			}

			defer cursor.Close(sCtx)

			invitations := make([]any, 0)
			for cursor.Next(sCtx) {
				user := make(bson.M)
				if err := cursor.Decode(&user); err != nil {
					log.WithError(err).Error("Failed to decode user document")

					return nil, err
				}

				invitations = append(invitations, bson.M{
					"_id":         user["_id"],
					"email":       user["email"],
					"created_at":  user["created_at"],
					"updated_at":  user["created_at"],
					"invitations": 1,
					"status":      "pending",
				})
			}

			if err := cursor.Err(); err != nil {
				log.WithError(err).Error("Cursor error while iterating invited users")

				return nil, err
			}

			if len(invitations) > 0 {
				_, err = db.Collection("user_invitations").InsertMany(sCtx, invitations)
				if err != nil {
					log.WithError(err).Error("Failed to insert invitations")

					return nil, err
				}

				log.WithField("count", len(invitations)).Info("Successfully migrated invitations to user_invitations collection")
			} else {
				log.Info("No invited users found to migrate")
			}

			deleteResult, err := db.Collection("users").DeleteMany(sCtx, bson.M{"status": "invited"})
			if err != nil {
				log.WithError(err).Error("Failed to delete invited users")

				return nil, err
			}

			log.WithField("deleted_count", deleteResult.DeletedCount).Info("Successfully removed invited users from users collection")

			return nil, nil
		})

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   115,
			"action":    "Down",
		}).Info("Applying migration down")

		session, err := db.Client().StartSession()
		if err != nil {
			return err
		}
		defer session.EndSession(ctx)

		_, err = session.WithTransaction(ctx, func(sCtx mongo.SessionContext) (any, error) {
			cursor, err := db.Collection("user_invitations").Find(sCtx, bson.M{})
			if err != nil {
				log.WithError(err).Error("Failed to find user invitations")

				return nil, err
			}
			defer cursor.Close(sCtx)

			var users []any
			for cursor.Next(sCtx) {
				invitation := make(bson.M)
				if err := cursor.Decode(&invitation); err != nil {
					log.WithError(err).Error("Failed to decode invitation document")

					return nil, err
				}

				users = append(users, bson.M{
					"_id":             invitation["_id"],
					"email":           invitation["email"],
					"created_at":      invitation["created_at"],
					"last_login":      nil,
					"status":          "invited",
					"origin":          nil,
					"external_id":     nil,
					"max_namespaces":  nil,
					"name":            nil,
					"username":        nil,
					"recovery_email":  nil,
					"email_marketing": nil,
					"password":        nil,
					"preferences":     bson.M{"preferred_namespace": nil, "auth_methods": nil},
					"mfa":             bson.M{"enabled": nil, "secret": nil, "recovery_codes": nil},
					"admin":           nil,
				})
			}

			if err := cursor.Err(); err != nil {
				log.WithError(err).Error("Cursor error while iterating invitations")

				return nil, err
			}

			if len(users) > 0 {
				_, err = db.Collection("users").InsertMany(sCtx, users)
				if err != nil {
					log.WithError(err).Error("Failed to insert users")

					return nil, err
				}

				log.WithField("count", len(users)).Info("Successfully restored invited users to users collection")
			} else {
				log.Info("No invitations found to revert")
			}

			return nil, nil
		})
		if err != nil {
			return err
		}

		if err := db.Collection("user_invitations").Drop(ctx); err != nil {
			log.WithError(err).Error("Failed to drop user_invitations collection")

			return err
		}

		log.Info("Successfully dropped user_invitations collection")

		return nil
	}),
}
