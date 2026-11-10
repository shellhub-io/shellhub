package main

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRegistrationSearch(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	t.Run("tenant registrations", func(t *testing.T) {
		testRegistrationSearchOn(t, compose, "/api/devices/tenant-registrations", "tenant", 0x90,
			func(hostname, mac string) requests.DeviceAuth { return newDeviceAuthRequest(t, hostname, mac) })
	})

	t.Run("provisioning key history", func(t *testing.T) {
		key := compose.CreateProvisioningKey(t, &requests.CreateProvisioningKey{
			Name: "searched",
			Mode: string(models.ProvisioningKeyModeManual),
		})

		testRegistrationSearchOn(t, compose, "/api/namespaces/provisioning-key/"+key.ID+"/history", "keyed", 0x9a,
			func(hostname, mac string) requests.DeviceAuth {
				return newKeyedDeviceAuthRequest(t, key.Key, hostname, mac)
			})
	})
}

func testRegistrationSearchOn(t *testing.T, compose *environment.DockerCompose, path, prefix string, macBlock byte, newRequest func(hostname, mac string) requests.DeviceAuth) {
	t.Helper()

	mac := func(n byte) string { return fmt.Sprintf("02:00:00:00:%02x:%02x", macBlock, n) }
	hostname := func(name string) string { return "searched-" + prefix + "-" + name }

	t.Run("the history is narrowed to the hostnames containing the search, in any case", func(t *testing.T) {
		pi := enroll(t, compose, newRequest(hostname("kitchen-pi"), mac(1)))
		enroll(t, compose, newRequest(hostname("garage"), mac(2)))
		tv := enroll(t, compose, newRequest(hostname("kitchen-tv"), mac(3)))

		events, total := searchRegistrations(t, compose, path, "SEARCHED-"+prefix+"-KITCHEN")

		assert.ElementsMatch(t, []string{pi.UID, tv.UID}, eventUIDs(events))
		assert.Equal(t, "2", total)
	})

	t.Run("a registration matched by an old hostname is not current once the device registered under another", func(t *testing.T) {
		req := newRequest(hostname("before"), mac(4))
		device := enroll(t, compose, req)
		compose.DeleteDevice(t, device.UID)

		req.Hostname = hostname("after")
		require.Equal(t, device.UID, enroll(t, compose, req).UID)

		before, _ := searchRegistrations(t, compose, path, hostname("before"))
		require.Len(t, before, 1)
		assert.False(t, before[0].IsCurrent, "the older registration was reported as the device's current one")

		after, _ := searchRegistrations(t, compose, path, hostname("after"))
		require.Len(t, after, 1)
		assert.True(t, after[0].IsCurrent)
	})

	t.Run("a filter on anything but the hostname containing a text, or a malformed one, is refused", func(t *testing.T) {
		for _, param := range []string{
			"filters[identity][contains]",
			"filters[hostname][eq]",
			"filters[hostname]",
		} {
			resp, err := compose.R(t.Context()).SetQueryParam(param, "searched").Get(path)
			require.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode(), "%s: %s", param, resp.String())
		}
	})
}

func searchRegistrations(t *testing.T, compose *environment.DockerCompose, path, hostname string) ([]models.ProvisioningKeyEvent, string) {
	t.Helper()

	events := []models.ProvisioningKeyEvent{}
	resp, err := compose.R(t.Context()).
		SetQueryParam("filters[hostname][contains]", hostname).
		SetResult(&events).
		Get(path)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	return events, resp.Header().Get("X-Total-Count")
}

func eventUIDs(events []models.ProvisioningKeyEvent) []string {
	uids := make([]string, len(events))
	for i, event := range events {
		uids[i] = event.DeviceUID
	}

	return uids
}
