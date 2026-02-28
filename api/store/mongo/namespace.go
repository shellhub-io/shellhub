package mongo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *Store) NamespaceList(ctx context.Context, opts ...store.QueryOption) ([]models.Namespace, int, error) {
	query := []bson.M{}

	// Only match for the respective tenant if requested
	if id := gateway.IDFromContext(ctx); id != nil {
		user, err := s.UserResolve(ctx, store.UserIDResolver, id.ID)
		if err != nil {
			return nil, 0, err
		}

		query = append(query, bson.M{
			"$match": bson.M{
				"members": bson.M{
					"$elemMatch": bson.M{
						"id": user.ID,
					},
				},
			},
		})
	}

	query = append(query,
		bson.M{
			"$addFields": bson.M{
				"members": bson.M{
					"$map": bson.M{
						"input": "$members",
						"as":    "member",
						"in": bson.M{
							"$mergeObjects": bson.A{
								"$$member",
								bson.M{
									"id": bson.M{
										"$toObjectId": "$$member.id",
									},
								},
							},
						},
					},
				},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "members.id",
				"foreignField": "_id",
				"as":           "userDetails",
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "user_invitations",
				"localField":   "members.id",
				"foreignField": "_id",
				"as":           "invitationDetails",
			},
		},
		bson.M{
			"$addFields": bson.M{
				"members": bson.M{
					"$map": bson.M{
						"input": "$members",
						"as":    "member",
						"in": bson.M{
							"$let": bson.M{
								"vars": bson.M{
									"userDoc": bson.M{
										"$arrayElemAt": bson.A{
											bson.M{
												"$filter": bson.M{
													"input": "$userDetails",
													"cond": bson.M{
														"$eq": bson.A{"$$this._id", "$$member.id"},
													},
												},
											},
											0,
										},
									},
									"inviteDoc": bson.M{
										"$arrayElemAt": bson.A{
											bson.M{
												"$filter": bson.M{
													"input": "$invitationDetails",
													"cond": bson.M{
														"$eq": bson.A{"$$this._id", "$$member.id"},
													},
												},
											},
											0,
										},
									},
								},
								"in": bson.M{
									"$mergeObjects": bson.A{
										"$$member",
										bson.M{
											"email": bson.M{
												"$ifNull": bson.A{
													"$$userDoc.email",
													"$$inviteDoc.email",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		bson.M{
			"$unset": bson.A{"userDetails", "invitationDetails"},
		},
	)

	for _, opt := range opts {
		if err := opt(context.WithValue(ctx, "query", &query)); err != nil {
			return nil, 0, err
		}
	}

	count, err := CountAllMatchingDocuments(ctx, s.db.Collection("namespaces"), query)
	if err != nil {
		return nil, 0, err
	}

	namespaces := make([]models.Namespace, 0)
	cursor, err := s.db.Collection("namespaces").Aggregate(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		namespace := new(models.Namespace)
		if err := cursor.Decode(namespace); err != nil {
			return namespaces, count, err
		}

		namespaces = append(namespaces, *namespace)
	}

	return namespaces, count, err
}

func (s *Store) NamespaceResolve(ctx context.Context, resolver store.NamespaceResolver, value string) (*models.Namespace, error) {
	namespace := new(models.Namespace)
	if _ = s.cache.Get(ctx, "namespace"+"/"+value, namespace); namespace != nil && namespace.TenantID != "" {
		return namespace, nil
	}

	matchStage := bson.M{}
	switch resolver {
	case store.NamespaceTenantIDResolver:
		matchStage["tenant_id"] = value
	case store.NamespaceNameResolver:
		matchStage["name"] = value
	}

	query := []bson.M{
		{
			"$match": matchStage,
		},
		{
			"$addFields": bson.M{
				"members": bson.M{
					"$map": bson.M{
						"input": "$members",
						"as":    "member",
						"in": bson.M{
							"$mergeObjects": bson.A{
								"$$member",
								bson.M{
									"id": bson.M{
										"$toObjectId": "$$member.id",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "members.id",
				"foreignField": "_id",
				"as":           "userDetails",
			},
		},
		{
			"$lookup": bson.M{
				"from":         "user_invitations",
				"localField":   "members.id",
				"foreignField": "_id",
				"as":           "invitationDetails",
			},
		},
		{
			"$addFields": bson.M{
				"members": bson.M{
					"$map": bson.M{
						"input": "$members",
						"as":    "member",
						"in": bson.M{
							"$let": bson.M{
								"vars": bson.M{
									"userDoc": bson.M{
										"$arrayElemAt": bson.A{
											bson.M{
												"$filter": bson.M{
													"input": "$userDetails",
													"cond": bson.M{
														"$eq": bson.A{"$$this._id", "$$member.id"},
													},
												},
											},
											0,
										},
									},
									"inviteDoc": bson.M{
										"$arrayElemAt": bson.A{
											bson.M{
												"$filter": bson.M{
													"input": "$invitationDetails",
													"cond": bson.M{
														"$eq": bson.A{"$$this._id", "$$member.id"},
													},
												},
											},
											0,
										},
									},
								},
								"in": bson.M{
									"$mergeObjects": bson.A{
										"$$member",
										bson.M{
											"email": bson.M{
												"$ifNull": bson.A{
													"$$userDoc.email",
													"$$inviteDoc.email",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			"$unset": bson.A{"userDetails", "invitationDetails"},
		},
	}

	cursor, err := s.db.Collection("namespaces").Aggregate(ctx, query)
	if err != nil {
		return nil, FromMongoError(err)
	}
	defer cursor.Close(ctx)

	cursor.Next(ctx)

	namespace = nil
	if err := cursor.Decode(&namespace); err != nil {
		return nil, FromMongoError(err)
	}

	if err := s.cache.Set(ctx, "namespace"+"/"+value, namespace, time.Minute); err != nil {
		log.Error(err)
	}

	return namespace, nil
}

func (s *Store) NamespaceGetPreferred(ctx context.Context, userID string) (*models.Namespace, error) {
	filter := bson.M{"members.id": userID}

	if user, _ := s.UserResolve(ctx, store.UserIDResolver, userID); user != nil {
		if user.Preferences.PreferredNamespace != "" {
			filter["tenant_id"] = user.Preferences.PreferredNamespace
		}
	}

	ns := new(models.Namespace)
	if err := s.db.Collection("namespaces").FindOne(ctx, filter).Decode(ns); err != nil {
		return nil, FromMongoError(err)
	}

	return ns, nil
}

func (s *Store) NamespaceCreate(ctx context.Context, namespace *models.Namespace) (string, error) {
	namespace.CreatedAt = clock.Now()
	if namespace.TenantID == "" {
		namespace.TenantID = uuid.Generate()
	}

	if _, err := s.db.Collection("namespaces").InsertOne(ctx, namespace); err != nil {
		return "", err
	}

	return namespace.TenantID, nil
}

func (s *Store) NamespaceConflicts(ctx context.Context, target *models.NamespaceConflicts) ([]string, bool, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"$or": []bson.M{
					{"name": target.Name},
				},
			},
		},
	}

	cursor, err := s.db.Collection("namespaces").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, false, FromMongoError(err)
	}
	defer cursor.Close(ctx)

	namespace := new(models.NamespaceConflicts)
	conflicts := make([]string, 0)
	for cursor.Next(ctx) {
		if err := cursor.Decode(&namespace); err != nil {
			return nil, false, FromMongoError(err)
		}

		if namespace.Name == target.Name {
			conflicts = append(conflicts, "name")
		}
	}

	return conflicts, len(conflicts) > 0, nil
}

func (s *Store) NamespaceUpdate(ctx context.Context, namespace *models.Namespace) error {
	res, err := s.db.
		Collection("namespaces").
		UpdateOne(ctx, bson.M{"tenant_id": namespace.TenantID}, bson.M{"$set": namespace})
	if err != nil {
		return FromMongoError(err)
	}

	if res.MatchedCount < 1 {
		return store.ErrNoDocuments
	}

	if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", namespace.TenantID}, "/")); err != nil {
		log.Error(err)
	}

	return nil
}

func (s *Store) NamespaceDelete(ctx context.Context, namespace *models.Namespace) error {
	deletedCount, err := s.NamespaceDeleteMany(ctx, []string{namespace.TenantID})
	switch {
	case err != nil:
		return err
	case deletedCount < 1:
		return store.ErrNoDocuments
	default:
		return nil
	}
}

func (s *Store) NamespaceDeleteMany(ctx context.Context, tenantIDs []string) (int64, error) {
	mongoSession, err := s.db.Client().StartSession()
	if err != nil {
		return 0, FromMongoError(err)
	}

	defer mongoSession.EndSession(ctx)

	fn := func(sessCtx mongo.SessionContext) (any, error) {
		r, err := s.db.Collection("namespaces").DeleteMany(sessCtx, bson.M{"tenant_id": bson.M{"$in": tenantIDs}})
		if err != nil {
			return 0, FromMongoError(err)
		}

		for _, tenantID := range tenantIDs {
			if err := s.cache.Delete(sessCtx, strings.Join([]string{"namespace", tenantID}, "/")); err != nil {
				log.Error(err)
			}
		}

		collections := []string{"devices", "sessions", "public_keys", "api_keys"}
		for _, collection := range collections {
			if _, err := s.db.Collection(collection).DeleteMany(sessCtx, bson.M{"tenant_id": bson.M{"$in": tenantIDs}}); err != nil {
				return 0, FromMongoError(err)
			}
		}

		_, err = s.db.
			Collection("users").
			UpdateMany(sessCtx, bson.M{"preferences.preferred_namespace": bson.M{"$in": tenantIDs}}, bson.M{"$set": bson.M{"preferences.preferred_namespace": ""}})
		if err != nil {
			return 0, FromMongoError(err)
		}

		return r.DeletedCount, nil
	}

	deletedCount, err := mongoSession.WithTransaction(ctx, fn)

	return deletedCount.(int64), err
}

func (s *Store) NamespaceSyncDeviceCounts(ctx context.Context) error {
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

	cursor, err := s.db.Collection("devices").Aggregate(ctx, pipeline)
	if err != nil {
		return FromMongoError(err)
	}
	defer cursor.Close(ctx)

	tenantIDs := make([]string, 0)

	for cursor.Next(ctx) {
		var result struct {
			ID     string `bson:"_id"`
			Counts []struct {
				Status string `bson:"status"`
				Count  int64  `bson:"count"`
			} `bson:"counts"`
		}

		if err := cursor.Decode(&result); err != nil {
			return FromMongoError(err)
		}

		tenantIDs = append(tenantIDs, result.ID)

		updateDoc := bson.M{
			"devices_accepted_count": int64(0),
			"devices_pending_count":  int64(0),
			"devices_rejected_count": int64(0),
			"devices_removed_count":  int64(0),
		}

		for _, c := range result.Counts {
			switch c.Status {
			case "accepted":
				updateDoc["devices_accepted_count"] = c.Count
			case "pending":
				updateDoc["devices_pending_count"] = c.Count
			case "rejected":
				updateDoc["devices_rejected_count"] = c.Count
			case "removed":
				updateDoc["devices_removed_count"] = c.Count
			}
		}

		if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": result.ID}, bson.M{"$set": updateDoc}); err != nil {
			return FromMongoError(err)
		}

		if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", result.ID}, "/")); err != nil {
			log.Error(err)
		}
	}

	if err := cursor.Err(); err != nil {
		return FromMongoError(err)
	}

	zeroDoc := bson.M{
		"$set": bson.M{
			"devices_accepted_count": int64(0),
			"devices_pending_count":  int64(0),
			"devices_rejected_count": int64(0),
			"devices_removed_count":  int64(0),
		},
	}

	if _, err := s.db.Collection("namespaces").UpdateMany(ctx, bson.M{"tenant_id": bson.M{"$nin": tenantIDs}}, zeroDoc); err != nil {
		return FromMongoError(err)
	}

	return nil
}

func (s *Store) NamespaceIncrementDeviceCount(ctx context.Context, tenantID string, status models.DeviceStatus, count int64) error {
	update := bson.M{
		"$inc": bson.M{
			fmt.Sprintf("devices_%s_count", string(status)): count,
		},
	}

	r, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": tenantID}, update)
	if err != nil {
		return FromMongoError(err)
	}

	if r.MatchedCount == 0 {
		return store.ErrNoDocuments
	}

	if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", tenantID}, "/")); err != nil {
		log.Error(err)
	}

	return nil
}
