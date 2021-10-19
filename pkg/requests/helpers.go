package requests

import (
	"errors"

	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
)

var ErrReport = errors.New("report error")

func HasBillingInstance(ns *models.Namespace) bool {
	if ns == nil || ns.Billing == nil {
		return false
	}

	return envs.HasBilling()
}

func HandleStatusResponse(status int) error {
	if status == 200 || status == 402 || status == 400 {
		return nil
	}

	return ErrReport
}
