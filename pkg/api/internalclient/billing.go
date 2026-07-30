package internalclient

import (
	"context"
)

// billingAPI defines methods for interacting with billing-related functionality.
type billingAPI interface {
	// BillingReport sends a billing report for the specified tenant and action.
	// It returns an error, if any.
	BillingReport(ctx context.Context, tenant string, action string) error
}

func (c *client) BillingReport(ctx context.Context, tenant string, action string) error {
	res, err := c.http.
		R().
		SetContext(ctx).
		SetHeader("X-Tenant-ID", tenant).
		SetQueryParam("action", action).
		Post(apiBaseURL + "/internal/billing/report")

	return HasError(res, err)
}
