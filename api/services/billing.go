package services

import (
	"context"
	"errors"

	"github.com/shellhub-io/shellhub/pkg/models"
)

// BillingService is the interface for enterprise/cloud billing service.
// When present, it's called directly instead of via HTTP.
type BillingService interface {
	Evaluate(ctx context.Context, tenant string) (*models.BillingEvaluation, error)
	Report(ctx context.Context, tenant string, action string) error
}

type BillingInterface interface {
	BillingEvaluate(ctx context.Context, tenant string) (bool, error)
	BillingReport(ctx context.Context, tenant string, action string) error
}

// BillingEvaluate evaluate in the billing service if the namespace can create accept more devices.
func (s *service) BillingEvaluate(ctx context.Context, tenant string) (bool, error) {
	if s.billingService == nil {
		return false, errors.New("billing service not available")
	}

	evaluation, err := s.billingService.Evaluate(ctx, tenant)
	if err != nil {
		return false, ErrEvaluate
	}

	return evaluation.CanAccept, nil
}

const (
	ReportDeviceAccept    = "device_accept"
	ReportNamespaceDelete = "namespace_delete"
)

func (s *service) BillingReport(ctx context.Context, tenant string, action string) error {
	if s.billingService == nil {
		return errors.New("billing service not available")
	}

	if err := s.billingService.Report(ctx, tenant, action); err != nil {
		// The adapter already maps billing errors to appropriate types
		// (ErrPaymentRequired for subscription issues, ErrReport for others)
		return err
	}

	return nil
}
