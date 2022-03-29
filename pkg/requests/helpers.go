package requests

import (
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func HasBillingInstance(ns *models.Namespace) bool {
	if ns == nil || ns.Billing == nil {
		return false
	}

	return envs.HasBilling()
}
