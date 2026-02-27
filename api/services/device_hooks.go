package services

import (
	"context"
	"fmt"

	"github.com/shellhub-io/shellhub/pkg/models"
)

// DeviceMergeHookFn is called when two devices are merged. The hook receives
// the tenant ID and both the old and new device models. Hooks run inside the
// same transaction as mergeDevice, so a returned error will roll back the
// entire merge.
type DeviceMergeHookFn func(ctx context.Context, tenantID string, oldDevice, newDevice *models.Device) error

var deviceMergeHooks []DeviceMergeHookFn

// OnDeviceMerge registers a hook that fires when two devices are merged.
// It must be called during package init, before the server starts handling
// requests. Cloud packages use this to handle tunnel UID transfers, etc.
func OnDeviceMerge(fn DeviceMergeHookFn) {
	if fn == nil {
		panic("services: OnDeviceMerge called with nil hook")
	}

	deviceMergeHooks = append(deviceMergeHooks, fn)
}

// fireDeviceMerge dispatches all registered merge hooks sequentially.
// The first error aborts execution (the caller's transaction rolls back).
func fireDeviceMerge(ctx context.Context, tenantID string, oldDevice, newDevice *models.Device) error {
	for _, fn := range deviceMergeHooks {
		if err := fn(ctx, tenantID, oldDevice, newDevice); err != nil {
			return fmt.Errorf("device merge hook failed: %w", err)
		}
	}

	return nil
}
