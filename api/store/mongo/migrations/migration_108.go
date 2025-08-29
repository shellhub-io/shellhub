package migrations

import (
	"context"
	"fmt"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration108 = migrate.Migration{
	Version:     108,
	Description: "Refactor tags structure in a separeted collection.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		type device struct {
			UID      string   `bson:"uid"`
			TenantID string   `bson:"tenant_id"`
			Tags     []string `bson:"tags"`
		}

		// Represents either a public key or a firewall rule
		type taggedResource struct {
			ID       primitive.ObjectID `bson:"_id"`
			TenantID string             `bson:"tenant_id"`
			Filter   struct {
				Tags []string `bson:"tags"`
			} `bson:"filter"`
		}

		type tag struct {
			ID primitive.ObjectID `bson:"_id"`
		}

		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   108,
			"action":    "Up",
		}).Info("Applying migration up")

		session, err := db.Client().StartSession()
		if err != nil {
			return err
		}

		defer session.EndSession(ctx)

		_, err = session.WithTransaction(ctx, func(sCtx mongo.SessionContext) (any, error) {
			cursor, err := db.Collection("devices").Find(sCtx, bson.M{"uid": bson.M{"$ne": nil}})
			if err != nil {
				return nil, err
			}

			defer cursor.Close(sCtx)

			tagMapping := make(map[string]map[string]primitive.ObjectID)
			for cursor.Next(sCtx) {
				d := new(device)
				if err := cursor.Decode(d); err != nil {
					return nil, err
				}

				if tagMapping[d.TenantID] == nil {
					tagMapping[d.TenantID] = make(map[string]primitive.ObjectID)
				}

				tagIDs := make([]primitive.ObjectID, 0, len(d.Tags))
				for _, tagName := range d.Tags {
					if id, exists := tagMapping[d.TenantID][tagName]; exists {
						tagIDs = append(tagIDs, id)

						continue
					}

					t := new(tag)
					if err := db.
						Collection("tags").
						FindOneAndUpdate(
							sCtx,
							bson.M{
								"tenant_id": d.TenantID,
								"name":      tagName,
							},
							bson.M{
								"$setOnInsert": bson.M{"created_at": clock.Now(), "updated_at": clock.Now()},
								"$set":         bson.M{"name": tagName, "tenant_id": d.TenantID},
							},
							options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
						).
						Decode(t); err != nil {
						return nil, err
					}

					tagMapping[d.TenantID][tagName] = t.ID
					tagIDs = append(tagIDs, t.ID)
				}

				if _, err := db.
					Collection("devices").
					UpdateOne(
						sCtx,
						bson.M{"uid": d.UID},
						bson.M{"$set": bson.M{"tag_ids": tagIDs}, "$unset": bson.M{"tags": ""}},
					); err != nil {
					return nil, err
				}
			}

			if err := cursor.Err(); err != nil {
				return nil, err
			}

			for _, coll := range []string{"public_keys", "firewall_rules"} {
				collCursor, err := db.Collection(coll).Find(sCtx, bson.M{"filter": bson.M{"$exists": true}})
				if err != nil {
					return nil, err
				}

				defer collCursor.Close(sCtx)

				for collCursor.Next(sCtx) {
					res := new(taggedResource)
					if err := collCursor.Decode(res); err != nil {
						return nil, err
					}

					tagIDs := make([]primitive.ObjectID, 0)
					for _, tagName := range res.Filter.Tags {
						if tagMapping[res.TenantID] == nil || tagMapping[res.TenantID][tagName] == primitive.NilObjectID {
							return nil, fmt.Errorf("[%s] document with ID %s references non-existent tag %q", coll, res.ID, tagName)
						}

						tagIDs = append(tagIDs, tagMapping[res.TenantID][tagName])
					}

					if _, err := db.
						Collection(coll).
						UpdateOne(
							sCtx,
							bson.M{"_id": res.ID},
							bson.M{
								"$set":   bson.M{"filter.tag_ids": tagIDs},
								"$unset": bson.M{"filter.tags": ""},
							},
						); err != nil {
						return nil, err
					}
				}

				if err := collCursor.Err(); err != nil {
					return nil, err
				}
			}

			return nil, nil
		})

		return err
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		return nil
	}),
}
