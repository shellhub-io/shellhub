package storetest

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDeviceList tests the DeviceList method across all implementations
func (s *Suite) TestDeviceList(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("succeeds when no devices are found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		devices, count, err := st.DeviceList(ctx, store.DeviceAcceptableIfNotAccepted,
			st.Options().Match(&query.Filters{}),
			st.Options().Sort(&query.Sorter{By: "last_seen", Order: query.OrderAsc}),
			st.Options().Paginate(&query.Paginator{Page: -1, PerPage: -1}),
		)

		require.NoError(t, err)
		assert.Empty(t, devices)
		assert.Equal(t, 0, count)
	})

	t.Run("succeeds when devices are found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create test devices
		s.CreateDevice(t, WithDeviceName("device-1"))
		s.CreateDevice(t, WithDeviceName("device-2"))
		s.CreateDevice(t, WithDeviceName("device-3"))
		s.CreateDevice(t, WithDeviceName("device-4"))

		// List all devices
		devices, count, err := st.DeviceList(ctx, store.DeviceAcceptableIfNotAccepted,
			st.Options().Match(&query.Filters{}),
			st.Options().Sort(&query.Sorter{By: "last_seen", Order: query.OrderAsc}),
			st.Options().Paginate(&query.Paginator{Page: -1, PerPage: -1}),
		)

		require.NoError(t, err)
		assert.Equal(t, 4, count)
		assert.Len(t, devices, 4)
	})

	t.Run("succeeds when devices are found with pagination", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create test devices
		s.CreateDevice(t, WithDeviceName("device-1"))
		s.CreateDevice(t, WithDeviceName("device-2"))
		s.CreateDevice(t, WithDeviceName("device-3"))
		s.CreateDevice(t, WithDeviceName("device-4"))

		// Get page 2 with 2 items per page
		devices, count, err := st.DeviceList(ctx, store.DeviceAcceptableIfNotAccepted,
			st.Options().Match(&query.Filters{}),
			st.Options().Sort(&query.Sorter{By: "last_seen", Order: query.OrderAsc}),
			st.Options().Paginate(&query.Paginator{Page: 2, PerPage: 2}),
		)

		require.NoError(t, err)
		assert.Equal(t, 4, count, "total count should be 4")
		assert.Len(t, devices, 2, "page 2 should have 2 items")
	})

	t.Run("succeeds when devices are found with order asc", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create test devices with different last seen times
		s.CreateDevice(t, WithDeviceName("device-1"))
		time.Sleep(10 * time.Millisecond)
		s.CreateDevice(t, WithDeviceName("device-2"))
		time.Sleep(10 * time.Millisecond)
		s.CreateDevice(t, WithDeviceName("device-3"))

		devices, count, err := st.DeviceList(ctx, store.DeviceAcceptableIfNotAccepted,
			st.Options().Match(&query.Filters{}),
			st.Options().Sort(&query.Sorter{By: "last_seen", Order: query.OrderAsc}),
			st.Options().Paginate(&query.Paginator{Page: -1, PerPage: -1}),
		)

		require.NoError(t, err)
		assert.Equal(t, 3, count)
		assert.Len(t, devices, 3)
	})

	t.Run("succeeds when devices are found with order desc", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create test devices
		s.CreateDevice(t, WithDeviceName("device-1"))
		time.Sleep(10 * time.Millisecond)
		s.CreateDevice(t, WithDeviceName("device-2"))
		time.Sleep(10 * time.Millisecond)
		s.CreateDevice(t, WithDeviceName("device-3"))

		devices, count, err := st.DeviceList(ctx, store.DeviceAcceptableIfNotAccepted,
			st.Options().Match(&query.Filters{}),
			st.Options().Sort(&query.Sorter{By: "last_seen", Order: query.OrderDesc}),
			st.Options().Paginate(&query.Paginator{Page: -1, PerPage: -1}),
		)

		require.NoError(t, err)
		assert.Equal(t, 3, count)
		assert.Len(t, devices, 3)
	})

	t.Run("succeeds when filtering by status", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create devices with different statuses
		s.CreateDevice(t, WithDeviceName("device-accepted"), WithDeviceStatus(models.DeviceStatusAccepted))
		s.CreateDevice(t, WithDeviceName("device-pending"), WithDeviceStatus(models.DeviceStatusPending))
		s.CreateDevice(t, WithDeviceName("device-accepted-2"), WithDeviceStatus(models.DeviceStatusAccepted))

		// Filter by pending status
		devices, count, err := st.DeviceList(ctx, store.DeviceAcceptableIfNotAccepted,
			st.Options().WithDeviceStatus(models.DeviceStatusPending),
			st.Options().Match(&query.Filters{}),
			st.Options().Sort(&query.Sorter{By: "last_seen", Order: query.OrderAsc}),
			st.Options().Paginate(&query.Paginator{Page: -1, PerPage: -1}),
		)

		require.NoError(t, err)
		assert.Equal(t, 1, count)
		assert.Len(t, devices, 1)
		assert.Equal(t, models.DeviceStatusPending, devices[0].Status)
	})
}

// TestDeviceResolve tests device resolution by different keys
func (s *Suite) TestDeviceResolve(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when device not found by UID", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		device, err := st.DeviceResolve(ctx, store.DeviceUIDResolver, "nonexistent")
		assert.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, device)
	})

	t.Run("succeeds resolving device by UID", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create test device
		tenantID := s.CreateNamespace(t, WithNamespaceName("test-ns"))
		deviceUID := s.CreateDevice(t,
			WithDeviceName("test-device"),
			WithTenantID(tenantID),
			WithDeviceStatus(models.DeviceStatusAccepted),
		)

		// Resolve by UID
		device, err := st.DeviceResolve(ctx, store.DeviceUIDResolver, string(deviceUID))
		require.NoError(t, err)
		require.NotNil(t, device)
		assert.Equal(t, string(deviceUID), device.UID)
		assert.Equal(t, "test-device", device.Name)
		assert.Equal(t, tenantID, device.TenantID)
		assert.Equal(t, models.DeviceStatusAccepted, device.Status)
	})

	t.Run("succeeds resolving device by hostname", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create test device
		tenantID := s.CreateNamespace(t, WithNamespaceName("test-ns"))
		deviceUID := s.CreateDevice(t,
			WithDeviceName("my-hostname"),
			WithTenantID(tenantID),
		)

		// Resolve by hostname
		device, err := st.DeviceResolve(ctx, store.DeviceHostnameResolver, "my-hostname")
		require.NoError(t, err)
		require.NotNil(t, device)
		assert.Equal(t, string(deviceUID), device.UID)
		assert.Equal(t, "my-hostname", device.Name)
	})

	t.Run("succeeds resolving device by MAC", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create device - CreateDevice helper already sets a MAC address
		deviceUID := s.CreateDevice(t, WithDeviceName("mac-test-device"))

		// Get the device to find its MAC
		device, err := st.DeviceResolve(ctx, store.DeviceUIDResolver, string(deviceUID))
		require.NoError(t, err)
		require.NotNil(t, device)
		mac := device.Identity.MAC

		// Resolve by MAC
		deviceByMAC, err := st.DeviceResolve(ctx, store.DeviceMACResolver, mac)
		require.NoError(t, err)
		require.NotNil(t, deviceByMAC)
		assert.Equal(t, string(deviceUID), deviceByMAC.UID)
	})
}

// TestDeviceCreate tests device creation
func (s *Suite) TestDeviceCreate(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("succeeds when creating new device", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create namespace first
		tenantID := s.CreateNamespace(t)

		// Create device
		device := &models.Device{
			UID: "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
			Identity: &models.DeviceIdentity{
				MAC: "aa:bb:cc:dd:ee:ff",
			},
			TenantID:  tenantID,
			LastSeen:  clock.Now(),
			PublicKey: "-",
			Info:      &models.DeviceInfo{},
		}

		insertedUID, err := st.DeviceCreate(ctx, device)
		require.NoError(t, err)
		assert.Equal(t, "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c", insertedUID)

		// Verify it was created
		created, err := st.DeviceResolve(ctx, store.DeviceUIDResolver, insertedUID)
		require.NoError(t, err)
		assert.Equal(t, tenantID, created.TenantID)
	})
}

// TestDeviceConflicts tests checking for device conflicts
func (s *Suite) TestDeviceConflicts(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("no conflicts when target is empty", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create device
		s.CreateDevice(t, WithDeviceName("existing-device"))

		// Check with empty target
		conflicts, ok, err := st.DeviceConflicts(ctx, &models.DeviceConflicts{})
		require.NoError(t, err)
		assert.Empty(t, conflicts)
		assert.False(t, ok)
	})

	t.Run("no conflicts with non existing name", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create device
		s.CreateDevice(t, WithDeviceName("existing-device"))

		// Check with different name
		conflicts, ok, err := st.DeviceConflicts(ctx, &models.DeviceConflicts{Name: "nonexistent"})
		require.NoError(t, err)
		assert.Empty(t, conflicts)
		assert.False(t, ok)
	})

	t.Run("conflict detected with existing name", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create device with specific name
		s.CreateDevice(t, WithDeviceName("conflicting-device"))

		// Check for conflict with same name
		conflicts, ok, err := st.DeviceConflicts(ctx, &models.DeviceConflicts{
			Name: "conflicting-device",
		})
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"name"}, conflicts)
		assert.True(t, ok)
	})
}

// TestDeviceUpdate tests device updates
func (s *Suite) TestDeviceUpdate(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when device is not found due to uid", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)

		err := st.DeviceUpdate(ctx, &models.Device{
			UID:      "nonexistent",
			TenantID: tenantID,
		})
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("fails when device is not found due to tenantID", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create device
		deviceUID := s.CreateDevice(t, WithDeviceName("test-device"))

		// Use a nil UUID that will never be generated by the system
		nonExistentTenant := "00000000-0000-0000-0000-000000000000"

		err := st.DeviceUpdate(ctx, &models.Device{
			UID:      string(deviceUID),
			TenantID: nonExistentTenant,
		})
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("succeeds when device is found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create device
		tenantID := s.CreateNamespace(t)
		deviceUID := s.CreateDevice(t,
			WithDeviceName("original-name"),
			WithTenantID(tenantID),
		)

		// Update device
		err := st.DeviceUpdate(ctx, &models.Device{
			UID:      string(deviceUID),
			TenantID: tenantID,
			Name:     "updated-name",
		})
		require.NoError(t, err)

		// Verify update
		device, err := st.DeviceResolve(ctx, store.DeviceUIDResolver, string(deviceUID))
		require.NoError(t, err)
		assert.Equal(t, "updated-name", device.Name)
	})
}

// TestDeviceHeartbeat tests device heartbeat updates
func (s *Suite) TestDeviceHeartbeat(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("succeeds when no devices match", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create some devices
		s.CreateDevice(t, WithDeviceName("device-1"))
		s.CreateDevice(t, WithDeviceName("device-2"))

		// Try to heartbeat non-existent devices
		modifiedCount, err := st.DeviceHeartbeat(ctx,
			[]string{"nonexistent1", "nonexistent2"},
			time.Now(),
		)
		require.NoError(t, err)
		assert.Equal(t, int64(0), modifiedCount)
	})

	t.Run("succeeds when devices match", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create test devices
		uid1 := s.CreateDevice(t, WithDeviceName("device-1"))
		uid2 := s.CreateDevice(t, WithDeviceName("device-2"))

		// Heartbeat for devices
		newTime := time.Now()
		modifiedCount, err := st.DeviceHeartbeat(ctx,
			[]string{string(uid1), string(uid2)},
			newTime,
		)
		require.NoError(t, err)
		assert.Equal(t, int64(2), modifiedCount)
	})
}

// TestDeviceDelete tests device deletion
func (s *Suite) TestDeviceDelete(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when device is not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		err := st.DeviceDelete(ctx, &models.Device{
			UID: "nonexistent",
		})
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("succeeds when device is found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create device
		deviceUID := s.CreateDevice(t, WithDeviceName("test-device"))

		// Delete it
		err := st.DeviceDelete(ctx, &models.Device{
			UID: string(deviceUID),
		})
		require.NoError(t, err)

		// Verify deletion
		_, err = st.DeviceResolve(ctx, store.DeviceUIDResolver, string(deviceUID))
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})
}

// TestDeviceDeleteMany tests bulk device deletion
func (s *Suite) TestDeviceDeleteMany(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("succeeds when no devices match", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create devices
		s.CreateDevice(t, WithDeviceName("device-1"))
		s.CreateDevice(t, WithDeviceName("device-2"))

		// Delete empty list
		deletedCount, err := st.DeviceDeleteMany(ctx, []string{})
		require.NoError(t, err)
		assert.Equal(t, int64(0), deletedCount)
	})

	t.Run("succeeds when devices match", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create test devices
		uid1 := s.CreateDevice(t, WithDeviceName("device-1"))
		uid2 := s.CreateDevice(t, WithDeviceName("device-2"))
		uid3 := s.CreateDevice(t, WithDeviceName("device-3"))

		// Delete first two
		uids := []string{string(uid1), string(uid2)}
		deletedCount, err := st.DeviceDeleteMany(ctx, uids)
		require.NoError(t, err)
		assert.Equal(t, int64(2), deletedCount)

		// Verify deletions
		for _, uid := range uids {
			_, err := st.DeviceResolve(ctx, store.DeviceUIDResolver, uid)
			assert.ErrorIs(t, err, store.ErrNoDocuments)
		}

		// Verify remaining device
		device, err := st.DeviceResolve(ctx, store.DeviceUIDResolver, string(uid3))
		require.NoError(t, err)
		assert.Equal(t, string(uid3), device.UID)
	})

	t.Run("deletes related sessions in cascade", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create devices with sessions
		uid1 := s.CreateDevice(t, WithDeviceName("device-1"))
		uid2 := s.CreateDevice(t, WithDeviceName("device-2"))
		uid3 := s.CreateDevice(t, WithDeviceName("device-3"))

		// Create sessions for each device
		session1UID := s.CreateSession(t, WithSessionDevice(uid1))
		session2UID := s.CreateSession(t, WithSessionDevice(uid2))
		session3UID := s.CreateSession(t, WithSessionDevice(uid3))

		// Delete first two devices
		uids := []string{string(uid1), string(uid2)}
		deletedCount, err := st.DeviceDeleteMany(ctx, uids)
		require.NoError(t, err)
		assert.Equal(t, int64(2), deletedCount)

		// Verify sessions of deleted devices are gone
		_, err = st.SessionResolve(ctx, store.SessionUIDResolver, string(session1UID))
		assert.ErrorIs(t, err, store.ErrNoDocuments)
		_, err = st.SessionResolve(ctx, store.SessionUIDResolver, string(session2UID))
		assert.ErrorIs(t, err, store.ErrNoDocuments)

		// Verify session of remaining device still exists
		session3, err := st.SessionResolve(ctx, store.SessionUIDResolver, string(session3UID))
		require.NoError(t, err)
		assert.Equal(t, string(session3UID), session3.UID)
	})

	t.Run("succeeds with mix of existing and non-existing UIDs", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create some devices
		uid1 := s.CreateDevice(t, WithDeviceName("device-1"))
		uid2 := s.CreateDevice(t, WithDeviceName("device-2"))

		// Mix existing and non-existing UIDs
		uids := []string{
			string(uid1),
			"non-existent-uid-1",
			string(uid2),
			"non-existent-uid-2",
		}

		// Should only delete existing devices
		deletedCount, err := st.DeviceDeleteMany(ctx, uids)
		require.NoError(t, err)
		assert.Equal(t, int64(2), deletedCount)

		// Verify devices are deleted
		_, err = st.DeviceResolve(ctx, store.DeviceUIDResolver, string(uid1))
		assert.ErrorIs(t, err, store.ErrNoDocuments)
		_, err = st.DeviceResolve(ctx, store.DeviceUIDResolver, string(uid2))
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("succeeds with devices that have multiple sessions", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create device
		deviceUID := s.CreateDevice(t, WithDeviceName("busy-device"))

		// Create multiple sessions for the same device
		session1UID := s.CreateSession(t, WithSessionDevice(deviceUID), WithSessionUser("user1"))
		session2UID := s.CreateSession(t, WithSessionDevice(deviceUID), WithSessionUser("user2"))
		session3UID := s.CreateSession(t, WithSessionDevice(deviceUID), WithSessionUser("user3"))

		// Delete device
		deletedCount, err := st.DeviceDeleteMany(ctx, []string{string(deviceUID)})
		require.NoError(t, err)
		assert.Equal(t, int64(1), deletedCount)

		// Verify all sessions are deleted
		_, err = st.SessionResolve(ctx, store.SessionUIDResolver, string(session1UID))
		assert.ErrorIs(t, err, store.ErrNoDocuments)
		_, err = st.SessionResolve(ctx, store.SessionUIDResolver, string(session2UID))
		assert.ErrorIs(t, err, store.ErrNoDocuments)
		_, err = st.SessionResolve(ctx, store.SessionUIDResolver, string(session3UID))
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})
}

// TestDeviceStatusUpdatedAt tests that StatusUpdatedAt is persisted through create, resolve, and update
func (s *Suite) TestDeviceStatusUpdatedAt(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("persists StatusUpdatedAt through create and resolve", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		statusUpdatedAt := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)

		tenantID := s.CreateNamespace(t)
		deviceUID := s.CreateDevice(t,
			WithDeviceName("status-time-device"),
			WithTenantID(tenantID),
			WithDeviceStatusUpdatedAt(statusUpdatedAt),
		)

		device, err := st.DeviceResolve(ctx, store.DeviceUIDResolver, string(deviceUID))
		require.NoError(t, err)
		require.NotNil(t, device)
		assert.Equal(t, statusUpdatedAt, device.StatusUpdatedAt.UTC())
	})

	t.Run("persists StatusUpdatedAt through update", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		initialTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		updatedTime := time.Date(2025, 7, 20, 14, 45, 0, 0, time.UTC)

		tenantID := s.CreateNamespace(t)
		deviceUID := s.CreateDevice(t,
			WithDeviceName("update-status-time-device"),
			WithTenantID(tenantID),
			WithDeviceStatusUpdatedAt(initialTime),
		)

		// Update the device with a new StatusUpdatedAt
		err := st.DeviceUpdate(ctx, &models.Device{
			UID:             string(deviceUID),
			TenantID:        tenantID,
			Name:            "update-status-time-device",
			StatusUpdatedAt: updatedTime,
		})
		require.NoError(t, err)

		device, err := st.DeviceResolve(ctx, store.DeviceUIDResolver, string(deviceUID))
		require.NoError(t, err)
		require.NotNil(t, device)
		assert.Equal(t, updatedTime, device.StatusUpdatedAt.UTC())
	})
}
