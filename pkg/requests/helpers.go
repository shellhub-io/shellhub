package requests

import (
	"errors"

	"github.com/shellhub-io/shellhub/pkg/models"
)

var ErrReport = errors.New("report error")

func HasBillingInstance(ns *models.Namespace) bool {
	if ns == nil || ns.Billing == nil || !ns.Billing.Active || ns.MaxDevices != -1 {
		return false
	}

	return true
}

func HandleStatusResponse(status int) error {
	if status == 200 || status == 402 || status == 400 {
		return nil
	}

	return ErrReport
}
