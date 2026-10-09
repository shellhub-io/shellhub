package storetest

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// WithDeviceOwner ties the device to the member with the given user ID.
func WithDeviceOwner(userID string) DeviceOption {
	return func(d *models.Device) {
		d.OwnerID = userID
	}
}

func ownerFilter(ownerID string) *query.Filters {
	return &query.Filters{Data: []query.Filter{{
		Type:   query.FilterTypeProperty,
		Params: &query.FilterProperty{Name: "owner_id", Operator: "eq", Value: ownerID},
	}}}
}

func (s *Suite) ownedDeviceFixture(t *testing.T) (tenantID, memberID string) {
	t.Helper()

	tenantID = s.CreateNamespace(t)
	memberID = s.CreateUser(t)
	s.CreateMembership(t, tenantID, memberID, string(authorizer.RoleOperator))

	return tenantID, memberID
}

// TestDeviceSetOwner shows an owner is set and cleared on an accepted device, must be a member of
// the device's namespace, and cannot reach a device in another namespace.
func (s *Suite) TestDeviceSetOwner(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("ties an accepted device to a member and clears it again", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID, memberID := s.ownedDeviceFixture(t)
		uid := s.CreateDevice(t, WithTenantID(tenantID))
		sc := scope.MustBounded(tenantID)

		require.NoError(t, st.DeviceSetOwner(ctx, sc, string(uid), memberID))

		device, err := st.DeviceResolve(ctx, sc, store.DeviceUIDResolver, string(uid))
		require.NoError(t, err)
		assert.Equal(t, memberID, device.OwnerID)

		require.NoError(t, st.DeviceSetOwner(ctx, sc, string(uid), ""))

		device, err = st.DeviceResolve(ctx, sc, store.DeviceUIDResolver, string(uid))
		require.NoError(t, err)
		assert.Empty(t, device.OwnerID)
	})

	t.Run("refuses an owner who is not a member of the device's namespace", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		outsider := s.CreateUser(t)
		uid := s.CreateDevice(t, WithTenantID(tenantID))

		assert.Error(t, st.DeviceSetOwner(ctx, scope.MustBounded(tenantID), string(uid), outsider))
	})

	t.Run("reports no documents for a device in another namespace", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID, memberID := s.ownedDeviceFixture(t)
		uid := s.CreateDevice(t)

		err := st.DeviceSetOwner(ctx, scope.MustBounded(tenantID), string(uid), memberID)
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})
}

// TestDeviceUpdateDoesNotClobberOwner shows the owner is written only through DeviceSetOwner and
// NamespaceDepartMember, so a stale snapshot written back by DeviceUpdate cannot restore or drop it.
func (s *Suite) TestDeviceUpdateDoesNotClobberOwner(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	require.NoError(t, s.provider.CleanDatabase(t))

	tenantID, memberID := s.ownedDeviceFixture(t)
	uid := s.CreateDevice(t, WithTenantID(tenantID))
	sc := scope.MustBounded(tenantID)

	stale, err := st.DeviceResolve(ctx, sc, store.DeviceUIDResolver, string(uid))
	require.NoError(t, err)

	require.NoError(t, st.DeviceSetOwner(ctx, sc, string(uid), memberID))
	require.NoError(t, st.DeviceUpdate(ctx, stale))

	device, err := st.DeviceResolve(ctx, sc, store.DeviceUIDResolver, string(uid))
	require.NoError(t, err)
	assert.Equal(t, memberID, device.OwnerID)
}

// TestDeviceListFiltersByOwner shows the owner_id filter matches only the devices that member paired.
func (s *Suite) TestDeviceListFiltersByOwner(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	require.NoError(t, s.provider.CleanDatabase(t))

	tenantID, memberID := s.ownedDeviceFixture(t)
	owned := s.CreateDevice(t, WithTenantID(tenantID), WithDeviceOwner(memberID))
	s.CreateDevice(t, WithTenantID(tenantID))

	devices, count, err := st.DeviceList(ctx, scope.MustBounded(tenantID), store.DeviceAcceptableIfNotAccepted,
		st.Options().Match(ownerFilter(memberID)))
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
	require.Len(t, devices, 1)
	assert.Equal(t, string(owned), devices[0].UID)
}

// TestDeviceOwnerMustBeAMember shows the database keeps an owner a current member: a membership
// that still owns an accepted device cannot be deleted, while a namespace holding one can.
func (s *Suite) TestDeviceOwnerMustBeAMember(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("a membership that owns a device cannot be deleted", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID, memberID := s.ownedDeviceFixture(t)
		s.CreateDevice(t, WithTenantID(tenantID), WithDeviceOwner(memberID))

		err := st.NamespaceDeleteMembership(ctx, scope.MustBounded(tenantID), &models.Member{ID: memberID})
		assert.Error(t, err)
	})

	t.Run("a namespace holding an owned device can be deleted", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID, memberID := s.ownedDeviceFixture(t)
		s.CreateDevice(t, WithTenantID(tenantID), WithDeviceOwner(memberID))

		ns, err := st.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenantID)
		require.NoError(t, err)
		assert.NoError(t, st.NamespaceDelete(ctx, ns))
	})
}

// TestNamespaceDepartMember shows a departure removes only the member's own devices in that
// namespace, revokes their keys, keeps what it is told to keep, and changes nothing when it fails.
func (s *Suite) TestNamespaceDepartMember(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	counts := func(t *testing.T, tenantID string) (accepted, removed int64) {
		t.Helper()

		ns, err := st.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenantID)
		require.NoError(t, err)

		return ns.DevicesAcceptedCount, ns.DevicesRemovedCount
	}

	t.Run("removes the member's devices, keys and membership", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID, memberID := s.ownedDeviceFixture(t)
		otherTenantID := s.CreateNamespace(t)
		s.CreateMembership(t, otherTenantID, memberID, string(authorizer.RoleOperator))
		otherMemberID := s.CreateUser(t)
		s.CreateMembership(t, tenantID, otherMemberID, string(authorizer.RoleOperator))
		sc := scope.MustBounded(tenantID)

		owned := s.CreateDevice(t, WithTenantID(tenantID), WithDeviceOwner(memberID))
		team := s.CreateDevice(t, WithTenantID(tenantID))
		othersDevice := s.CreateDevice(t, WithTenantID(tenantID), WithDeviceOwner(otherMemberID))
		elsewhere := s.CreateDevice(t, WithTenantID(otherTenantID), WithDeviceOwner(memberID))
		require.NoError(t, st.NamespaceIncrementDeviceCount(ctx, sc, models.DeviceStatusAccepted, 3))

		digest := s.CreateAPIKey(t, WithAPIKeyTenant(tenantID), WithAPIKeyCreatedBy(memberID))

		departed, err := st.NamespaceDepartMember(ctx, sc, memberID, store.MemberDeparture{})
		require.NoError(t, err)
		assert.Equal(t, []string{string(owned)}, departed.RemovedDevices)
		assert.Equal(t, []string{digest}, departed.APIKeyDigests)

		device, err := st.DeviceResolve(ctx, sc, store.DeviceUIDResolver, string(owned))
		require.NoError(t, err)
		assert.Equal(t, models.DeviceStatusRemoved, device.Status)
		assert.NotNil(t, device.RemovedAt)
		assert.Empty(t, device.OwnerID)

		for uid, want := range map[models.UID]string{team: "", othersDevice: otherMemberID} {
			device, err := st.DeviceResolve(ctx, sc, store.DeviceUIDResolver, string(uid))
			require.NoError(t, err)
			assert.Equal(t, models.DeviceStatusAccepted, device.Status)
			assert.Equal(t, want, device.OwnerID)
		}

		device, err = st.DeviceResolve(ctx, scope.MustBounded(otherTenantID), store.DeviceUIDResolver, string(elsewhere))
		require.NoError(t, err)
		assert.Equal(t, models.DeviceStatusAccepted, device.Status)
		assert.Equal(t, memberID, device.OwnerID)

		accepted, removed := counts(t, tenantID)
		assert.Equal(t, int64(2), accepted)
		assert.Equal(t, int64(1), removed)

		ns, err := st.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenantID)
		require.NoError(t, err)
		_, stillMember := ns.FindMember(memberID)
		assert.False(t, stillMember)
	})

	t.Run("keeps the membership when told to", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID, memberID := s.ownedDeviceFixture(t)
		sc := scope.MustBounded(tenantID)
		owned := s.CreateDevice(t, WithTenantID(tenantID), WithDeviceOwner(memberID))

		departed, err := st.NamespaceDepartMember(ctx, sc, memberID, store.MemberDeparture{KeepMembership: true})
		require.NoError(t, err)
		assert.Equal(t, []string{string(owned)}, departed.RemovedDevices)

		ns, err := st.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenantID)
		require.NoError(t, err)
		_, stillMember := ns.FindMember(memberID)
		assert.True(t, stillMember)
	})

	t.Run("keeps the listed devices as team devices", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID, memberID := s.ownedDeviceFixture(t)
		sc := scope.MustBounded(tenantID)
		kept := s.CreateDevice(t, WithTenantID(tenantID), WithDeviceOwner(memberID))
		gone := s.CreateDevice(t, WithTenantID(tenantID), WithDeviceOwner(memberID))

		departed, err := st.NamespaceDepartMember(ctx, sc, memberID, store.MemberDeparture{KeepDevices: []string{string(kept)}})
		require.NoError(t, err)
		assert.Equal(t, []string{string(gone)}, departed.RemovedDevices)

		device, err := st.DeviceResolve(ctx, sc, store.DeviceUIDResolver, string(kept))
		require.NoError(t, err)
		assert.Equal(t, models.DeviceStatusAccepted, device.Status)
		assert.Empty(t, device.OwnerID)
	})

	t.Run("refuses to keep a device the member does not own, changing nothing", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID, memberID := s.ownedDeviceFixture(t)
		sc := scope.MustBounded(tenantID)
		owned := s.CreateDevice(t, WithTenantID(tenantID), WithDeviceOwner(memberID))
		team := s.CreateDevice(t, WithTenantID(tenantID))

		_, err := st.NamespaceDepartMember(ctx, sc, memberID, store.MemberDeparture{KeepDevices: []string{string(team)}})
		require.ErrorIs(t, err, store.ErrDeviceNotOwned)

		device, err := st.DeviceResolve(ctx, sc, store.DeviceUIDResolver, string(owned))
		require.NoError(t, err)
		assert.Equal(t, models.DeviceStatusAccepted, device.Status)
		assert.Equal(t, memberID, device.OwnerID)

		ns, err := st.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenantID)
		require.NoError(t, err)
		_, stillMember := ns.FindMember(memberID)
		assert.True(t, stillMember)
	})

	t.Run("clears the owner a removed device kept, so the membership can go", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID, memberID := s.ownedDeviceFixture(t)
		sc := scope.MustBounded(tenantID)
		stranded := s.CreateDevice(t, WithTenantID(tenantID), WithDeviceOwner(memberID),
			WithDeviceStatus(models.DeviceStatusRemoved))

		departed, err := st.NamespaceDepartMember(ctx, sc, memberID, store.MemberDeparture{})
		require.NoError(t, err)
		assert.Empty(t, departed.RemovedDevices, "a device that was already removed does not leave again")

		device, err := st.DeviceResolve(ctx, sc, store.DeviceUIDResolver, string(stranded))
		require.NoError(t, err)
		assert.Equal(t, models.DeviceStatusRemoved, device.Status)
		assert.Empty(t, device.OwnerID)
	})

	t.Run("joins the transaction carried by the context", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID, memberID := s.ownedDeviceFixture(t)
		sc := scope.MustBounded(tenantID)
		owned := s.CreateDevice(t, WithTenantID(tenantID), WithDeviceOwner(memberID))

		rollback := assert.AnError
		err := st.WithTransaction(ctx, func(ctx context.Context) error {
			if _, err := st.NamespaceDepartMember(ctx, sc, memberID, store.MemberDeparture{}); err != nil {
				return err
			}

			return rollback
		})
		require.ErrorIs(t, err, rollback)

		device, err := st.DeviceResolve(ctx, sc, store.DeviceUIDResolver, string(owned))
		require.NoError(t, err)
		assert.Equal(t, models.DeviceStatusAccepted, device.Status)
		assert.Equal(t, memberID, device.OwnerID)
	})

	t.Run("reports no documents for someone who is not a member", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)

		_, err := st.NamespaceDepartMember(ctx, scope.MustBounded(tenantID), s.CreateUser(t), store.MemberDeparture{})
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})
}

// TestDeviceUpdateUnlessRemoved shows the guarded update writes a live device and leaves a removed
// one alone, so a stale snapshot cannot bring a device back.
func (s *Suite) TestDeviceUpdateUnlessRemoved(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("writes a device that is not removed", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		uid := s.CreateDevice(t, WithTenantID(tenantID))
		sc := scope.MustBounded(tenantID)

		device, err := st.DeviceResolve(ctx, sc, store.DeviceUIDResolver, string(uid))
		require.NoError(t, err)

		device.RemoteAddr = "10.0.0.9"
		require.NoError(t, st.DeviceUpdateUnlessRemoved(ctx, device))

		device, err = st.DeviceResolve(ctx, sc, store.DeviceUIDResolver, string(uid))
		require.NoError(t, err)
		assert.Equal(t, "10.0.0.9", device.RemoteAddr)
	})

	t.Run("leaves a device removed since it was read", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID, memberID := s.ownedDeviceFixture(t)
		uid := s.CreateDevice(t, WithTenantID(tenantID), WithDeviceOwner(memberID))
		sc := scope.MustBounded(tenantID)

		stale, err := st.DeviceResolve(ctx, sc, store.DeviceUIDResolver, string(uid))
		require.NoError(t, err)

		_, err = st.NamespaceDepartMember(ctx, sc, memberID, store.MemberDeparture{KeepMembership: true})
		require.NoError(t, err)

		require.ErrorIs(t, st.DeviceUpdateUnlessRemoved(ctx, stale), store.ErrNoDocuments)

		device, err := st.DeviceResolve(ctx, sc, store.DeviceUIDResolver, string(uid))
		require.NoError(t, err)
		assert.Equal(t, models.DeviceStatusRemoved, device.Status)
	})
}
