package services

import (
	"bufio"
	"bytes"
	"context"
	"slices"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/worker"
	log "github.com/sirupsen/logrus"
)

const (
	TaskDevicesHeartbeat = worker.TaskPattern("api:heartbeat")
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
