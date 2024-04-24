package workers

import (
	"bufio"
	"bytes"
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
)

// heartbeat worker manages heartbeat tasks, signaling the online status of devices.
// It aggregates heartbeat data and updates the online status of devices accordingly.
// The maximum number of devices to wait for before triggering is defined by the `SHELLHUB_ASYNQ_GROUP_MAX_SIZE` (default is 500).
// Another triggering mechanism involves a timeout defined in the `SHELLHUB_ASYNQ_GROUP_MAX_DELAY` environment variable.
func (w *Workers) registerHeartbeat() {
	w.mux.HandleFunc(TaskHeartbeat, func(ctx context.Context, task *asynq.Task) error {
		log.
			WithFields(log.Fields{
				"component": "worker",
				"task":      TaskHeartbeat,
			}).
			Trace("Executing heartbeat worker.")

		scanner := bufio.NewScanner(bytes.NewReader(task.Payload()))
		scanner.Split(bufio.ScanLines)

		devices := make([]models.ConnectedDevice, 0)
		for scanner.Scan() {
			parts := strings.Split(scanner.Text(), "=")
			if len(parts) != 2 {
				log.WithFields(
					log.Fields{
						"component": "worker",
						"task":      TaskHeartbeat,
					}).
					Warn("failed to parse queue payload due to lack of '='.")

				continue
			}

			lastSeen, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				log.WithFields(
					log.Fields{
						"component": "worker",
						"task":      TaskHeartbeat,
					}).
					WithError(err).
					Warn("failed to parse timestamp to integer.")

				continue
			}

			parts = strings.Split(parts[0], ":")
			if len(parts) != 2 {
				log.WithFields(
					log.Fields{
						"component": "worker",
						"task":      TaskHeartbeat,
					}).
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

		if err := w.store.DeviceSetOnline(ctx, devices); err != nil {
			log.
				WithError(err).
				WithFields(log.Fields{
					"component": "worker",
					"task":      TaskHeartbeat,
				}).
				Error("failed to set devices as online")

			return err
		}

		return nil
	})
}
