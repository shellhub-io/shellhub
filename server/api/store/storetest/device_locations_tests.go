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

// TestDeviceLocations pins what the device map reads: accepted, located devices counted per
// rounded position, within the scope and the filter given.
func (s *Suite) TestDeviceLocations(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	offline := func() DeviceOption { return WithDeviceDisconnectedAt(s.Clock().Now()) }
	stale := func() DeviceOption { return WithDeviceLastSeen(s.Clock().Now().Add(-time.Hour)) }

	t.Run("counts the devices and the online devices at each rounded position", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		s.CreateDevice(t, WithTenantID(tenantID), WithDevicePosition(47.3769, 8.5417))
		s.CreateDevice(t, WithTenantID(tenantID), WithDevicePosition(47.3772, 8.5421))
		s.CreateDevice(t, WithTenantID(tenantID), WithDevicePosition(47.3768, 8.5419), offline())
		s.CreateDevice(t, WithTenantID(tenantID), WithDevicePosition(46.9480, 7.4474), stale())

		locations, err := st.DeviceLocations(ctx, scope.MustBounded(tenantID))
		require.NoError(t, err)

		assert.ElementsMatch(t, []models.DeviceLocation{
			{Latitude: 47.38, Longitude: 8.54, Devices: 3, Online: 2},
			{Latitude: 46.95, Longitude: 7.45, Devices: 1, Online: 0},
		}, locations)
	})

	t.Run("leaves out devices that are not accepted or have no position", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		s.CreateDevice(t, WithTenantID(tenantID), WithDevicePosition(47.38, 8.54))
		s.CreateDevice(t, WithTenantID(tenantID), WithDevicePosition(47.38, 8.54), WithDeviceStatus(models.DeviceStatusPending))
		s.CreateDevice(t, WithTenantID(tenantID), WithDevicePosition(47.38, 8.54), WithDeviceStatus(models.DeviceStatusRejected))
		s.CreateDevice(t, WithTenantID(tenantID), WithDevicePosition(0, 0))
		s.CreateDevice(t, WithTenantID(tenantID))

		locations, err := st.DeviceLocations(ctx, scope.MustBounded(tenantID))
		require.NoError(t, err)

		assert.Equal(t, []models.DeviceLocation{{Latitude: 47.38, Longitude: 8.54, Devices: 1, Online: 1}}, locations)
	})

	t.Run("counts only the scope's namespace", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		other := s.CreateNamespace(t)
		s.CreateDevice(t, WithTenantID(tenantID), WithDevicePosition(47.38, 8.54))
		s.CreateDevice(t, WithTenantID(other), WithDevicePosition(47.38, 8.54))
		s.CreateDevice(t, WithTenantID(other), WithDevicePosition(52.52, 13.40))

		locations, err := st.DeviceLocations(ctx, scope.MustBounded(tenantID))
		require.NoError(t, err)

		assert.Equal(t, []models.DeviceLocation{{Latitude: 47.38, Longitude: 8.54, Devices: 1, Online: 1}}, locations)
	})

	t.Run("counts only the devices the filter matches", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		s.CreateDevice(t, WithTenantID(tenantID), WithDevicePosition(47.38, 8.54))
		s.CreateDevice(t, WithTenantID(tenantID), WithDevicePosition(52.52, 13.40))

		filters := &query.Filters{Data: []query.Filter{{
			Type:   query.FilterTypeProperty,
			Params: &query.FilterProperty{Name: "latitude", Operator: "gt", Value: 50},
		}}}

		locations, err := st.DeviceLocations(ctx, scope.MustBounded(tenantID), st.Options().Match(filters))
		require.NoError(t, err)

		assert.Equal(t, []models.DeviceLocation{{Latitude: 52.52, Longitude: 13.40, Devices: 1, Online: 1}}, locations)
	})

	t.Run("returns no locations for a namespace with no located device", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		s.CreateDevice(t, WithTenantID(tenantID))

		locations, err := st.DeviceLocations(ctx, scope.MustBounded(tenantID))
		require.NoError(t, err)

		assert.Empty(t, locations)
	})
}

// TestDeviceListFiltersByLocated shows the located filter splits the device list into the devices
// with a position and those the store keeps at 0,0, a device on the equator counting as located.
func (s *Suite) TestDeviceListFiltersByLocated(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	require.NoError(t, s.provider.CleanDatabase(t))

	tenantID := s.CreateNamespace(t)
	located := s.CreateDevice(t, WithTenantID(tenantID), WithDevicePosition(47.38, 8.54))
	equator := s.CreateDevice(t, WithTenantID(tenantID), WithDevicePosition(0, 8.54))
	unlocated := s.CreateDevice(t, WithTenantID(tenantID))

	list := func(t *testing.T, value bool) []string {
		t.Helper()

		filters := &query.Filters{Data: []query.Filter{{
			Type:   query.FilterTypeProperty,
			Params: &query.FilterProperty{Name: "located", Operator: "bool", Value: value},
		}}}

		devices, _, err := st.DeviceList(ctx, scope.MustBounded(tenantID), store.DeviceAcceptableIfNotAccepted, st.Options().Match(filters))
		require.NoError(t, err)

		uids := make([]string, len(devices))
		for i, device := range devices {
			uids[i] = device.UID
		}

		return uids
	}

	assert.ElementsMatch(t, []string{string(located), string(equator)}, list(t, true))
	assert.ElementsMatch(t, []string{string(unlocated)}, list(t, false))
}
