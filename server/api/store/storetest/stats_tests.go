package storetest

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetStats locks the counts the dashboard reads.
func (s *Suite) TestGetStats(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("succeeds without tenantID", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenant1 := s.CreateNamespace(t)
		s.CreateDevice(t, WithTenantID(tenant1), WithDeviceStatus("accepted"))
		s.CreateDevice(t, WithTenantID(tenant1), WithDeviceStatus("pending"))

		deviceUID := s.CreateDevice(t, WithTenantID(tenant1), WithDeviceStatus("accepted"))
		s.CreateSession(t, WithSessionDevice(deviceUID), WithSessionActive(true))

		tenant2 := s.CreateNamespace(t)
		s.CreateDevice(t, WithTenantID(tenant2), WithDeviceStatus("accepted"))

		stats, err := st.GetStats(ctx, scope.NewUnbounded(reasonTestQueryMechanics))
		require.NoError(t, err)
		require.NotNil(t, stats)

		assert.Equal(t, 3, stats.RegisteredDevices) // 2 accepted from tenant1 + 1 from tenant2
		assert.Equal(t, 1, stats.ActiveSessions)
		assert.Equal(t, 1, stats.PendingDevices)
	})

	t.Run("succeeds with specific tenantID", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenant1 := s.CreateNamespace(t)
		s.CreateDevice(t, WithTenantID(tenant1), WithDeviceStatus("accepted"))
		s.CreateDevice(t, WithTenantID(tenant1), WithDeviceStatus("accepted"))
		s.CreateDevice(t, WithTenantID(tenant1), WithDeviceStatus("pending"))

		deviceUID := s.CreateDevice(t, WithTenantID(tenant1), WithDeviceStatus("accepted"))
		s.CreateSession(t, WithSessionDevice(deviceUID), WithSessionActive(true))

		tenant2 := s.CreateNamespace(t)
		s.CreateDevice(t, WithTenantID(tenant2), WithDeviceStatus("accepted"))

		stats, err := st.GetStats(ctx, scope.MustBounded(tenant1))
		require.NoError(t, err)
		require.NotNil(t, stats)

		assert.Equal(t, 3, stats.RegisteredDevices) // 3 accepted devices from tenant1
		assert.Equal(t, 1, stats.ActiveSessions)
		assert.Equal(t, 1, stats.PendingDevices)
	})

	t.Run("succeeds with non-existent tenantID", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenant1 := s.CreateNamespace(t)
		s.CreateDevice(t, WithTenantID(tenant1), WithDeviceStatus("accepted"))

		deviceUID := s.CreateDevice(t, WithTenantID(tenant1), WithDeviceStatus("accepted"))
		s.CreateSession(t, WithSessionDevice(deviceUID), WithSessionActive(true))

		stats, err := st.GetStats(ctx, scope.MustBounded("99999999-9999-4999-9999-999999999999"))
		require.NoError(t, err)
		require.NotNil(t, stats)

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

// TestCountRegisteredDevices covers the narrow count the license and firewall evaluations use.
// It answers the same question as GetStats' RegisteredDevices field, so the two must agree on
// what "registered" means: accepted only, pending and rejected excluded.
func (s *Suite) TestCountRegisteredDevices(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("counts accepted devices across every namespace when unbounded", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenant1 := s.CreateNamespace(t)
		s.CreateDevice(t, WithTenantID(tenant1), WithDeviceStatus("accepted"))
		s.CreateDevice(t, WithTenantID(tenant1), WithDeviceStatus("accepted"))
		s.CreateDevice(t, WithTenantID(tenant1), WithDeviceStatus("pending"))
		s.CreateDevice(t, WithTenantID(tenant1), WithDeviceStatus("rejected"))

		tenant2 := s.CreateNamespace(t)
		s.CreateDevice(t, WithTenantID(tenant2), WithDeviceStatus("accepted"))

		count, err := st.CountRegisteredDevices(ctx, scope.NewUnbounded(reasonTestQueryMechanics))
		require.NoError(t, err)

		assert.Equal(t, 3, count, "2 accepted from tenant1 + 1 from tenant2, ignoring pending and rejected")
	})

	t.Run("counts only the scoped namespace when bounded", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenant1 := s.CreateNamespace(t)
		s.CreateDevice(t, WithTenantID(tenant1), WithDeviceStatus("accepted"))
		s.CreateDevice(t, WithTenantID(tenant1), WithDeviceStatus("accepted"))
		s.CreateDevice(t, WithTenantID(tenant1), WithDeviceStatus("pending"))

		tenant2 := s.CreateNamespace(t)
		s.CreateDevice(t, WithTenantID(tenant2), WithDeviceStatus("accepted"))

		count, err := st.CountRegisteredDevices(ctx, scope.MustBounded(tenant1))
		require.NoError(t, err)

		assert.Equal(t, 2, count)
	})

	t.Run("returns zero for a namespace with no accepted devices", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		s.CreateDevice(t, WithTenantID(tenantID), WithDeviceStatus("pending"))

		count, err := st.CountRegisteredDevices(ctx, scope.MustBounded(tenantID))
		require.NoError(t, err)

		assert.Equal(t, 0, count)
	})

	t.Run("agrees with GetStats", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		s.CreateDevice(t, WithTenantID(tenantID), WithDeviceStatus("accepted"))
		s.CreateDevice(t, WithTenantID(tenantID), WithDeviceStatus("accepted"))
		s.CreateDevice(t, WithTenantID(tenantID), WithDeviceStatus("rejected"))

		sc := scope.NewUnbounded(reasonTestQueryMechanics)

		stats, err := st.GetStats(ctx, sc)
		require.NoError(t, err)

		count, err := st.CountRegisteredDevices(ctx, sc)
		require.NoError(t, err)

		assert.Equal(t, stats.RegisteredDevices, count)
	})
}
