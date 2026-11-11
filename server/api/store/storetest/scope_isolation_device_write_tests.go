package storetest

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type deviceWrite struct {
	description string
	prepare     func(t *testing.T, s *Suite, uid models.UID)
	write       func(ctx context.Context, st store.Store, sc scope.Scope, device *models.Device) error
	written     func(device *models.Device) bool
}

func deviceWrites() []deviceWrite {
	renamed := func(device *models.Device) bool { return device != nil && device.Name == "renamed" }

	return []deviceWrite{
		{
			description: "DeviceUpdate",
			write: func(ctx context.Context, st store.Store, sc scope.Scope, device *models.Device) error {
				device.Name = "renamed"

				return st.DeviceUpdate(ctx, sc, device)
			},
			written: renamed,
		},
		{
			description: "DeviceUpdateUnlessRemoved",
			write: func(ctx context.Context, st store.Store, sc scope.Scope, device *models.Device) error {
				device.Name = "renamed"

				return st.DeviceUpdateUnlessRemoved(ctx, sc, device)
			},
			written: renamed,
		},
		{
			description: "DeviceAssertNotRemoved",
			write: func(ctx context.Context, st store.Store, sc scope.Scope, device *models.Device) error {
				return st.DeviceAssertNotRemoved(ctx, sc, device)
			},
		},
		{
			description: "DeviceOffline",
			write: func(ctx context.Context, st store.Store, sc scope.Scope, device *models.Device) error {
				return st.DeviceOffline(ctx, sc, device.UID, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
			},
			written: func(device *models.Device) bool { return device != nil && device.DisconnectedAt != nil },
		},
		{
			description: "DeviceSetCustomField",
			write: func(ctx context.Context, st store.Store, sc scope.Scope, device *models.Device) error {
				return st.DeviceSetCustomField(ctx, sc, device.UID, "key", "value")
			},
			written: func(device *models.Device) bool { return device != nil && device.CustomFields["key"] == "value" },
		},
		{
			description: "DeviceDeleteCustomField",
			prepare: func(t *testing.T, s *Suite, uid models.UID) {
				t.Helper()

				require.NoError(t, s.provider.Store().DeviceSetCustomField(context.Background(), s.scopeOf(t, uid), string(uid), "key", "value"))
			},
			write: func(ctx context.Context, st store.Store, sc scope.Scope, device *models.Device) error {
				return st.DeviceDeleteCustomField(ctx, sc, device.UID, "key")
			},
			written: func(device *models.Device) bool {
				if device == nil {
					return false
				}

				_, kept := device.CustomFields["key"]

				return !kept
			},
		},
		{
			description: "DeviceDelete",
			write: func(ctx context.Context, st store.Store, sc scope.Scope, device *models.Device) error {
				return st.DeviceDelete(ctx, sc, device)
			},
			written: func(device *models.Device) bool { return device == nil },
		},
	}
}

// TestScopeIsolationDeviceWrites locks that every single-device write leaves a device of another
// namespace untouched, refuses an unbounded scope before writing, and writes within the owner's.
func (s *Suite) TestScopeIsolationDeviceWrites(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	for _, tc := range deviceWrites() {
		t.Run(tc.description, func(t *testing.T) {
			require.NoError(t, s.provider.CleanDatabase(t))

			owner := s.CreateNamespace(t)
			other := s.CreateNamespace(t)
			uid := s.CreateDevice(t, WithTenantID(owner), WithDeviceName("dev"))

			if tc.prepare != nil {
				tc.prepare(t, s, uid)
			}

			snapshot := func(tenantID string) *models.Device {
				device, err := st.DeviceResolve(ctx, scope.MustBounded(owner), store.DeviceUIDResolver, string(uid))
				require.NoError(t, err)

				device.TenantID = tenantID

				return device
			}

			current := func() *models.Device {
				device, err := st.DeviceResolve(ctx, scope.MustBounded(owner), store.DeviceUIDResolver, string(uid))
				if err != nil {
					require.ErrorIs(t, err, store.ErrNoDocuments)

					return nil
				}

				return device
			}

			unchanged := func() {
				t.Helper()

				if tc.written != nil {
					assert.False(t, tc.written(current()), "the device must be left as it was")
				}
			}

			require.ErrorIs(t, tc.write(ctx, st, scope.MustBounded(other), snapshot(other)), store.ErrNoDocuments)
			unchanged()

			require.ErrorIs(t, tc.write(ctx, st, scope.NewUnbounded(reasonTestQueryMechanics), snapshot(owner)), store.ErrInvalidScope)
			unchanged()

			require.ErrorIs(t, tc.write(ctx, st, scope.Scope{}, snapshot(owner)), store.ErrInvalidScope)
			unchanged()

			require.NoError(t, tc.write(ctx, st, scope.MustBounded(owner), snapshot(owner)))
			if tc.written != nil {
				assert.True(t, tc.written(current()))
			}
		})
	}
}

// TestScopeIsolationDeviceModelTenant locks that a model-taking device write refuses a model whose
// tenant is not the bounded scope's, so the model cannot move the write into another namespace.
func (s *Suite) TestScopeIsolationDeviceModelTenant(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)
	uid := s.CreateDevice(t, WithTenantID(owner), WithDeviceName("dev"))

	device, err := st.DeviceResolve(ctx, scope.MustBounded(owner), store.DeviceUIDResolver, string(uid))
	require.NoError(t, err)

	device.Name = "renamed"

	require.ErrorIs(t, st.DeviceUpdate(ctx, scope.MustBounded(other), device), store.ErrInvalidScope)
	require.ErrorIs(t, st.DeviceUpdateUnlessRemoved(ctx, scope.MustBounded(other), device), store.ErrInvalidScope)
	require.ErrorIs(t, st.DeviceAssertNotRemoved(ctx, scope.MustBounded(other), device), store.ErrInvalidScope)
	require.ErrorIs(t, st.DeviceDelete(ctx, scope.MustBounded(other), device), store.ErrInvalidScope)

	kept, err := st.DeviceResolve(ctx, scope.MustBounded(owner), store.DeviceUIDResolver, string(uid))
	require.NoError(t, err)
	assert.Equal(t, "dev", kept.Name)

	for _, spelling := range []string{
		strings.ToUpper(owner),
		"{" + owner + "}",
		strings.ReplaceAll(owner, "-", ""),
	} {
		device.TenantID = spelling
		require.NoError(t, st.DeviceUpdate(ctx, scope.MustBounded(owner), device), "%s names the owner's namespace", spelling)
		require.NoError(t, st.DeviceUpdate(ctx, scope.MustBounded(spelling), device), "%s bounds to the owner's namespace", spelling)

		gone, err := st.DeviceHeartbeat(ctx, scope.MustBounded(spelling),
			[]store.DeviceBeat{{UID: string(uid), TenantID: spelling, At: s.Clock().Now()}}, 0)
		require.NoError(t, err)
		assert.Empty(t, gone, "a beat whose tenant is spelled %s finds its device", spelling)
	}
}

// TestScopeIsolationDeviceCreate locks that DeviceCreate inserts only into the namespace its
// bounded scope names.
func (s *Suite) TestScopeIsolationDeviceCreate(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)

	device := func() *models.Device {
		return &models.Device{
			UID:       "device-" + uniqueHex(t, 8),
			TenantID:  owner,
			Name:      "dev",
			Identity:  &models.DeviceIdentity{MAC: nextDeviceMAC()},
			Info:      &models.DeviceInfo{},
			PublicKey: "-",
			Status:    models.DeviceStatusPending,
		}
	}

	_, err := st.DeviceCreate(ctx, scope.MustBounded(other), device())
	require.ErrorIs(t, err, store.ErrInvalidScope)

	_, err = st.DeviceCreate(ctx, scope.NewUnbounded(reasonTestQueryMechanics), device())
	require.ErrorIs(t, err, store.ErrInvalidScope)

	_, count, err := st.DeviceList(ctx, scope.NewUnbounded(reasonTestQueryMechanics), store.DeviceAcceptableAsFalse)
	require.NoError(t, err)
	require.Zero(t, count, "a refused create must write nothing")

	created := device()
	uid, err := st.DeviceCreate(ctx, scope.MustBounded(owner), created)
	require.NoError(t, err)
	assert.Equal(t, created.UID, uid)

	_, err = st.DeviceResolve(ctx, scope.MustBounded(owner), store.DeviceUIDResolver, uid)
	require.NoError(t, err, "the device lands in the scope's namespace")

	_, err = st.DeviceResolve(ctx, scope.MustBounded(other), store.DeviceUIDResolver, uid)
	require.ErrorIs(t, err, store.ErrNoDocuments, "the device is not in any other namespace")
}

// TestScopeIsolationDeviceHeartbeat locks that a beat stamps its device only within the beat's own
// namespace, and only within the scope when the scope is bounded.
func (s *Suite) TestScopeIsolationDeviceHeartbeat(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)
	seen := s.Clock().Now().UTC().Add(-time.Hour).Truncate(time.Second)
	uid := s.CreateDevice(t, WithTenantID(owner), WithDeviceName("dev"), WithDeviceLastSeen(seen))
	beat := s.Clock().Now().UTC().Truncate(time.Second)

	lastSeen := func() time.Time {
		device, err := st.DeviceResolve(ctx, scope.MustBounded(owner), store.DeviceUIDResolver, string(uid))
		require.NoError(t, err)

		return device.LastSeen
	}

	gone, err := st.DeviceHeartbeat(ctx, scope.NewUnbounded(reasonTestQueryMechanics),
		[]store.DeviceBeat{{UID: string(uid), TenantID: other, At: beat}}, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{string(uid)}, gone, "a beat naming another namespace finds no device")
	assert.WithinDuration(t, seen, lastSeen(), time.Millisecond)

	gone, err = st.DeviceHeartbeat(ctx, scope.MustBounded(other),
		[]store.DeviceBeat{{UID: string(uid), TenantID: owner, At: beat}}, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{string(uid)}, gone, "a scope bounded to another namespace finds no device")
	assert.WithinDuration(t, seen, lastSeen(), time.Millisecond)

	_, err = st.DeviceHeartbeat(ctx, scope.Scope{}, []store.DeviceBeat{{UID: string(uid), TenantID: owner, At: beat}}, 0)
	require.ErrorIs(t, err, store.ErrInvalidScope)
	assert.WithinDuration(t, seen, lastSeen(), time.Millisecond)

	gone, err = st.DeviceHeartbeat(ctx, scope.MustBounded(owner),
		[]store.DeviceBeat{{UID: string(uid), TenantID: owner, At: beat}}, 0)
	require.NoError(t, err)
	assert.Empty(t, gone)
	assert.WithinDuration(t, beat, lastSeen(), time.Millisecond)
}

// TestScopeIsolationDeviceDeleteMany locks that a bounded DeviceDeleteMany leaves a listed device
// of another namespace, and its sessions, in place.
func (s *Suite) TestScopeIsolationDeviceDeleteMany(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)
	owned := s.CreateDevice(t, WithTenantID(owner), WithDeviceName("owned"))
	foreign := s.CreateDevice(t, WithTenantID(other), WithDeviceName("foreign"))
	foreignSession := s.CreateSession(t, WithSessionDevice(foreign))

	_, err := st.DeviceDeleteMany(ctx, scope.Scope{}, []string{string(owned), string(foreign)})
	require.ErrorIs(t, err, store.ErrInvalidScope)

	deleted, err := st.DeviceDeleteMany(ctx, scope.MustBounded(owner), []string{string(owned), string(foreign)})
	require.NoError(t, err)
	require.Len(t, deleted, 1)
	assert.Equal(t, string(owned), deleted[0].UID)

	_, err = st.DeviceResolve(ctx, scope.MustBounded(other), store.DeviceUIDResolver, string(foreign))
	require.NoError(t, err, "a device of another namespace must survive")

	_, err = st.SessionResolve(ctx, scope.MustBounded(other), store.SessionUIDResolver, string(foreignSession))
	require.NoError(t, err, "a session of a device that survived must survive")
}

// TestScopeIsolationDeviceMaintenance locks that the presence sweep and the expired ephemeral
// listing stay within a bounded scope and refuse an unconstructed one.
func (s *Suite) TestScopeIsolationDeviceMaintenance(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	now := s.Clock().Now()
	stale := []DeviceOption{
		WithDeviceEphemeral(5),
		WithDeviceStatusUpdatedAt(now.Add(-time.Hour)),
		WithDeviceLastSeen(now.Add(-time.Hour)),
	}

	t.Run("DeviceSweepStalePresence", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		owner := s.CreateNamespace(t)
		other := s.CreateNamespace(t)
		s.CreateDevice(t, append(stale, WithTenantID(owner), WithDeviceName("owned"))...)
		foreign := s.CreateDevice(t, append(stale, WithTenantID(other), WithDeviceName("foreign"))...)

		_, err := st.DeviceSweepStalePresence(ctx, scope.Scope{}, now.Add(-10*time.Minute), 1000)
		require.ErrorIs(t, err, store.ErrInvalidScope)

		stamped, err := st.DeviceSweepStalePresence(ctx, scope.MustBounded(owner), now.Add(-10*time.Minute), 1000)
		require.NoError(t, err)
		assert.Equal(t, int64(1), stamped)

		device, err := st.DeviceResolve(ctx, scope.MustBounded(other), store.DeviceUIDResolver, string(foreign))
		require.NoError(t, err)
		assert.Nil(t, device.DisconnectedAt, "a device of another namespace must be left unswept")
	})

	t.Run("DeviceListExpiredEphemeral", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		owner := s.CreateNamespace(t)
		other := s.CreateNamespace(t)
		owned := s.CreateDevice(t, append(stale, WithTenantID(owner), WithDeviceName("owned"))...)
		foreign := s.CreateDevice(t, append(stale, WithTenantID(other), WithDeviceName("foreign"))...)

		_, err := st.DeviceListExpiredEphemeral(ctx, scope.Scope{})
		require.ErrorIs(t, err, store.ErrInvalidScope)

		for tenantID, want := range map[string]string{owner: string(owned), other: string(foreign)} {
			devices, err := st.DeviceListExpiredEphemeral(ctx, scope.MustBounded(tenantID))
			require.NoError(t, err)
			require.Len(t, devices, 1)
			assert.Equal(t, want, devices[0].UID)
		}
	})
}
