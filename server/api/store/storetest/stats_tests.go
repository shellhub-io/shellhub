package storetest

import (
	"context"
	"testing"

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
		stats, err := st.GetStats(ctx, "")
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
		stats, err := st.GetStats(ctx, tenant1)
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
		stats, err := st.GetStats(ctx, "99999999-9999-4999-9999-999999999999")
		require.NoError(t, err)
		require.NotNil(t, stats)

		// Should return zero stats
		assert.Equal(t, 0, stats.RegisteredDevices)
		assert.Equal(t, 0, stats.ActiveSessions)
		assert.Equal(t, 0, stats.PendingDevices)
		assert.Equal(t, 0, stats.RejectedDevices)
	})
}
