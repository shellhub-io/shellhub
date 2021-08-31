package requests

import (
	req "github.com/shellhub-io/shellhub/pkg/api/client"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	BillingURL = "billing-api"
)

func HasBillingInstance(ns *models.Namespace) bool {
	if ns == nil {
		return false
	}

	if ns.Billing == nil || !ns.Billing.Active || ns.MaxDevices != -1 {
		return false
	}

	return true
}

func HandleResponse(status int, err error) error {
	switch status {
	case 402:
		return nil
	case 200:
		return nil
	case 400:
		return nil
	}

	return err
}

func HandleCustomerDeletion(ns *models.Namespace, c req.Client) error {
	if !HasBillingInstance(ns) {
		return nil
	}

	status, err := c.DeleteCustomer(ns, BillingURL)
	if err != nil {
		return err
	}

	return HandleResponse(status, ErrDeletionReport)
}

func HandleReportUsage(ur *models.UsageRecord, c req.Client) error {
	if !HasBillingInstance(ur.Namespace) {
		return nil
	}

	status, err := c.ReportUsage(ur, BillingURL)
	if err != nil {
		return err
	}

	return HandleResponse(status, ErrReportUsage)
}
