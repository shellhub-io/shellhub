package services

import (
	req "github.com/shellhub-io/shellhub/pkg/api/internalclient"
)

type BillingInterface interface {
	BillingEvaluate(req.Client, string) (bool, error)
	BillingReport(req.Client, string, string) error
}

// BillingEvaluate evaluate in the billing service if the namespace can create accept more devices.
func (s *service) BillingEvaluate(client req.Client, tenant string) (bool, error) {
	evaluation, _, err := client.BillingEvaluate(tenant)
	if err != nil {
		return false, ErrEvaluate
	}

	return evaluation.CanAccept, nil
}

const (
	ReportDeviceAccept    = "device_accept"
	ReportNamespaceDelete = "namespace_delete"
)

func (s *service) BillingReport(client req.Client, tenant string, action string) error {
	status, err := client.BillingReport(tenant, action)
	if err != nil {
		return err
	}

	switch status {
	case 200:
		return nil
	case 402:
		return ErrPaymentRequired
	default:
		return ErrReport
	}
}
