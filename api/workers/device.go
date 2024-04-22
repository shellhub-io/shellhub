package workers

import (
	"context"
	"strings"

	"github.com/hibiken/asynq"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
)

type DeviceTask struct{}

func (*DeviceTask) parseConnectedDevicesQueue(payload []byte) (string, string, error) {
	builder := &strings.Builder{}
	if _, err := builder.Write(payload); err != nil {
		return "", "", err
	}

	parts := strings.Split(builder.String(), ":")
	if len(parts) != 2 {
		return "", "", nil
	}

	tenant := parts[0]
	target := parts[1]

	return tenant, target, nil
}

func (d *DeviceTask) IncreaseConnectedDevices(w *Workers) func(ctx context.Context, task *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		tenant, target, err := d.parseConnectedDevicesQueue(task.Payload())
		if err != nil {
			log.WithError(err).
				WithFields(log.Fields{"task": TaskConnectedDevicesIncrease}).
				Error("failed to parse task payload")

			return err
		}

		// TODO: validator.IsUUID(target)
		if len(target) > 15 {
			// NOTE: d will always be not nil until a unexpected error happens.
			device, err := w.store.DeviceGetByUID(ctx, models.UID(target), tenant)
			if err != nil {
				log.
					WithError(err).
					WithFields(log.Fields{
						"component": "worker",
						"task":      TaskConnectedDevicesIncrease,
						"tenant":    tenant,
					}).
					Error("failed to retrieve uid target")

				return nil
			}

			target = string(device.Status)
		}

		if err := w.cache.IncreaseConnectedDevices(ctx, tenant, models.DeviceStatus(target), 1); err != nil {
			log.
				WithError(err).
				WithFields(log.Fields{
					"component": "worker",
					"task":      TaskConnectedDevicesIncrease,
					"tenant":    tenant,
				}).
				Error("failed to adjust connected devices count")

			return err
		}

		return nil
	}
}

func (d *DeviceTask) DecreaseConnectedDevices(w *Workers) func(ctx context.Context, task *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		tenant, target, err := d.parseConnectedDevicesQueue(task.Payload())
		if err != nil {
			log.WithError(err).
				WithFields(log.Fields{"task": TaskConnectedDevicesDecrease}).
				Error("failed to parse task payload")

			return err
		}

		// TODO: validator.IsUUID(target)
		if len(target) > 15 {
			// NOTE: d will always be not nil until a unexpected error happens.
			device, err := w.store.DeviceGetByUID(ctx, models.UID(target), tenant)
			if err != nil {
				log.
					WithError(err).
					WithFields(log.Fields{
						"component": "worker",
						"task":      TaskConnectedDevicesDecrease,
						"tenant":    tenant,
					}).
					Error("failed to retrieve uid target")

				return nil
			}

			target = string(device.Status)
		}

		if err := w.cache.DecreaseConnectedDevices(ctx, tenant, models.DeviceStatus(target), 1); err != nil {
			log.
				WithError(err).
				WithFields(log.Fields{
					"component": "worker",
					"task":      TaskConnectedDevicesDecrease,
					"tenant":    tenant,
				}).
				Error("failed to adjust connected devices count")

			return err
		}

		return nil
	}
}
