package reports

import (
	"github.com/shellhub-io/shellhub/pkg/api/client"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func HandleReports(ns *models.Namespace, uid models.UID, inc bool) error {
	if ns.Billing == nil || !ns.Billing.Active {
		return nil
	}

	status, err := client.NewClient().ReportUsage(
		&models.UsageRecord{
			UUID:      string(uid),
			Inc:       inc,
			Timestamp: clock.Now().Unix(),
			Namespace: ns,
		},
		"billing-api",
	)
	if err != nil {
		return err
	}

	switch status {
	case 402:
		return nil
	case 200:
		return nil
	case 400:
		return nil
	}

	return ErrReportUsage
}
