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
		log.WithFields(
			log.Fields{
				"component": "worker",
				"task":      TaskHeartbeat,
			}).
			Trace("Executing heartbeat worker.")

		scanner := bufio.NewScanner(bytes.NewReader(task.Payload()))
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			parts := strings.SplitN(scanner.Text(), ":", 2)
			uid := parts[0]

			i, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				log.WithFields(
					log.Fields{
						"component": "worker",
						"task":      TaskHeartbeat,
						"index":     rune(i),
					}).
					WithError(err).
					Warn("Failed to parse timestamp to integer.")

				continue
			}

			timestamp := time.Unix(i, 0)

			w.store.DeviceSetOnline(ctx, models.UID(uid), timestamp, true) //nolint:errcheck
		}

		return nil
	})
}
