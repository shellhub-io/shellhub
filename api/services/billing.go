package services

import (
	req "github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func shouldReport(ns *models.Namespace) bool {
	if ns == nil || ns.Billing == nil {
		return false
	}

	return envs.HasBilling()
}

func createReportUsage(client req.Client, ns *models.Namespace, inc bool, device *models.Device) error {
	if !shouldReport(ns) {
		return nil
	}

	record := &models.UsageRecord{
		Device:    device,
		Inc:       inc,
		Timestamp: clock.Now().Unix(),
		Namespace: ns,
	}

	status, err := client.ReportUsage(record)
	if err != nil {
		return err
	}

	return reportStatusToError(status)
}

func deleteReportUsage(client req.Client, ns *models.Namespace) error {
	if !shouldReport(ns) {
		return nil
	}

	status, err := client.ReportDelete(ns)
	if err != nil {
		return err
	}

	return reportStatusToError(status)
}

func reportStatusToError(status int) error {
	if status == 200 || status == 402 || status == 400 {
		return nil
	}

	return ErrReport
}
