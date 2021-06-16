package mongo

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
)

func (s *Store) BillingUpdateCustomer(ctx context.Context, namespace *models.Namespace, custID string) error {
	if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": namespace.TenantID}, bson.M{"$set": bson.M{"billing.customer_id": custID}}); err != nil {
		return fromMongoError(err)
	}

	return nil
}

func (s *Store) BillingUpdatePaymentID(ctx context.Context, namespace *models.Namespace, paymentID string) error {
	if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": namespace.TenantID}, bson.M{"$set": bson.M{"billing.payment_method_id": paymentID}}); err != nil {
		return fromMongoError(err)
	}

	return nil
}

func (s *Store) BillingUpdateSubscription(ctx context.Context, namespace *models.Namespace, billing *models.Billing) error {
	if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": namespace.TenantID}, bson.M{"$set": bson.M{"billing.subscription_id": billing.SubscriptionID, "billing.current_period_end": billing.CurrentPeriodEnd, "billing.price_id": billing.PriceID, "billing.active": billing.Active}}); err != nil {
		return fromMongoError(err)
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

func (s *Store) BillingUpdateSubscriptionPeriodEnd(ctx context.Context, subscriptionID string, periodEnd time.Time) error {
	ns := new(models.Namespace)
	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"billing.subscription_id": subscriptionID}).Decode(&ns); err != nil {
		return fromMongoError(err)
	}

	if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"billing.subscription_id": subscriptionID}, bson.M{"$set": bson.M{"billing.current_period_end": periodEnd}}); err != nil {
		return fromMongoError(err)
	}

	return nil
}

func (s *Store) BillingUpdateDeviceLimit(ctx context.Context, subscriptionID string, newLimit int) (*models.Namespace, error) {
	ns := new(models.Namespace)
	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"billing.subscription_id": subscriptionID}).Decode(&ns); err != nil {
		return nil, fromMongoError(err)
	}

	if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"billing.subscription_id": subscriptionID}, bson.M{"$set": bson.M{"max_devices": newLimit}}); err != nil {
		return nil, fromMongoError(err)
	}

	return ns, nil
}

func (s *Store) BillingRemoveInstance(ctx context.Context, subsID string) error {
	ns := new(models.Namespace)
	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"billing.subscription_id": subsID}).Decode(&ns); err != nil {
		return fromMongoError(err)
	}
	if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": ns.TenantID}, bson.M{"$unset": bson.M{"billing": 1}}); err != nil {
		return fromMongoError(err)
	}

	return nil
}

func (s *Store) BillingDeleteCustomer(ctx context.Context, namespace *models.Namespace) error {
	if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": namespace.TenantID}, bson.M{"$unset": bson.M{"billing.current_period_end": 1, "billing.price_id": 1, "billing.customer_id": 1, "billing.payment_failed": 1, "billing.payment_method_id": 1}}); err != nil {
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
