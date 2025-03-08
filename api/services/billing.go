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
	return false, nil
}

const (
	ReportDeviceAccept    = "device_accept"
	ReportNamespaceDelete = "namespace_delete"
)

func (s *service) BillingReport(client req.Client, tenant string, action string) error {
	return nil
}
