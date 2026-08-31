package services

import (
	"context"
	"maps"
	"slices"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/worker"
	"github.com/shellhub-io/shellhub/server/api/store"
	log "github.com/sirupsen/logrus"
)

const (
	CronDeviceCleanup             = worker.CronSpec("0 2 * * *")
	CronNamespaceDeviceCountSync  = worker.CronSpec("0 3 * * *")
	CronEphemeralCleanup          = worker.CronSpec("*/5 * * * *")
	CronEnrollmentCallbackCleanup = worker.CronSpec("0 4 * * *")
	CronSSHApprovalCleanup        = worker.CronSpec("*/10 * * * *")
	CronSessionCleanup            = worker.CronSpec("0 1 * * *")
)

const (
	sessionCleanupBatchSize = 1000

	sessionCleanupMaxBatches = 100

	sessionCleanupBatchPause = 200 * time.Millisecond
)

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

// SSHApprovalCleanup prunes SSH login approvals once expired. An expired row is
// already unreadable and undecidable, so this only keeps the table from growing;
// it runs often because the rows are short-lived and one is written per login
// that actually needs a browser step.
func (s *service) SSHApprovalCleanup() worker.CronHandler {
	return func(ctx context.Context) error {
		deleted, err := s.store.SSHApprovalCleanup(ctx, clock.Now())
		if err != nil {
			return err
		}

		if deleted > 0 {
			log.WithField("deleted", deleted).Info("pruned expired ssh approvals")
		}

		return nil
	}
}

// SessionCleanup enforces the instance's session retention window: sessions that started longer
// ago than retention are deleted, taking their events and recordings with them.
//
// A retention that is not positive means "keep forever" and prunes nothing. The guard matters
// more than it looks: read as a window, a zero would put the cutoff at now and delete every
// session on the instance.
func (s *service) SessionCleanup(retention time.Duration) worker.CronHandler {
	return func(ctx context.Context) error {
		return s.sessionCleanup(ctx, retention, sessionCleanupBatchPause)
	}
}

func (s *service) sessionCleanup(ctx context.Context, retention, pause time.Duration) error {
	if retention <= 0 {
		return nil
	}

	cutoff := clock.Now().Add(-retention)

	total := int64(0)
	batches := 0

	for batches < sessionCleanupMaxBatches {
		sessions, err := s.store.SessionListExpired(ctx, cutoff, sessionCleanupBatchSize)
		if err != nil {
			log.WithError(err).WithField("deleted", total).Error("failed to list expired sessions")

			return err
		}

		if len(sessions) == 0 {
			break
		}

		deletable, err := s.pruneRecordings(ctx, sessions)
		if err != nil {
			log.WithError(err).WithField("deleted", total).Error("failed to prune recordings of expired sessions")

			return err
		}

		if len(deletable) == 0 {
			log.WithFields(log.Fields{"deleted": total, "blocked": len(sessions)}).
				Warn("no expired session in the batch could be deleted; ending the run")

			break
		}

		deleted, err := s.store.SessionDeleteMany(ctx, deletable)
		if err != nil {
			log.WithError(err).WithField("deleted", total).Error("failed to prune expired sessions")

			return err
		}

		total += deleted
		batches++

		if len(sessions) < sessionCleanupBatchSize {
			break
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pause):
		}
	}

	if total > 0 {
		log.WithFields(log.Fields{"deleted": total, "cutoff": cutoff, "capped": batches == sessionCleanupMaxBatches}).
			Info("pruned sessions past the retention window")
	}

	return nil
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

		sc := scope.NewUnbounded("device-cleanup cron sweeps every namespace, bucketing its deletions per namespace afterwards")

		_, totalCount, err := s.store.DeviceList(ctx, sc, store.DeviceAcceptableAsFalse, s.store.Options().Match(filter))
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

			devices, _, err := s.store.DeviceList(ctx, sc, store.DeviceAcceptableAsFalse, opts...)
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

		for _, tenantID := range slices.Sorted(maps.Keys(deletedPerTenant)) {
			deletedCount := deletedPerTenant[tenantID]
			if err := s.store.NamespaceIncrementDeviceCount(ctx, scope.MustBounded(tenantID), models.DeviceStatusRemoved, -deletedCount); err != nil {
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

		for _, tenantID := range slices.Sorted(maps.Keys(deletedPerTenant)) {
			for _, status := range slices.Sorted(maps.Keys(deletedPerTenant[tenantID])) {
				count := deletedPerTenant[tenantID][status]
				if err := s.store.NamespaceIncrementDeviceCount(ctx, scope.MustBounded(tenantID), status, -count); err != nil {
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
