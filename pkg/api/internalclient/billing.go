package internalclient

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

// billingAPI defines methods for interacting with billing-related functionality.
type billingAPI interface {
	// BillingReport sends a billing report for the specified tenant and action.
	// It returns an error, if any.
	BillingReport(ctx context.Context, tenant string, action string) error

	// BillingEvaluate evaluates the billing status for the specified tenant.
	// It returns the billing evaluation result and an error, if any.
	BillingEvaluate(ctx context.Context, tenantID string) (*models.BillingEvaluation, error)
}

func (c *client) BillingReport(ctx context.Context, tenant string, action string) error {
	res, err := c.http.
		R().
		SetContext(ctx).
		SetHeader("X-Tenant-ID", tenant).
		SetQueryParam("action", action).
		Post(c.config.APIBaseURL + "/internal/billing/report")

	return HasError(res, err)
}

func (c *client) BillingEvaluate(ctx context.Context, tenantID string) (*models.BillingEvaluation, error) {
	eval := new(models.BillingEvaluation)

	resp, err := c.http.
		R().
		SetContext(ctx).
		SetHeader("X-Tenant-ID", tenantID).
		SetResult(&eval).
		Post(c.config.APIBaseURL + "/internal/billing/evaluate")
	if err := HasError(resp, err); err != nil {
		return nil, err
	}

	return eval, nil
}
