package mongo

import (
	"context"
	"strings"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

func (s *Store) BillingUpdateInstance(ctx context.Context, namespace *models.Namespace, billing *models.Billing) error {
	if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": namespace.TenantID}, bson.M{"$set": bson.M{"billing": billing}}); err != nil {
		return fromMongoError(err)
	}

	if err := s.cache.Delete(ctx, strings.Join([]string{"namespace", namespace.TenantID}, "/")); err != nil {
		logrus.Error(err)
	}

	return nil
}

func (s *Store) BillingUpdatePaymentFailed(ctx context.Context, subscriptionID string, set bool, pf *models.PaymentFailed) (*models.Namespace, error) {
	ns := new(models.Namespace)
	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"billing.subscription_id": subscriptionID}).Decode(&ns); err != nil {
		return nil, fromMongoError(err)
	}

	if set {
		if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": ns.TenantID}, bson.M{"$set": bson.M{"billing.payment_failed": pf}}); err != nil {
			return nil, fromMongoError(err)
		}

		return s.NamespaceGet(ctx, ns.TenantID)
	}

	if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": ns.TenantID}, bson.M{"$unset": bson.M{"billing.payment_failed": 1}}); err != nil {
		return nil, fromMongoError(err)
	}

	return s.NamespaceGet(ctx, ns.TenantID)
}

func (s *Store) BillingUpdateDeviceLimit(ctx context.Context, tenantID string, newLimit int) (*models.Namespace, error) {
	ns := new(models.Namespace)
	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": tenantID}).Decode(&ns); err != nil {
		return nil, fromMongoError(err)
	}

	if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$set": bson.M{"max_devices": newLimit}}); err != nil {
		return nil, fromMongoError(err)
	}

	return ns, nil
}

func (s *Store) BillingRemoveInstance(ctx context.Context, custID string) error {
	ns := new(models.Namespace)
	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"billing.customer_id": custID}).Decode(&ns); err != nil {
		return fromMongoError(err)
	}
	if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": ns.TenantID}, bson.M{"$unset": bson.M{"billing": 1}}); err != nil {
		return fromMongoError(err)
	}

	return nil
}

func (s *Store) BillingDeleteSubscription(ctx context.Context, tenantID string) error {
	if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$set": bson.M{"billing.active": false}}); err != nil {
		return fromMongoError(err)
	}

	return nil
}

func (s *Store) BillingActiveInstances(ctx context.Context) ([]models.Namespace, int, error) {
	filter := bson.M{
		"$and": []bson.M{
			{
				"billing": bson.M{
					"$ne": nil,
				},
			},
			{
				"billing.active": true,
			},
		},
	}

	instances := make([]models.Namespace, 0)

	cursor, err := s.db.Collection("namespaces").Find(ctx, filter, nil)
	if err != nil {
		return nil, 0, fromMongoError(err)
	}

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		inst := new(models.Namespace)
		err = cursor.Decode(&inst)
		if err != nil {
			return instances, 0, err
		}

		countDevice, err := s.db.Collection("devices").CountDocuments(ctx, bson.M{"tenant_id": inst.TenantID, "status": "accepted"})
		if err != nil {
			return instances, 0, err
		}

		inst.DevicesCount = int(countDevice)

		instances = append(instances, *inst)
	}

	return instances, len(instances), nil
}
