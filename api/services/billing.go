package services

import (
	"context"
	"errors"

	req "github.com/shellhub-io/shellhub/pkg/api/internalclient"
)

type BillingInterface interface {
	BillingEvaluate(ctx context.Context, client req.Client, tenant string) (bool, error)
	BillingReport(ctx context.Context, client req.Client, tenant string, action string) error
}

// BillingEvaluate evaluate in the billing service if the namespace can create accept more devices.
func (s *service) BillingEvaluate(ctx context.Context, client req.Client, tenant string) (bool, error) {
	evaluation, err := client.BillingEvaluate(ctx, tenant)
	if err != nil {
		return false, ErrEvaluate
	}

	return evaluation.CanAccept, nil
}

const (
	ReportDeviceAccept    = "device_accept"
	ReportNamespaceDelete = "namespace_delete"
)

func (s *service) BillingReport(ctx context.Context, client req.Client, tenant string, action string) error {
	if err := client.BillingReport(ctx, tenant, action); err != nil {
		var e *req.Error
		if ok := errors.As(err, &e); !ok {
			return ErrReport
		}

		switch e.Code {
		case 402:
			return ErrPaymentRequired
		default:
			return ErrReport
		}
	}

	return nil
}
