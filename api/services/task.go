package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/worker"
)

const (
	TaskDevicesHeartbeat = worker.TaskPattern("api:heartbeat")
)

// DevicesHeartbeat creates a task handler for processing device heartbeat signals. The payload format is a
// newline-separated list of device UIDs.
func (s *service) DevicesHeartbeat() worker.TaskHandler {
	return func(ctx context.Context, payload []byte) error {
		return nil
	}
}
