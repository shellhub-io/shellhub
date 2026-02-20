package internalclient

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type licenseAPI interface {
	LicenseEvaluate(ctx context.Context) (*models.BillingEvaluation, error)
}

func (c *client) LicenseEvaluate(ctx context.Context) (*models.BillingEvaluation, error) {
	eval := new(models.BillingEvaluation)

	resp, err := c.http.
		R().
		SetContext(ctx).
		SetResult(&eval).
		Get(c.config.APIBaseURL + "/internal/license/evaluate")
	if err := HasError(resp, err); err != nil {
		return nil, err
	}

	return eval, nil
}
