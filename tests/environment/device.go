package environment

import (
	"testing"
	"time"
)

// AgeDevice moves the device uid's creation, last status change and the last time its tunnel was
// seen back by age, failing t unless exactly that device changed. It stands in for the minutes an
// ephemeral device's timeout takes to lapse.
func (dc *DockerCompose) AgeDevice(t *testing.T, uid string, age time.Duration) {
	t.Helper()

	dc.updateOne(t,
		"UPDATE devices SET created_at = created_at - :'age'::interval, status_updated_at = status_updated_at - :'age'::interval, last_seen = last_seen - :'age'::interval WHERE id = :'uid'",
		map[string]string{"uid": uid, "age": interval(age)})
}

// AgeDeviceRemoval moves the time the device uid was removed back by age, failing t unless exactly
// that device changed. Only a removed device has the stamp, so it fails for any other. It stands in
// for the days a removed device waits before the cleanup purges it.
func (dc *DockerCompose) AgeDeviceRemoval(t *testing.T, uid string, age time.Duration) {
	t.Helper()

	dc.updateOne(t,
		"UPDATE devices SET removed_at = removed_at - :'age'::interval WHERE id = :'uid' AND removed_at IS NOT NULL",
		map[string]string{"uid": uid, "age": interval(age)})
}
