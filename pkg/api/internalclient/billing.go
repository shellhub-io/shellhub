package internalclient

import (
	"net/http"

	"github.com/shellhub-io/shellhub/pkg/models"
)

// billingAPI defines methods for interacting with billing-related functionality.
type billingAPI interface {
	// BillingReport sends a billing report for the specified tenant and action.
	// It returns the HTTP status code of the response and an error, if any.
	BillingReport(tenant string, action string) (int, error)

	// BillingEvaluate evaluates the billing status for the specified tenant.
	// It returns the billing evaluation result, the HTTP status code of the response, and an error, if any.
	BillingEvaluate(tenantID string) (*models.BillingEvaluation, int, error)
}

func (c *client) BillingReport(tenant string, action string) (int, error) {
	res, err := c.http.
		R().
		SetHeader("X-Tenant-ID", tenant).
		SetQueryParam("action", action).
		Post("http://billing-api:8080/internal/billing/report")
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return res.StatusCode(), nil
}

func (c *client) BillingEvaluate(tenantID string) (*models.BillingEvaluation, int, error) {
	eval := new(models.BillingEvaluation)

	resp, err := c.http.
		R().
		SetHeader("X-Tenant-ID", tenantID).
		SetResult(&eval).
		Post("http://billing-api:8080/internal/billing/evaluate")
	if err != nil {
		return eval, resp.StatusCode(), err
	}

	return eval, resp.StatusCode(), nil
}
