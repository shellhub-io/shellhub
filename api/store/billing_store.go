package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type BillingStore interface {
	BillingUpdateInstance(ctx context.Context, namespace *models.Namespace, billing *models.Billing) error
	BillingUpdatePaymentFailed(ctx context.Context, subscriptionID string, set bool, pf *models.PaymentFailed) (*models.Namespace, error)
	BillingUpdateDeviceLimit(ctx context.Context, tenantID string, newLimit int) (*models.Namespace, error)
	BillingDeleteSubscription(ctx context.Context, tenantID string) error
	BillingRemoveInstance(ctx context.Context, custID string) error
}
