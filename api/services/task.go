package services

import (
	"bufio"
	"bytes"
	"context"
	"maps"
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
	TaskDevicesHeartbeat          = worker.TaskPattern("api:heartbeat")
	CronDeviceCleanup             = worker.CronSpec("0 2 * * *")
	CronNamespaceDeviceCountSync  = worker.CronSpec("0 3 * * *")
	CronEphemeralCleanup          = worker.CronSpec("*/5 * * * *")
	CronEnrollmentCallbackCleanup = worker.CronSpec("0 4 * * *")
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

		mCount, err := s.store.DeviceHeartbeat(ctx, uids, clock.Now())
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

// EphemeralCleanup removes devices enrolled with an ephemeral install key that have stayed offline
// past their own per-device timeout. It runs on its own, more frequent schedule than the daily
// removed-device cleanup.
func (s *service) EphemeralCleanup() worker.CronHandler {
	return func(ctx context.Context) error {
		return s.store.WithTransaction(ctx, s.ephemeralCleanup())
	}
}

// EnrollmentCallbackCleanup prunes single-use callback redemption records once older than the maximum
// token TTL, past which the token has expired and can no longer gate a replay. The table only gains a
// row per resolved deferred webhook, so this keeps its growth bounded.
func (s *service) EnrollmentCallbackCleanup() worker.CronHandler {
	return func(ctx context.Context) error {
		cutoff := clock.Now().Add(-time.Duration(models.InstallKeyWebhookMaxCallbackTTL) * time.Second)

		deleted, err := s.store.EnrollmentCallbackCleanup(ctx, cutoff)
		if err != nil {
			return err
		}

		if deleted > 0 {
			log.WithField("deleted", deleted).Info("pruned expired enrollment callback redemptions")
		}

		return nil
	}
}

func (s *service) deviceCleanup() store.TransactionCb {
	return func(ctx context.Context) error {
		log.Info("Starting device cleanup for removed devices")

		filter := &query.Filters{
			Data: []query.Filter{
				{
					Type:   query.FilterTypeProperty,
					Params: &query.FilterProperty{Name: "removed_at", Operator: "lt", Value: clock.Now().AddDate(0, 0, -30)},
				},
			},
		}

		sorter := &query.Sorter{
			By:       "removed_at",
			Order:    query.OrderAsc,
			Tiebreak: "id",
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
				s.store.Options().Paginate(&query.Paginator{Page: page + 1, PerPage: pageSize}),
			}

			devices, _, err := s.store.DeviceList(ctx, store.DeviceAcceptableAsFalse, opts...)
			if err != nil {
				log.WithFields(log.Fields{"page": page, "error": err}).Error("Failed to list removed devices for page")

				return err
			}

			if len(devices) == 0 {
				continue
			}

			log.WithFields(log.Fields{"page": page + 1, "total_pages": totalPages, "devices": len(devices)}).
				Info("Processing page of removed devices")

			uids := make([]string, len(devices))
			for i, device := range devices {
				uids[i] = device.UID
				totalDeleted++
				deletedPerTenant[device.TenantID]++
			}

			if _, err := s.store.DeviceDeleteMany(ctx, uids); err != nil {
				log.WithField("page", page+1).
					WithError(err).
					Info("Failed to delete devices")

				return err
			}

			log.WithFields(log.Fields{"page": page + 1, "total_pages": totalPages, "devices": len(devices), "total_deleted": totalDeleted}).
				Info("Processing page of removed devices")

			if page < totalPages-1 {
				time.Sleep(100 * time.Millisecond)
			}
		}

		// Iterate tenants in a stable order so the decrement sequence is
		// deterministic (map iteration order is randomized in Go).
		for _, tenantID := range slices.Sorted(maps.Keys(deletedPerTenant)) {
			deletedCount := deletedPerTenant[tenantID]
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

func (s *service) ephemeralCleanup() store.TransactionCb {
	return func(ctx context.Context) error {
		log.Info("Starting cleanup for offline ephemeral devices")

		// The store selects ephemeral devices offline longer than their own per-device timeout. It is
		// capped per run, so a large scale-down drains across successive cron ticks.
		devices, err := s.store.DeviceListExpiredEphemeral(ctx)
		if err != nil {
			log.WithError(err).Error("Failed to list offline ephemeral devices")

			return err
		}

		if len(devices) == 0 {
			log.Info("No offline ephemeral devices found, cleanup completed")

			return nil
		}

		log.WithField("total_devices", len(devices)).Info("Found offline ephemeral devices to cleanup")

		// Decrement per (tenant, status): ephemeral devices are usually accepted, but a pending one
		// (whose accept failed) must not be counted against the accepted total.
		uids := make([]string, len(devices))
		deletedPerTenant := make(map[string]map[models.DeviceStatus]int64)
		for i, device := range devices {
			uids[i] = device.UID
			if deletedPerTenant[device.TenantID] == nil {
				deletedPerTenant[device.TenantID] = make(map[models.DeviceStatus]int64)
			}
			deletedPerTenant[device.TenantID][device.Status]++
		}

		if _, err := s.store.DeviceDeleteMany(ctx, uids); err != nil {
			log.WithError(err).Error("Failed to delete offline ephemeral devices")

			return err
		}

		// Iterate tenants and statuses in a stable order so the decrement sequence is deterministic.
		for _, tenantID := range slices.Sorted(maps.Keys(deletedPerTenant)) {
			for _, status := range slices.Sorted(maps.Keys(deletedPerTenant[tenantID])) {
				count := deletedPerTenant[tenantID][status]
				if err := s.store.NamespaceIncrementDeviceCount(ctx, tenantID, status, -count); err != nil {
					log.WithFields(log.Fields{"tenant_id": tenantID, "status": status, "deleted_count": count, "error": err}).
						Error("Failed to decrement ephemeral device count for namespace")

					return err
				}
			}
		}

		log.WithField("total_deleted", len(devices)).Info("Ephemeral device cleanup completed successfully")

		return nil
	}
}

func (s *service) NamespaceDeviceCountSync() worker.CronHandler {
	return func(ctx context.Context) error {
		log.Info("Starting namespace device count sync")

		if err := s.store.NamespaceSyncDeviceCounts(ctx); err != nil {
			log.WithError(err).Error("Failed to sync namespace device counts")

			return err
		}

		log.Info("Namespace device count sync completed")

		return nil
	}
}
