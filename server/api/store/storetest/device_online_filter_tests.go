package storetest

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDeviceListFiltersByOnline shows the online filter splits the device list at
// [models.DeviceOnlineWindow]: a device is online only when seen after the window opened and not
// disconnected since, every other device is offline, and a value that is not a boolean is an error.
func (s *Suite) TestDeviceListFiltersByOnline(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	require.NoError(t, s.provider.CleanDatabase(t))

	now := s.pinClock(t, time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC)).Now()
	windowOpened := now.Add(-models.DeviceOnlineWindow)

	tenantID := s.CreateNamespace(t)
	seenInWindow := s.CreateDevice(t, WithTenantID(tenantID), WithDeviceLastSeen(windowOpened.Add(time.Second)))
	seenAsWindowOpened := s.CreateDevice(t, WithTenantID(tenantID), WithDeviceLastSeen(windowOpened))
	seenBeforeWindow := s.CreateDevice(t, WithTenantID(tenantID), WithDeviceLastSeen(windowOpened.Add(-time.Second)))
	disconnected := s.CreateDevice(t, WithTenantID(tenantID), WithDeviceLastSeen(now), WithDeviceDisconnectedAt(now))

	online := []string{string(seenInWindow)}
	offline := []string{string(seenAsWindowOpened), string(seenBeforeWindow), string(disconnected)}

	list := func(value any) ([]models.Device, error) {
		filters := &query.Filters{Data: []query.Filter{{
			Type:   query.FilterTypeProperty,
			Params: &query.FilterProperty{Name: "online", Operator: "bool", Value: value},
		}}}

		devices, _, err := st.DeviceList(ctx, scope.MustBounded(tenantID), store.DeviceAcceptableIfNotAccepted, st.Options().Match(filters))

		return devices, err
	}

	cases := []struct {
		description string
		value       any
		want        []string
	}{
		{description: "bool true lists online devices", value: true, want: online},
		{description: "bool false lists offline devices", value: false, want: offline},
		{description: "string true lists online devices", value: "true", want: online},
		{description: "string false lists offline devices", value: "false", want: offline},
		{description: "JSON number one lists online devices", value: float64(1), want: online},
		{description: "JSON number zero lists offline devices", value: float64(0), want: offline},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			devices, err := list(tc.value)
			require.NoError(t, err)

			uids := make([]string, len(devices))
			for i, device := range devices {
				uids[i] = device.UID
			}

			assert.ElementsMatch(t, tc.want, uids)
		})
	}

	for _, value := range []any{42, "maybe"} {
		_, err := list(value)
		assert.Error(t, err, "value %v", value)
	}
}
