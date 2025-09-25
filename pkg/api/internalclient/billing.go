package internalclient

import (
	"context"
	"errors"

	"github.com/shellhub-io/shellhub/pkg/models"
)

// billingAPI defines methods for interacting with billing-related functionality.
type billingAPI interface {
	// BillingReport sends a billing report for the specified tenant and action.
	// It returns the HTTP status code of the response and an error, if any.
	BillingReport(ctx context.Context, tenant string, action string) (int, error)

	// BillingEvaluate evaluates the billing status for the specified tenant.
	// It returns the billing evaluation result, the HTTP status code of the response, and an error, if any.
	BillingEvaluate(ctx context.Context, tenantID string) (*models.BillingEvaluation, int, error)
}

var ErrBillingRequestFailed = errors.New("billing request failed")

func (c *client) BillingReport(ctx context.Context, tenant string, action string) (int, error) {
	res, err := c.http.
		R().
		SetContext(ctx).
		SetHeader("X-Tenant-ID", tenant).
		SetQueryParam("action", action).
		Post(c.Config.EnterpriseBaseURL + "/internal/billing/report")
	if err != nil {
		// TODO: It shouldn't return the status code.
		return res.StatusCode(), errors.Join(ErrBillingRequestFailed, err)
	}

	return res.StatusCode(), nil
}

func (c *client) BillingEvaluate(ctx context.Context, tenantID string) (*models.BillingEvaluation, int, error) {
	eval := new(models.BillingEvaluation)

	resp, err := c.http.
		R().
		SetContext(ctx).
		SetHeader("X-Tenant-ID", tenantID).
		SetResult(&eval).
		Post(c.Config.EnterpriseBaseURL + "/internal/billing/evaluate")
	if err != nil {
		return nil, resp.StatusCode(), errors.Join(ErrBillingRequestFailed, err)
	}

	return eval, resp.StatusCode(), nil
}
