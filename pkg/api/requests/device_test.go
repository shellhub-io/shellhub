package requests

import (
	"testing"

	"github.com/gustavosbarreto/structsnapshot"
	"github.com/stretchr/testify/assert"
)

func TestDeviceSnapShot(t *testing.T) {
	deviceStructList := []interface{}{
		DeviceParam{},
		DeviceGet{},
		DeviceDelete{},
		DeviceRename{},
		DevicePendingStatus{},
		DeviceUpdateStatus{},
		DeviceLookup{},
		DeviceOffline{},
		DeviceHeartbeat{},
		DeviceCreateTag{},
		DeviceRemoveTag{},
		DeviceUpdateTag{},
		DeviceIdentity{},
		DeviceInfo{},
		DeviceGetPublicURL{},
		DeviceUpdate{},
		DevicePublicURLAddress{},
	}

	for _, d := range deviceStructList {
		snapShot, err := structsnapshot.TakeSnapshot(d)
		assert.NoError(t, err)

		loadedSnapshot, err := structsnapshot.LoadSnapshot(d)
		assert.NoError(t, err)

		assert.Equal(t, loadedSnapshot, snapShot)
	}
}
