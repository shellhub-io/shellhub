package storetest

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/scope"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *Suite) TestGetStats(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("succeeds without tenantID", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create data for tenant1
		tenant1 := s.CreateNamespace(t)
		s.CreateDevice(t, WithTenantID(tenant1), WithDeviceStatus("accepted"))
		s.CreateDevice(t, WithTenantID(tenant1), WithDeviceStatus("pending"))

		// Create a device with active session
		deviceUID := s.CreateDevice(t, WithTenantID(tenant1), WithDeviceStatus("accepted"))
		s.CreateSession(t, WithSessionDevice(deviceUID), WithSessionActive(true))

		// Create data for tenant2
		tenant2 := s.CreateNamespace(t)
		s.CreateDevice(t, WithTenantID(tenant2), WithDeviceStatus("accepted"))

		// Get global stats (no tenantID filter)
		stats, err := st.GetStats(ctx, scope.NewUnbounded(reasonTestQueryMechanics))
		require.NoError(t, err)
		require.NotNil(t, stats)

		// Should count all accepted devices and sessions across all tenants
		// RegisteredDevices counts only "accepted" devices, not "pending" or "rejected"
		assert.Equal(t, 3, stats.RegisteredDevices) // 2 accepted from tenant1 + 1 from tenant2
		assert.Equal(t, 1, stats.ActiveSessions)
		assert.Equal(t, 1, stats.PendingDevices)
	})

	t.Run("succeeds with specific tenantID", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create data for tenant1
		tenant1 := s.CreateNamespace(t)
		s.CreateDevice(t, WithTenantID(tenant1), WithDeviceStatus("accepted"))
		s.CreateDevice(t, WithTenantID(tenant1), WithDeviceStatus("accepted"))
		s.CreateDevice(t, WithTenantID(tenant1), WithDeviceStatus("pending"))

		// Create a device with active session
		deviceUID := s.CreateDevice(t, WithTenantID(tenant1), WithDeviceStatus("accepted"))
		s.CreateSession(t, WithSessionDevice(deviceUID), WithSessionActive(true))

		// Create data for tenant2 (should not be counted)
		tenant2 := s.CreateNamespace(t)
		s.CreateDevice(t, WithTenantID(tenant2), WithDeviceStatus("accepted"))

		// Get stats for tenant1 only
		stats, err := st.GetStats(ctx, scope.MustBounded(tenant1))
		require.NoError(t, err)
		require.NotNil(t, stats)

		// Should count only tenant1 accepted devices
		assert.Equal(t, 3, stats.RegisteredDevices) // 3 accepted devices from tenant1
		assert.Equal(t, 1, stats.ActiveSessions)
		assert.Equal(t, 1, stats.PendingDevices)
	})

	t.Run("succeeds with non-existent tenantID", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create data for an existing tenant
		tenant1 := s.CreateNamespace(t)
		s.CreateDevice(t, WithTenantID(tenant1), WithDeviceStatus("accepted"))

		deviceUID := s.CreateDevice(t, WithTenantID(tenant1), WithDeviceStatus("accepted"))
		s.CreateSession(t, WithSessionDevice(deviceUID), WithSessionActive(true))

		// Query with non-existent tenant ID
		stats, err := st.GetStats(ctx, scope.MustBounded("99999999-9999-4999-9999-999999999999"))
		require.NoError(t, err)
		require.NotNil(t, stats)

		// Should return zero stats
		assert.Equal(t, 0, stats.RegisteredDevices)
		assert.Equal(t, 0, stats.ActiveSessions)
		assert.Equal(t, 0, stats.PendingDevices)
		assert.Equal(t, 0, stats.RejectedDevices)
	})
}

// TestGetStatsOnlineBoundary pins the clock to assert the two-minute window that decides
// whether a device counts as online. The other stats tests create devices at the current
// time, so they pass whether the query reads the package clock or the wall clock; this one
// does not.
func (s *Suite) TestGetStatsOnlineBoundary(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	now := time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC)

	t.Run("counts only devices seen within the window", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))
		pinClock(t, now)

		tenantID := s.CreateNamespace(t)
		s.CreateDevice(t, WithTenantID(tenantID), WithDeviceStatus("accepted"), WithDeviceLastSeen(now.Add(-1*time.Minute)))
		s.CreateDevice(t, WithTenantID(tenantID), WithDeviceStatus("accepted"), WithDeviceLastSeen(now.Add(-3*time.Minute)))

		stats, err := st.GetStats(ctx, scope.MustBounded(tenantID))
		require.NoError(t, err)
		require.NotNil(t, stats)

		assert.Equal(t, 1, stats.OnlineDevices)
		assert.Equal(t, 2, stats.RegisteredDevices, "both devices are registered regardless of presence")
	})

	t.Run("moving the clock past the window takes a device offline", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))
		clk := pinClock(t, now)

		tenantID := s.CreateNamespace(t)
		s.CreateDevice(t, WithTenantID(tenantID), WithDeviceStatus("accepted"), WithDeviceLastSeen(now.Add(-1*time.Minute)))

		stats, err := st.GetStats(ctx, scope.MustBounded(tenantID))
		require.NoError(t, err)
		assert.Equal(t, 1, stats.OnlineDevices)

		clk.now = now.Add(2 * time.Minute)

		stats, err = st.GetStats(ctx, scope.MustBounded(tenantID))
		require.NoError(t, err)
		assert.Equal(t, 0, stats.OnlineDevices, "the window follows the clock, not wall time")
	})
}
