package mongo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mongo/queries"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *Store) NamespaceList(ctx context.Context, paginator query.Paginator, filters query.Filters, opts ...store.NamespaceQueryOption) ([]models.Namespace, int, error) {
	query := []bson.M{}

	queryMatch, err := queries.FromFilters(&filters)
	if err != nil {
		return nil, 0, FromMongoError(err)
	}
	query = append(query, queryMatch...)

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
						"status": bson.M{
							"$ne": models.MemberStatusPending,
						},
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
		if err := cursor.Decode(namespace); err != nil {
			return namespaces, count, err
		}

		for _, opt := range opts {
			if err := opt(context.WithValue(ctx, "db", s.db), namespace); err != nil { //nolint:revive
				return nil, 0, err
			}
		}

		namespaces = append(namespaces, *namespace)
	}

	return namespaces, count, err
}

func (s *Store) NamespaceGet(ctx context.Context, tenantID string, opts ...store.NamespaceQueryOption) (*models.Namespace, error) {
	var ns *models.Namespace

	if _ = s.cache.Get(ctx, strings.Join([]string{"namespace", tenantID}, "/"), &ns); ns != nil && ns.TenantID != "" {
		goto Opts
	}

	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": tenantID}).Decode(&ns); err != nil {
		return ns, FromMongoError(err)
	}

	if err := s.cache.Set(ctx, strings.Join([]string{"namespace", tenantID}, "/"), ns, time.Minute); err != nil {
		log.Error(err)
	}

Opts:
	for _, opt := range opts {
		if err := opt(context.WithValue(ctx, "db", s.db), ns); err != nil { //nolint:revive
			return nil, err
		}
	}

	return ns, nil
}

func (s *Store) NamespaceGetByName(ctx context.Context, name string, opts ...store.NamespaceQueryOption) (*models.Namespace, error) {
	var ns *models.Namespace

	if _ = s.cache.Get(ctx, strings.Join([]string{"namespace", name}, "/"), &ns); ns != nil && ns.TenantID != "" {
		goto Opts
	}

	if ns != nil {
		return ns, nil
	}

	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"name": name}).Decode(&ns); err != nil {
		return nil, FromMongoError(err)
	}

Opts:
	for _, opt := range opts {
		if err := opt(context.WithValue(ctx, "db", s.db), ns); err != nil { //nolint:revive
			return nil, err
		}
	}

	return ns, nil
}

func (s *Store) NamespaceGetPreferred(ctx context.Context, userID string, opts ...store.NamespaceQueryOption) (*models.Namespace, error) {
	filter := bson.M{"members.id": userID}

	if user, _, _ := s.UserGetByID(ctx, userID, false); user != nil {
		if user.Preferences.PreferredNamespace != "" {
			filter["tenant_id"] = user.Preferences.PreferredNamespace
		}
	}

	ns := new(models.Namespace)
	if err := s.db.Collection("namespaces").FindOne(ctx, filter).Decode(ns); err != nil {
		return nil, FromMongoError(err)
	}

	for _, opt := range opts {
		if err := opt(context.WithValue(ctx, "db", s.db), ns); err != nil { //nolint:revive
			return nil, err
		}
	}

	return ns, nil
}

func (s *Store) NamespaceCreate(ctx context.Context, namespace *models.Namespace) (*models.Namespace, error) {
	namespace.CreatedAt = clock.Now()

	_, err := s.db.Collection("namespaces").InsertOne(ctx, namespace)
	if err != nil {
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
		r, err := s.db.Collection("namespaces").DeleteOne(sessCtx, bson.M{"tenant_id": tenantID})
		if err != nil {
			return nil, FromMongoError(err)
		}

		if r.DeletedCount < 1 {
			return nil, store.ErrNoDocuments
		}

		if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", tenantID}, "/")); err != nil {
			log.Error(err)
		}

		collections := []string{"devices", "sessions", "firewall_rules", "public_keys", "recorded_sessions", "api_keys"}
		for _, collection := range collections {
			if _, err := s.db.Collection(collection).DeleteMany(sessCtx, bson.M{"tenant_id": tenantID}); err != nil {
				return nil, FromMongoError(err)
			}
		}

		_, err = s.db.
			Collection("users").
			UpdateMany(ctx, bson.M{"preferred_namespace": tenantID}, bson.M{"$set": bson.M{"preferred_namespace": ""}})
		if err != nil {
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
		log.Error(err)
	}

	return nil
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
		log.Error(err)
	}

	return nil
}

func (s *Store) NamespaceAddMember(ctx context.Context, tenantID string, member *models.Member) error {
	err := s.db.
		Collection("namespaces").
		FindOne(ctx, bson.M{"tenant_id": tenantID, "members": bson.M{"$elemMatch": bson.M{"id": member.ID}}}).
		Err()
	if err == nil {
		return ErrNamespaceDuplicatedMember
	}

	memberBson := bson.M{
		"id":         member.ID,
		"added_at":   member.AddedAt,
		"expires_at": member.ExpiresAt,
		"role":       member.Role,
		"status":     member.Status,
	}

	res, err := s.db.
		Collection("namespaces").
		UpdateOne(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$addToSet": bson.M{"members": memberBson}})
	if err != nil {
		return FromMongoError(err)
	}

	if res.MatchedCount < 1 {
		return store.ErrNoDocuments
	}

	if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", tenantID}, "/")); err != nil {
		log.Error(err)
	}

	return nil
}

func (s *Store) NamespaceUpdateMember(ctx context.Context, tenantID string, memberID string, changes *models.MemberChanges) error {
	filter := bson.M{"tenant_id": tenantID, "members": bson.M{"$elemMatch": bson.M{"id": memberID}}}
	update := bson.M{}

	if changes.Role != "" {
		update["members.$.role"] = changes.Role
	}

	if changes.Status != "" {
		update["members.$.status"] = changes.Status
	}

	if changes.ExpiresAt != nil {
		update["members.$.expires_at"] = *changes.ExpiresAt
	}

	ns, err := s.db.Collection("namespaces").UpdateOne(ctx, filter, bson.M{"$set": update})
	if err != nil {
		return FromMongoError(err)
	}

	if ns.MatchedCount < 1 {
		return ErrUserNotFound
	}

	if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", tenantID}, "/")); err != nil {
		log.Error(err)
	}

	return nil
}

func (s *Store) NamespaceRemoveMember(ctx context.Context, tenantID string, memberID string) error {
	session, err := s.db.Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	fn := func(_ mongo.SessionContext) (interface{}, error) {
		res, err := s.db.
			Collection("namespaces").
			UpdateOne(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$pull": bson.M{"members": bson.M{"id": memberID}}})
		if err != nil {
			return nil, FromMongoError(err)
		}

		switch {
		case res.MatchedCount < 1: // tenant not found
			return nil, store.ErrNoDocuments
		case res.ModifiedCount < 1: // member not found
			return nil, ErrUserNotFound
		}

		objID, err := primitive.ObjectIDFromHex(memberID)
		if err != nil {
			return nil, err
		}

		_, err = s.db.
			Collection("users").
			UpdateOne(ctx, bson.M{"_id": objID, "preferred_namespace": tenantID}, bson.M{"$set": bson.M{"preferred_namespace": ""}})

		return nil, FromMongoError(err)
	}

	if _, err := session.WithTransaction(ctx, fn); err != nil {
		return err
	}

	if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", tenantID}, "/")); err != nil {
		log.Error(err)
	}

	return nil
}

func (s *Store) NamespaceSetSessionRecord(ctx context.Context, sessionRecord bool, tenantID string) error {
	ns, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$set": bson.M{"settings.session_record": sessionRecord}})
	if err != nil {
		return FromMongoError(err)
	}

	if ns.MatchedCount < 1 {
		return store.ErrNoDocuments
	}

	if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", tenantID}, "/")); err != nil {
		log.Error(err)
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

	return nil
}
