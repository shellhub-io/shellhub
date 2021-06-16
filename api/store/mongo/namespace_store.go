package mongo

import (
	"context"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

func (s *Store) NamespaceList(ctx context.Context, pagination paginator.Query, filters []models.Filter, export bool) ([]models.Namespace, int, error) {
	query := []bson.M{}
	queryMatch, err := buildFilterQuery(filters)
	if err != nil {
		return nil, 0, fromMongoError(err)
	}

	if len(queryMatch) > 0 {
		query = append(query, queryMatch...)
	}

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
	if id := apicontext.IDFromContext(ctx); id != nil {
		user, _, err := s.UserGetByID(ctx, id.ID, false)
		if err != nil {
			return nil, 0, err
		}

		query = append(query, bson.M{
			"$match": bson.M{
				"members": user.ID,
			},
		})
	}

	queryCount := append(query, bson.M{"$count": "count"})
	count, err := aggregateCount(ctx, s.db.Collection("namespaces"), queryCount)
	if err != nil {
		return nil, 0, err
	}

	if pagination.Page != 0 && pagination.PerPage != 0 {
		query = append(query, buildPaginationQuery(pagination)...)
	}

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
		return ns, fromMongoError(err)
	}

	if err := s.cache.Set(ctx, strings.Join([]string{"namespace", tenantID}, "/"), ns, time.Minute); err != nil {
		logrus.Error(err)
	}

count:
	countDevice, err := s.db.Collection("devices").CountDocuments(ctx, bson.M{"tenant_id": tenantID, "status": "accepted"})
	if err != nil {
		return nil, fromMongoError(err)
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
		return nil, fromMongoError(err)
	}

	return ns, nil
}

func (s *Store) NamespaceCreate(ctx context.Context, namespace *models.Namespace) (*models.Namespace, error) {
	_, err := s.db.Collection("namespaces").InsertOne(ctx, namespace)

	return namespace, err
}

func (s *Store) NamespaceDelete(ctx context.Context, tenantID string) error {
	if _, err := s.db.Collection("namespaces").DeleteOne(ctx, bson.M{"tenant_id": tenantID}); err != nil {
		return fromMongoError(err)
	}

	if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", tenantID}, "/")); err != nil {
		logrus.Error(err)
	}

	collections := []string{"devices", "sessions", "connected_devices", "firewall_rules", "public_keys", "recorded_sessions"}
	for _, collection := range collections {
		if _, err := s.db.Collection(collection).DeleteMany(ctx, bson.M{"tenant_id": tenantID}); err != nil {
			return fromMongoError(err)
		}
	}

	return nil
}

func (s *Store) NamespaceRename(ctx context.Context, tenantID, name string) (*models.Namespace, error) {
	if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$set": bson.M{"name": name}}); err != nil {
		return nil, fromMongoError(err)
	}

	if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", tenantID}, "/")); err != nil {
		logrus.Error(err)
	}

	return s.NamespaceGet(ctx, tenantID)
}

func (s *Store) NamespaceUpdate(ctx context.Context, tenantID string, namespace *models.Namespace) error {
	if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$set": bson.M{"name": namespace.Name, "max_devices": namespace.MaxDevices, "settings.session_record": namespace.Settings.SessionRecord}}); err != nil {
		return fromMongoError(err)
	}

	if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", tenantID}, "/")); err != nil {
		logrus.Error(err)
	}

	return nil
}

func (s *Store) NamespaceAddMember(ctx context.Context, tenantID, id string) (*models.Namespace, error) {
	result, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$addToSet": bson.M{"members": id}})
	if err != nil {
		return nil, fromMongoError(err)
	}

	if result.ModifiedCount == 0 {
		return nil, ErrDuplicateID
	}

	if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", tenantID}, "/")); err != nil {
		logrus.Error(err)
	}

	return s.NamespaceGet(ctx, tenantID)
}

func (s *Store) NamespaceRemoveMember(ctx context.Context, tenantID, id string) (*models.Namespace, error) {
	result, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$pull": bson.M{"members": id}})
	if err != nil {
		return nil, fromMongoError(err)
	}
	if result.ModifiedCount == 0 {
		return nil, ErrUserNotFound
	}

	if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", tenantID}, "/")); err != nil {
		logrus.Error(err)
	}

	return s.NamespaceGet(ctx, tenantID)
}

func (s *Store) NamespaceGetFirst(ctx context.Context, id string) (*models.Namespace, error) {
	ns := new(models.Namespace)
	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"members": id}).Decode(&ns); err != nil {
		return nil, fromMongoError(err)
	}

	return ns, nil
}

func (s *Store) NamespaceSetSessionRecord(ctx context.Context, sessionRecord bool, tenantID string) error {
	if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$set": bson.M{"settings.session_record": sessionRecord}}); err != nil {
		return fromMongoError(err)
	}

	return nil
}

func (s *Store) NamespaceGetSessionRecord(ctx context.Context, tenantID string) (bool, error) {
	var settings struct {
		Settings *models.NamespaceSettings `json:"settings" bson:"settings"`
	}

	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": tenantID}).Decode(&settings); err != nil {
		return false, fromMongoError(err)
	}

	return settings.Settings.SessionRecord, nil
}

func (s *Store) NamespaceUpdateSubscription(ctx context.Context, ns *models.Namespace, billing *models.Billing) error {
	if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": ns.TenantID}, bson.M{"$set": bson.M{"billing.subscription_id": billing.SubscriptionID, "billing.current_period_end": billing.CurrentPeriodEnd, "billing.price_id": billing.PriceID}}); err != nil {
		return fromMongoError(err)
	}

	return nil
}

func (s *Store) NamespaceDeleteSubscription(ctx context.Context, tenantID string) error {
	if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$unset": bson.M{"billing.subscription_id": 1, "billing.current_period_end": 1, "billing.price_id": 1}}); err != nil {
		return fromMongoError(err)
	}
	return nil
}

func (s *Store) NamespaceSetPaymentFailed(ctx context.Context, tenantID string) error {
	ns := new(models.Namespace)
	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": tenantID}).Decode(&ns); err != nil {
		return fromMongoError(err)
	}

	if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$set": bson.M{"billing.payment_failed": true}}); err != nil {
		return fromMongoError(err)
	}
	return nil
}

func (s *Store) NamespaceUpdateDeviceLimit(ctx context.Context, subscriptionID string, newLimit int) error {
	ns := new(models.Namespace)
	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"billing.subscription_id": subscriptionID}).Decode(&ns); err != nil {
		return fromMongoError(err)
	}

	if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"billing.subscription_id": subscriptionID}, bson.M{"$set": bson.M{"max_devices": newLimit, "billing.payment_failed": false}}); err != nil {
		return fromMongoError(err)
	}
	return nil
}

func (s *Store) NamespaceUpdateCustomer(ctx context.Context, namespace *models.Namespace, custID string) error {
	if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": namespace.TenantID}, bson.M{"$set": bson.M{"billing.customer_id": custID}}); err != nil {
		return fromMongoError(err)
	}

	return nil
}

func (s *Store) NamespaceUpdatePaymentID(ctx context.Context, namespace *models.Namespace, paymentID string) error {
	if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": namespace.TenantID}, bson.M{"$set": bson.M{"billing.payment_method_id": paymentID}}); err != nil {
		return fromMongoError(err)
	}

	return nil
}

func (s *Store) NamespaceDeleteCustomer(ctx context.Context, namespace *models.Namespace) error {
	if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": namespace.TenantID}, bson.M{"$unset": bson.M{"billing": 1}}); err != nil {
		return fromMongoError(err)
	}

	return nil
}
