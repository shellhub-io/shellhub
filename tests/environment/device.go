package environment

import (
	"testing"
	"time"
)

// AgeDeviceDisconnection moves the time the device uid disconnected to age ago, failing t unless
// exactly that device changed. Only a device that has disconnected has the stamp, so it fails for a
// device still online. It stands in for the minutes an ephemeral device's timeout takes to lapse.
func (dc *DockerCompose) AgeDeviceDisconnection(t *testing.T, uid string, age time.Duration) {
	t.Helper()

	dc.updateOne(t,
		"UPDATE devices SET disconnected_at = disconnected_at - :'age'::interval WHERE id = :'uid' AND disconnected_at IS NOT NULL",
		map[string]string{"uid": uid, "age": interval(age)})
}

// AgeDeviceRemoval moves the time the device uid was removed to age ago, failing t unless exactly
// that device changed. Only a removed device has the stamp, so it fails for any other. It stands in
// for the days a removed device waits before the cleanup purges it.
func (dc *DockerCompose) AgeDeviceRemoval(t *testing.T, uid string, age time.Duration) {
	t.Helper()

	dc.updateOne(t,
		"UPDATE devices SET removed_at = removed_at - :'age'::interval WHERE id = :'uid' AND removed_at IS NOT NULL",
		map[string]string{"uid": uid, "age": interval(age)})
}
