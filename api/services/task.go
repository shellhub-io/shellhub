package services

import (
	"bufio"
	"bytes"
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/worker"
	log "github.com/sirupsen/logrus"
)

const (
	TaskDevicesHeartbeat = worker.TaskPattern("api:heartbeat")
)

// Device Heartbeat sets the device status to "online". It processes in batch.
func (s *service) DevicesHeartbeat() worker.TaskHandler {
	return func(ctx context.Context, payload []byte) error {
		log.WithField("task", TaskDevicesHeartbeat.String()).
			Info("executing heartbeat task")

		scanner := bufio.NewScanner(bytes.NewReader(payload))
		scanner.Split(bufio.ScanLines)

		devices := make([]models.ConnectedDevice, 0)
		for scanner.Scan() {
			parts := strings.Split(scanner.Text(), "=")
			if len(parts) != 2 {
				log.WithField("task", TaskDevicesHeartbeat.String()).
					Warn("failed to parse queue payload due to lack of '='.")

				continue
			}

			lastSeen, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				log.WithField("task", TaskDevicesHeartbeat.String()).
					WithError(err).
					Warn("failed to parse timestamp to integer.")

				continue
			}

			parts = strings.Split(parts[0], ":")
			if len(parts) != 2 {
				log.WithField("task", TaskDevicesHeartbeat.String()).
					Warn("failed to parse queue payload due to lack of ':'.")

				continue
			}

			device := models.ConnectedDevice{
				UID:      parts[1],
				TenantID: parts[0],
				LastSeen: time.Unix(lastSeen, 0),
			}

			devices = append(devices, device)
		}

		if err := s.store.DeviceSetOnline(ctx, devices); err != nil {
			log.WithField("task", TaskDevicesHeartbeat.String()).
				WithError(err).
				Error("failed to complete the heartbeat task")

			return err
		}

		log.WithField("task", TaskDevicesHeartbeat.String()).
			Info("finishing heartbeat task")

		return nil
	}
}
