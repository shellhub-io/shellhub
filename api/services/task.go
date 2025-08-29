package services

import (
	"bufio"
	"bytes"
	"context"
	"slices"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/worker"
	log "github.com/sirupsen/logrus"
)

const (
	TaskDevicesHeartbeat = worker.TaskPattern("api:heartbeat")
	CronDeviceCleanup    = worker.CronSpec("0 2 * * *")
)

// DevicesHeartbeat creates a task handler for processing device heartbeat signals. The payload format is a
// newline-separated list of device UIDs.
func (s *service) DevicesHeartbeat() worker.TaskHandler {
	return func(ctx context.Context, payload []byte) error {
		log.WithField("task", TaskDevicesHeartbeat.String()).
			Info("executing heartbeat task")

		scanner := bufio.NewScanner(bytes.NewReader(payload))
		scanner.Split(bufio.ScanLines)

		uids := make([]string, 0)
		for scanner.Scan() {
			uid := scanner.Text()
			if uid == "" {
				continue
			}

			uids = append(uids, uid)
		}

		slices.Sort(uids)
		uids = slices.Compact(uids)

		mCount, err := s.store.DeviceBulkUpdate(ctx, uids, &models.DeviceChanges{LastSeen: clock.Now(), DisconnectedAt: nil})
		if err != nil {
			log.WithField("task", TaskDevicesHeartbeat.String()).
				WithError(err).
				Error("failed to complete the heartbeat task")

			return err
		}

		log.WithField("task", TaskDevicesHeartbeat.String()).
			WithField("modified_count", mCount).
			Info("finishing heartbeat task")

		return nil
	}
}

func (s *service) DeviceCleanup() worker.CronHandler {
	return func(ctx context.Context) error {
		return s.store.WithTransaction(ctx, s.deviceCleanup())
	}
}

func (s *service) deviceCleanup() store.TransactionCb {
	return func(ctx context.Context) error {
		log.Info("Starting device cleanup for removed devices")

		filter := &query.Filters{
			Data: []query.Filter{
				{
					Type:   query.FilterTypeProperty,
					Params: &query.FilterProperty{Name: "status", Operator: "eq", Value: string(models.DeviceStatusRemoved)},
				},
				{
					Type:   query.FilterTypeProperty,
					Params: &query.FilterProperty{Name: "status_updated_at", Operator: "lt", Value: time.Now().AddDate(0, 0, -30)},
				},
			},
		}

		sorter := &query.Sorter{
			By:    "status_updated_at",
			Order: query.OrderAsc,
		}

		_, totalCount, err := s.store.DeviceList(ctx, store.DeviceAcceptableAsFalse, s.store.Options().Match(filter))
		if err != nil {
			log.WithError(err).Error("Failed to get total count of removed devices")

			return err
		}

		if totalCount == 0 {
			log.Info("No removed devices found, cleanup completed")

			return nil
		}

		log.WithField("total_devices", totalCount).Info("Found removed devices to cleanup")

		const pageSize = 1000
		totalDeleted := 0

		deletedPerTenant := make(map[string]int64)
		totalPages := (totalCount + pageSize - 1) / pageSize

		for page := range totalPages {
			opts := []store.QueryOption{
				s.store.Options().Match(filter),
				s.store.Options().Sort(sorter),
				s.store.Options().Paginate(&query.Paginator{Page: page, PerPage: pageSize}),
			}

			devices, _, err := s.store.DeviceList(ctx, store.DeviceAcceptableAsFalse, opts...)
			if err != nil {
				log.WithFields(log.Fields{"page": page, "error": err}).Error("Failed to list removed devices for page")

				return err
			}

			log.WithFields(log.Fields{"page": page + 1, "total_pages": totalPages, "devices": len(devices)}).
				Info("Processing page of removed devices")

			for _, device := range devices {
				if err := s.store.DeviceDelete(ctx, models.UID(device.UID)); err != nil {
					log.WithError(err).WithFields(log.Fields{"device_uid": device.UID}).Error("Failed to delete removed device")

					return err
				}

				log.WithFields(log.Fields{"device_uid": device.UID}).Debug("Successfully deleted removed device")
				totalDeleted++
				deletedPerTenant[device.TenantID]++
			}

			log.WithFields(log.Fields{"page": page + 1, "total_pages": totalPages, "devices": len(devices), "total_deleted": totalDeleted}).
				Info("Processing page of removed devices")

			if page < totalPages-1 {
				time.Sleep(100 * time.Millisecond)
			}
		}

		for tenantID, deletedCount := range deletedPerTenant {
			if err := s.store.NamespaceIncrementDeviceCount(ctx, tenantID, models.DeviceStatusRemoved, -deletedCount); err != nil {
				log.WithFields(log.Fields{"tenant_id": tenantID, "deleted_count": deletedCount, "error": err}).
					Error("Failed to decrement removed device count for namespace")

				return err
			}
		}

		log.WithFields(log.Fields{"total_found": totalCount, "total_deleted": totalDeleted}).
			Info("Device cleanup completed successfully")

		return nil
	}
}
