package store

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type BillingStore interface {
	BillingUpdateCustomer(ctx context.Context, namespace *models.Namespace, custID string) error
	BillingUpdatePaymentID(ctx context.Context, namespace *models.Namespace, paymentID string) error
	BillingUpdateSubscription(ctx context.Context, namespace *models.Namespace, billing *models.Billing) error
	BillingUpdatePaymentFailed(ctx context.Context, subscriptionID string, set bool, pf *models.PaymentFailed) (*models.Namespace, error)
	BillingUpdateSubscriptionPeriodEnd(ctx context.Context, subscriptionID string, periodEnd time.Time) error
	BillingUpdateDeviceLimit(ctx context.Context, subscriptionID string, newLimit int) (*models.Namespace, error)
	BillingDeleteCustomer(ctx context.Context, namespace *models.Namespace) error
	BillingDeleteSubscription(ctx context.Context, tenantID string) error
	BillingRemoveInstance(ctx context.Context, subsID string) error
}
