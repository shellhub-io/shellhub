package mongo

import (
	"context"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mongo/queries"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *Store) NamespaceList(ctx context.Context, paginator query.Paginator, filters query.Filters, export bool) ([]models.Namespace, int, error) {
	query := []bson.M{}
	queryMatch, err := queries.FromFilters(&filters)
	if err != nil {
		return nil, 0, FromMongoError(err)
	}
	query = append(query, queryMatch...)

	if export {
		query = []bson.M{
			{
				"$lookup": bson.M{
					"from":         "devices",
					"localField":   "tenant_id",
					"foreignField": "tenant_id",
					"as":           "devices",
				},
			},
			{
				"$addFields": bson.M{
					"devices": bson.M{"$size": "$devices"},
				},
			},
			{
				"$lookup": bson.M{
					"from":         "sessions",
					"localField":   "devices.uid",
					"foreignField": "device_uid",
					"as":           "sessions",
				},
			},
			{
				"$addFields": bson.M{
					"sessions": bson.M{"$size": "$sessions"},
				},
			},
		}
	}

	if len(queryMatch) > 0 {
		query = append(query, queryMatch...)
	}

	// Only match for the respective tenant if requested
	if id := gateway.IDFromContext(ctx); id != nil {
		user, _, err := s.UserGetByID(ctx, id.ID, false)
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

	queryCount := query
	queryCount = append(queryCount, bson.M{"$count": "count"})
	count, err := AggregateCount(ctx, s.db.Collection("namespaces"), queryCount)
	if err != nil {
		return nil, 0, err
	}

	query = append(query, queries.FromPaginator(&paginator)...)

	namespaces := make([]models.Namespace, 0)
	cursor, err := s.db.Collection("namespaces").Aggregate(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		namespace := new(models.Namespace)
		err = cursor.Decode(&namespace)
		if err != nil {
			return namespaces, count, err
		}

		countDevice, err := s.db.Collection("devices").CountDocuments(ctx, bson.M{"tenant_id": namespace.TenantID, "status": "accepted"})
		if err != nil {
			return namespaces, 0, err
		}

		namespace.DevicesCount = int(countDevice)

		namespaces = append(namespaces, *namespace)
	}

	return namespaces, count, err
}

func (s *Store) NamespaceGet(ctx context.Context, tenantID string) (*models.Namespace, error) {
	var ns *models.Namespace

	if err := s.cache.Get(ctx, strings.Join([]string{"namespace", tenantID}, "/"), &ns); err != nil {
		logrus.Error(err)
	}

	if ns != nil {
		goto count
	}

	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": tenantID}).Decode(&ns); err != nil {
		return ns, FromMongoError(err)
	}

	if err := s.cache.Set(ctx, strings.Join([]string{"namespace", tenantID}, "/"), ns, time.Minute); err != nil {
		logrus.Error(err)
	}

count:
	countDevice, err := s.db.Collection("devices").CountDocuments(ctx, bson.M{"tenant_id": tenantID, "status": "accepted"})
	if err != nil {
		return nil, FromMongoError(err)
	}

	ns.DevicesCount = int(countDevice)

	return ns, nil
}

func (s *Store) NamespaceGetByName(ctx context.Context, name string) (*models.Namespace, error) {
	var ns *models.Namespace

	if err := s.cache.Get(ctx, strings.Join([]string{"namespace", name}, "/"), &ns); err != nil {
		logrus.Error(err)
	}

	if ns != nil {
		return ns, nil
	}

	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"name": name}).Decode(&ns); err != nil {
		return nil, FromMongoError(err)
	}

	return ns, nil
}

func (s *Store) NamespaceCreate(ctx context.Context, namespace *models.Namespace) (*models.Namespace, error) {
	session, err := s.db.Client().StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	if _, err := session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		_, err := s.db.Collection("namespaces").InsertOne(sessCtx, namespace)
		if err != nil {
			return nil, err
		}

		objID, err := primitive.ObjectIDFromHex(namespace.Owner)
		if err != nil {
			return nil, FromMongoError(err)
		}

		if _, err := s.db.Collection("users").UpdateOne(sessCtx, bson.M{"_id": objID}, bson.M{"$inc": bson.M{"namespaces": 1}}); err != nil {
			return nil, FromMongoError(err)
		}

		return nil, nil
	}); err != nil {
		return nil, err
	}

	return namespace, err
}

func (s *Store) NamespaceDelete(ctx context.Context, tenantID string) error {
	session, err := s.db.Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	if _, err := session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		ns, err := s.NamespaceGet(ctx, tenantID)
		if err != nil {
			return nil, err
		}

		if _, err := s.db.Collection("namespaces").DeleteOne(sessCtx, bson.M{"tenant_id": tenantID}); err != nil {
			return nil, FromMongoError(err)
		}

		if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", tenantID}, "/")); err != nil {
			logrus.Error(err)
		}

		collections := []string{"devices", "sessions", "connected_devices", "firewall_rules", "public_keys", "recorded_sessions"}
		for _, collection := range collections {
			if _, err := s.db.Collection(collection).DeleteMany(sessCtx, bson.M{"tenant_id": tenantID}); err != nil {
				return nil, FromMongoError(err)
			}
		}

		objID, err := primitive.ObjectIDFromHex(ns.Owner)
		if err != nil {
			return nil, FromMongoError(err)
		}

		if _, err := s.db.Collection("users").UpdateOne(sessCtx, bson.M{"_id": objID}, bson.M{"$inc": bson.M{"namespaces": -1}}); err != nil {
			return nil, FromMongoError(err)
		}

		return nil, nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *Store) NamespaceEdit(ctx context.Context, tenant string, changes *models.NamespaceChanges) error {
	res, err := s.db.
		Collection("namespaces").
		UpdateOne(ctx, bson.M{"tenant_id": tenant}, bson.M{"$set": changes})
	if err != nil {
		return FromMongoError(err)
	}

	if res.MatchedCount < 1 {
		return store.ErrNoDocuments
	}

	if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", tenant}, "/")); err != nil {
		logrus.Error(err)
	}

	return nil
}

func (s *Store) NamespaceRename(ctx context.Context, tenantID string, name string) (*models.Namespace, error) {
	if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$set": bson.M{"name": name}}); err != nil {
		return nil, FromMongoError(err)
	}

	if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", tenantID}, "/")); err != nil {
		logrus.Error(err)
	}

	return s.NamespaceGet(ctx, tenantID)
}

func (s *Store) NamespaceUpdate(ctx context.Context, tenantID string, namespace *models.Namespace) error {
	ns, err := s.db.Collection("namespaces").UpdateOne(
		ctx,
		bson.M{
			"tenant_id": tenantID,
		},
		bson.M{
			"$set": bson.M{
				"name":                    namespace.Name,
				"max_devices":             namespace.MaxDevices,
				"settings.session_record": namespace.Settings.SessionRecord,
			},
		},
	)
	if err != nil {
		return FromMongoError(err)
	}

	if ns.MatchedCount < 1 {
		return store.ErrNoDocuments
	}

	if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", tenantID}, "/")); err != nil {
		logrus.Error(err)
	}

	return nil
}

func (s *Store) NamespaceAddMember(ctx context.Context, tenantID string, memberID string, memberRole string) (*models.Namespace, error) {
	result := s.db.Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": tenantID, "members": bson.M{"$elemMatch": bson.M{"id": memberID}}})
	if result.Err() == nil {
		return nil, ErrNamespaceDuplicatedMember
	}

	_, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$addToSet": bson.M{"members": bson.M{"id": memberID, "role": memberRole}}})
	if err != nil {
		return nil, FromMongoError(err)
	}

	if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", tenantID}, "/")); err != nil {
		logrus.Error(err)
	}

	return s.NamespaceGet(ctx, tenantID)
}

func (s *Store) NamespaceRemoveMember(ctx context.Context, tenantID string, memberID string) (*models.Namespace, error) {
	ns, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$pull": bson.M{"members": bson.M{"id": memberID}}})
	if err != nil {
		return nil, FromMongoError(err)
	}

	switch {
	// tenant not found
	case ns.MatchedCount < 1:
		return nil, store.ErrNoDocuments
	// member not found
	case ns.ModifiedCount < 1:
		return nil, ErrUserNotFound
	}

	if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", tenantID}, "/")); err != nil {
		logrus.Error(err)
	}

	return s.NamespaceGet(ctx, tenantID)
}

func (s *Store) NamespaceEditMember(ctx context.Context, tenantID string, memberID string, memberNewRole string) error {
	ns, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": tenantID, "members.id": memberID}, bson.M{"$set": bson.M{"members.$.role": memberNewRole}})
	if err != nil {
		return FromMongoError(err)
	}

	if ns.MatchedCount < 1 {
		return ErrUserNotFound
	}

	if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", tenantID}, "/")); err != nil {
		logrus.Error(err)
	}

	return nil
}

func (s *Store) NamespaceGetFirst(ctx context.Context, id string) (*models.Namespace, error) {
	ns := new(models.Namespace)
	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"members": bson.M{"$elemMatch": bson.M{"id": id}}}).Decode(&ns); err != nil {
		return nil, FromMongoError(err)
	}

	return ns, nil
}

func (s *Store) NamespaceSetSessionRecord(ctx context.Context, sessionRecord bool, tenantID string) error {
	ns, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$set": bson.M{"settings.session_record": sessionRecord}})
	if err != nil {
		return FromMongoError(err)
	}

	if ns.MatchedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (s *Store) NamespaceGetSessionRecord(ctx context.Context, tenantID string) (bool, error) {
	var settings struct {
		Settings *models.NamespaceSettings `json:"settings" bson:"settings"`
	}

	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": tenantID}).Decode(&settings); err != nil {
		return false, FromMongoError(err)
	}

	return settings.Settings.SessionRecord, nil
}
