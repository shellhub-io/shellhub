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

func fireDeviceMerge(ctx context.Context, tenantID string, oldDevice, newDevice *models.Device) error {
	for _, fn := range deviceMergeHooks {
		if err := fn(ctx, tenantID, oldDevice, newDevice); err != nil {
			return fmt.Errorf("device merge hook failed: %w", err)
		}
	}

	return nil
}

// DeviceRemovedHookFn is called after a device leaves its namespace, once the removal has
// committed. It cannot fail the removal, which has already happened. It must be safe to call more
// than once for the same device, and for a removal another process made.
type DeviceRemovedHookFn func(ctx context.Context, tenantID, uid string)

var deviceRemovedHooks []DeviceRemovedHookFn

// OnDeviceRemoved registers a hook that fires after a device is removed or deleted, on every path
// in this process that does it, and again when the device's tunnel next beats, which is how a
// removal made by another process or replica reaches this one. It must be called before the server
// starts handling requests. The server uses it to close the removed device's tunnel.
func OnDeviceRemoved(fn DeviceRemovedHookFn) {
	if fn == nil {
		panic("services: OnDeviceRemoved called with nil hook")
	}

	deviceRemovedHooks = append(deviceRemovedHooks, fn)
}

func fireDeviceRemoved(ctx context.Context, tenantID, uid string) {
	for _, fn := range deviceRemovedHooks {
		fn(ctx, tenantID, uid)
	}
}
