package storetest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestTunnelUpdateDeviceUID tests updating device UID in tunnels
func (s *Suite) TestTunnelUpdateDeviceUID(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("succeeds when no tunnels exist", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create namespace and devices
		tenantID := s.CreateNamespace(t)
		oldDeviceUID := s.CreateDevice(t, WithDeviceName("old-device"), WithTenantID(tenantID))
		newDeviceUID := s.CreateDevice(t, WithDeviceName("new-device"), WithTenantID(tenantID))

		// Update device UID in tunnels (none exist, should succeed without error)
		err := st.TunnelUpdateDeviceUID(ctx, tenantID, string(oldDeviceUID), string(newDeviceUID))
		require.NoError(t, err)
	})

	t.Run("succeeds with valid tenant and device UIDs", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create namespace and devices
		tenantID := s.CreateNamespace(t)
		oldDeviceUID := s.CreateDevice(t, WithDeviceName("device-old"), WithTenantID(tenantID))
		newDeviceUID := s.CreateDevice(t, WithDeviceName("device-new"), WithTenantID(tenantID))

		// Note: We cannot easily create tunnels via the store interface
		// as there's no TunnelCreate method. This test verifies the method
		// executes without error even when no tunnels are present.
		// In a real scenario with tunnels, this would update their device UIDs.

		err := st.TunnelUpdateDeviceUID(ctx, tenantID, string(oldDeviceUID), string(newDeviceUID))
		require.NoError(t, err)
	})

	t.Run("succeeds with different tenant IDs", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create two different namespaces
		tenant1 := s.CreateNamespace(t, WithNamespaceName("tenant1"))
		tenant2 := s.CreateNamespace(t, WithNamespaceName("tenant2"))

		// Create devices in each namespace
		device1Old := s.CreateDevice(t, WithDeviceName("device1-old"), WithTenantID(tenant1))
		device1New := s.CreateDevice(t, WithDeviceName("device1-new"), WithTenantID(tenant1))

		// Update only affects tenant1's tunnels
		err := st.TunnelUpdateDeviceUID(ctx, tenant1, string(device1Old), string(device1New))
		require.NoError(t, err)

		// Update for tenant2 (even though no tunnels exist)
		device2Old := s.CreateDevice(t, WithDeviceName("device2-old"), WithTenantID(tenant2))
		device2New := s.CreateDevice(t, WithDeviceName("device2-new"), WithTenantID(tenant2))

		err = st.TunnelUpdateDeviceUID(ctx, tenant2, string(device2Old), string(device2New))
		require.NoError(t, err)
	})

	t.Run("succeeds when old device UID does not match any tunnels", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		nonExistentOldUID := "0000000000000000000000000000000000000000000000000000000000000000"
		newDeviceUID := s.CreateDevice(t, WithDeviceName("new-device"), WithTenantID(tenantID))

		// Should succeed even though oldUID doesn't exist
		err := st.TunnelUpdateDeviceUID(ctx, tenantID, nonExistentOldUID, string(newDeviceUID))
		require.NoError(t, err)
	})

	t.Run("succeeds with same old and new UIDs", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		deviceUID := s.CreateDevice(t, WithDeviceName("same-device"), WithTenantID(tenantID))

		// Update with same UID (idempotent operation)
		err := st.TunnelUpdateDeviceUID(ctx, tenantID, string(deviceUID), string(deviceUID))
		require.NoError(t, err)
	})
}
