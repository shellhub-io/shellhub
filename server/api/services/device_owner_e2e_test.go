package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type memoryCache struct {
	storecache.Cache
	mu      sync.Mutex
	entries map[string][]byte
}

func newMemoryCache() *memoryCache {
	return &memoryCache{Cache: storecache.NewNullCache(), entries: map[string][]byte{}}
}

func (c *memoryCache) Get(_ context.Context, key string, value any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	raw, ok := c.entries[key]
	if !ok {
		return storecache.ErrGetNotFound
	}

	return json.Unmarshal(raw, value)
}

func (c *memoryCache) Set(_ context.Context, key string, value any, _ time.Duration) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = raw

	return nil
}

func (c *memoryCache) SetNX(ctx context.Context, key string, value any, ttl time.Duration) (bool, error) {
	c.mu.Lock()
	_, exists := c.entries[key]
	c.mu.Unlock()

	if exists {
		return false, nil
	}

	return true, c.Set(ctx, key, value, ttl)
}

func (c *memoryCache) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)

	return nil
}

type ownershipE2E struct {
	*enrollmentE2E
	ownerID string
}

func setupOwnershipE2E(t *testing.T) *ownershipE2E {
	t.Helper()

	e := setupEnrollmentE2E(t)
	e.svc = NewService(e.st, privateKey, publicKey, newMemoryCache())

	ns, err := e.st.NamespaceResolve(context.Background(), store.NamespaceTenantIDResolver, e.tenantID)
	require.NoError(t, err)

	return &ownershipE2E{enrollmentE2E: e, ownerID: ns.Owner}
}

func (e *ownershipE2E) member(t *testing.T, name string, role authorizer.Role) string {
	t.Helper()
	ctx := context.Background()

	id, err := e.st.UserCreate(ctx, &models.User{
		Origin:        models.UserOriginLocal,
		Status:        models.UserStatusConfirmed,
		MaxNamespaces: -1,
		UserData:      models.UserData{Name: name, Email: name + "@example.com", Username: name},
		Password:      models.UserPassword{Hash: "hash"},
	})
	require.NoError(t, err)

	require.NoError(t, e.st.NamespaceCreateMembership(ctx, scope.MustBounded(e.tenantID), &models.Member{ID: id, Role: role}))

	return id
}

func pairingRequest(mac, publicKey string) *requests.DevicePairingCreate {
	return &requests.DevicePairingCreate{
		Hostname:  "host-" + mac,
		PublicKey: publicKey,
		Identity:  &requests.DeviceIdentity{MAC: mac},
		Info:      &requests.DeviceInfo{ID: "debian", PrettyName: "Debian", Version: "v0.1.0", Arch: "amd64", Platform: "docker"},
	}
}

func (e *ownershipE2E) pair(t *testing.T, userID, mac, publicKey string) *models.DevicePairingAccepted {
	t.Helper()
	ctx := context.Background()

	pairing, err := e.svc.CreateDevicePairing(ctx, pairingRequest(mac, publicKey))
	require.NoError(t, err)

	accepted, err := e.svc.AcceptDevicePairing(ctx, userID, &requests.DevicePairingAccept{Code: pairing.Code, TenantID: e.tenantID})
	require.NoError(t, err)

	return accepted
}

func (e *ownershipE2E) enrollKey(t *testing.T, mac, publicKey string) string {
	t.Helper()

	res, err := e.svc.AuthDevice(context.Background(), e.authRequest(mac, publicKey))
	require.NoError(t, err)

	return res.UID
}

func (e *ownershipE2E) authRequest(mac, publicKey string) requests.DeviceAuth {
	return requests.DeviceAuth{
		TenantID:  e.tenantID,
		Hostname:  "host-" + mac,
		Identity:  &requests.DeviceIdentity{MAC: mac},
		Info:      &requests.DeviceInfo{ID: "debian", PrettyName: "Debian", Version: "v0.1.0", Arch: "amd64", Platform: "docker"},
		PublicKey: publicKey,
	}
}

func (e *ownershipE2E) accept(t *testing.T, uid string) {
	t.Helper()

	require.NoError(t, e.svc.UpdateDeviceStatus(context.Background(), &requests.DeviceUpdateStatus{
		TenantID: e.tenantID,
		UID:      uid,
		Status:   string(models.DeviceStatusAccepted),
	}))
}

func (e *ownershipE2E) accepted(t *testing.T) []models.Device {
	t.Helper()

	devices, _, err := e.st.DeviceList(context.Background(), scope.MustBounded(e.tenantID), store.DeviceAcceptableAsFalse,
		e.st.Options().WithDeviceStatus(models.DeviceStatusAccepted))
	require.NoError(t, err)

	return devices
}

func TestPairingTiesTheDeviceToTheAcceptingMember(t *testing.T) {
	e := setupOwnershipE2E(t)
	member := e.member(t, "member", authorizer.RoleOperator)

	accepted := e.pair(t, member, "aa:bb:cc:dd:ee:01", "pk-1")

	assert.Equal(t, member, accepted.OwnerID)
	device := e.device(t, accepted.UID)
	assert.Equal(t, models.DeviceStatusAccepted, device.Status)
	assert.Equal(t, member, device.OwnerID)
}

func TestNamespaceCredentialsEnrollTeamDevices(t *testing.T) {
	e := setupOwnershipE2E(t)

	uid := e.enrollKey(t, "aa:bb:cc:dd:ee:01", "pk-1")
	e.accept(t, uid)

	device := e.device(t, uid)
	assert.Equal(t, models.DeviceStatusAccepted, device.Status)
	assert.Empty(t, device.OwnerID)
}

func TestMergeOnAcceptGivesTheTeamPriority(t *testing.T) {
	const mac = "aa:bb:cc:dd:ee:01"

	t.Run("an owned device merging into a team device stays the team's", func(t *testing.T) {
		e := setupOwnershipE2E(t)
		member := e.member(t, "member", authorizer.RoleOperator)

		team := e.enrollKey(t, mac, "pk-old")
		e.accept(t, team)
		e.disconnect(t, team)

		accepted := e.pair(t, member, mac, "pk-new")

		assert.Empty(t, accepted.OwnerID)
		devices := e.accepted(t)
		require.Len(t, devices, 1)
		assert.Equal(t, accepted.UID, devices[0].UID)
		assert.Empty(t, devices[0].OwnerID)
	})

	t.Run("a team device merging into an owned device makes it the team's", func(t *testing.T) {
		e := setupOwnershipE2E(t)
		member := e.member(t, "member", authorizer.RoleOperator)

		owned := e.pair(t, member, mac, "pk-old")
		e.disconnect(t, owned.UID)

		uid := e.enrollKey(t, mac, "pk-new")
		e.accept(t, uid)

		devices := e.accepted(t)
		require.Len(t, devices, 1)
		assert.Equal(t, uid, devices[0].UID)
		assert.Empty(t, devices[0].OwnerID)
	})

	t.Run("an owned device merging into another member's device changes owner", func(t *testing.T) {
		e := setupOwnershipE2E(t)
		first := e.member(t, "first", authorizer.RoleOperator)
		second := e.member(t, "second", authorizer.RoleOperator)

		old := e.pair(t, first, mac, "pk-old")
		e.disconnect(t, old.UID)

		accepted := e.pair(t, second, mac, "pk-new")

		assert.Equal(t, second, accepted.OwnerID)
		devices := e.accepted(t)
		require.Len(t, devices, 1)
		assert.Equal(t, second, devices[0].OwnerID)
	})
}

func TestPairingRefusesToMergeAwayAConnectedDevice(t *testing.T) {
	const mac = "aa:bb:cc:dd:ee:02"
	ctx := context.Background()
	e := setupOwnershipE2E(t)
	first := e.member(t, "first", authorizer.RoleOperator)
	member := e.member(t, "member", authorizer.RoleOperator)

	old := e.pair(t, first, mac, "pk-old").UID

	pairing, err := e.svc.CreateDevicePairing(ctx, pairingRequest(mac, "pk-new"))
	require.NoError(t, err)

	_, err = e.svc.AcceptDevicePairing(ctx, member, &requests.DevicePairingAccept{Code: pairing.Code, TenantID: e.tenantID})
	require.ErrorIs(t, err, ErrDeviceMACConnected)

	devices := e.accepted(t)
	require.Len(t, devices, 1)
	assert.Equal(t, old, devices[0].UID)

	t.Run("pairing again once the old agent is gone merges", func(t *testing.T) {
		e.disconnect(t, old)

		pending, err := e.st.DeviceResolve(ctx, scope.MustBounded(e.tenantID), store.DeviceMACResolver, mac,
			e.st.Options().WithDeviceStatus(models.DeviceStatusPending))
		require.NoError(t, err)
		require.NoError(t, e.svc.cache.Delete(ctx, deviceAuthCacheKey(pending.UID)), "the approval comes back after the auth cache expired")

		accepted := e.pair(t, member, mac, "pk-new")

		devices := e.accepted(t)
		require.Len(t, devices, 1)
		assert.Equal(t, accepted.UID, devices[0].UID)
		assert.Equal(t, member, devices[0].OwnerID, "a retried pairing still ties the device to the member who approved it")
	})
}

func TestAnAcceptedDeviceKeepsItsOwner(t *testing.T) {
	ctx := context.Background()
	e := setupOwnershipE2E(t)
	first := e.member(t, "first", authorizer.RoleOperator)

	accepted := e.pair(t, first, "aa:bb:cc:dd:ee:01", "pk-1")

	pairing, err := e.svc.CreateDevicePairing(ctx, pairingRequest("aa:bb:cc:dd:ee:01", "pk-1"))
	require.NoError(t, err)
	assert.Equal(t, models.DeviceStatusAccepted, pairing.Status)
	assert.Empty(t, pairing.Code)

	assert.Equal(t, first, e.device(t, accepted.UID).OwnerID)
}

func TestARemovedPairedDeviceGoesBackToPairing(t *testing.T) {
	ctx := context.Background()

	t.Run("its own re-authentication is refused", func(t *testing.T) {
		e := setupOwnershipE2E(t)
		member := e.member(t, "member", authorizer.RoleOperator)

		accepted := e.pair(t, member, "aa:bb:cc:dd:ee:01", "pk-1")
		require.NoError(t, e.svc.DeleteDevice(ctx, models.UID(accepted.UID), e.tenantID))

		_, err := e.svc.AuthDevice(ctx, e.authRequest("aa:bb:cc:dd:ee:01", "pk-1"))
		require.ErrorIs(t, err, ErrAuthUnathorized)
		assert.Equal(t, models.DeviceStatusRemoved, e.device(t, accepted.UID).Status)
	})

	t.Run("a device enrolled by tenant revives as pending", func(t *testing.T) {
		e := setupOwnershipE2E(t)

		uid := e.enrollKey(t, "aa:bb:cc:dd:ee:01", "pk-1")
		e.accept(t, uid)
		require.NoError(t, e.svc.DeleteDevice(ctx, models.UID(uid), e.tenantID))

		e.enrollKey(t, "aa:bb:cc:dd:ee:01", "pk-1")
		assert.Equal(t, models.DeviceStatusPending, e.device(t, uid).Status)
	})

	t.Run("a user provisioning key revives it as pending under that key", func(t *testing.T) {
		e := setupOwnershipE2E(t)
		member := e.member(t, "member", authorizer.RoleOperator)
		e.provisioningKey(t, digest(1), "fleet", models.ProvisioningKeyModeManual, models.ProvisioningKeyTypeUser, clearSecret)

		accepted := e.pair(t, member, "aa:bb:cc:dd:ee:01", "pk-1")
		require.NoError(t, e.svc.DeleteDevice(ctx, models.UID(accepted.UID), e.tenantID))

		req := e.authRequest("aa:bb:cc:dd:ee:01", "pk-1")
		req.ProvisioningKey = plaintextFor(1)
		_, err := e.svc.AuthDevice(ctx, req)
		require.NoError(t, err)

		device := e.device(t, accepted.UID)
		assert.Equal(t, models.DeviceStatusPending, device.Status)
		assert.Equal(t, digest(1), device.ProvisioningKeyID)
	})

	t.Run("pairing it again accepts it under the new owner", func(t *testing.T) {
		e := setupOwnershipE2E(t)
		first := e.member(t, "first", authorizer.RoleOperator)
		second := e.member(t, "second", authorizer.RoleOperator)

		accepted := e.pair(t, first, "aa:bb:cc:dd:ee:01", "pk-1")
		require.NoError(t, e.svc.DeleteDevice(ctx, models.UID(accepted.UID), e.tenantID))

		again := e.pair(t, second, "aa:bb:cc:dd:ee:01", "pk-1")

		assert.Equal(t, accepted.UID, again.UID)
		device := e.device(t, again.UID)
		assert.Equal(t, models.DeviceStatusAccepted, device.Status)
		assert.Equal(t, second, device.OwnerID)
	})
}

type departureE2E struct {
	*ownershipE2E
	cache    *memoryCache
	removals *[]removedDevice
}

func setupDepartureE2E(t *testing.T) *departureE2E {
	t.Helper()

	e := setupOwnershipE2E(t)
	cache := newMemoryCache()
	e.svc = NewService(e.st, privateKey, publicKey, cache)

	return &departureE2E{ownershipE2E: e, cache: cache, removals: recordDeviceRemovals(t)}
}

func (e *departureE2E) cached(t *testing.T, key string) bool {
	t.Helper()

	var value any

	return e.cache.Get(context.Background(), key, &value) == nil
}

func (e *departureE2E) seedCache(t *testing.T, keys ...string) {
	t.Helper()

	for _, key := range keys {
		require.NoError(t, e.cache.Set(context.Background(), key, "x", 0))
	}
}

func (e *departureE2E) isMember(t *testing.T, userID string) (authorizer.Role, bool) {
	t.Helper()

	ns, err := e.st.NamespaceResolve(context.Background(), store.NamespaceTenantIDResolver, e.tenantID)
	require.NoError(t, err)

	member, ok := ns.FindMember(userID)
	if !ok {
		return authorizer.RoleInvalid, false
	}

	return member.Role, true
}

func TestRemovingAMemberRemovesTheirDevices(t *testing.T) {
	ctx := context.Background()
	e := setupDepartureE2E(t)
	member := e.member(t, "member", authorizer.RoleOperator)

	owned := e.pair(t, member, "aa:bb:cc:dd:ee:01", "pk-1")
	team := e.enrollKey(t, "aa:bb:cc:dd:ee:02", "pk-2")
	e.accept(t, team)

	e.seedCache(t, "auth_device/"+owned.UID, "token_"+e.tenantID+member, "token_"+e.tenantID+e.ownerID)

	_, err := e.svc.RemoveNamespaceMember(ctx, &requests.NamespaceRemoveMember{UserID: e.ownerID, TenantID: e.tenantID, MemberID: member})
	require.NoError(t, err)

	device := e.device(t, owned.UID)
	assert.Equal(t, models.DeviceStatusRemoved, device.Status)
	assert.Empty(t, device.OwnerID)
	assert.Equal(t, models.DeviceStatusAccepted, e.device(t, team).Status)

	assert.Equal(t, []removedDevice{{tenantID: e.tenantID, uid: owned.UID}}, *e.removals)
	assert.False(t, e.cached(t, "auth_device/"+owned.UID))
	assert.False(t, e.cached(t, "token_"+e.tenantID+member), "the removed member's token is uncached")
	assert.True(t, e.cached(t, "token_"+e.tenantID+e.ownerID), "the actor's token is not")

	_, stillMember := e.isMember(t, member)
	assert.False(t, stillMember)
}

func TestLeavingANamespaceTakesTheMembersDevices(t *testing.T) {
	ctx := context.Background()
	e := setupDepartureE2E(t)
	member := e.member(t, "member", authorizer.RoleOperator)

	owned := e.pair(t, member, "aa:bb:cc:dd:ee:01", "pk-1")

	_, err := e.svc.LeaveNamespace(ctx, &requests.LeaveNamespace{UserID: member, TenantID: e.tenantID, AuthenticatedTenantID: "99999999-9999-4999-9999-999999999999"})
	require.NoError(t, err)

	assert.Equal(t, models.DeviceStatusRemoved, e.device(t, owned.UID).Status)
	assert.Equal(t, []removedDevice{{tenantID: e.tenantID, uid: owned.UID}}, *e.removals)
}

func TestDemotingAMemberWhoCannotAcceptDevices(t *testing.T) {
	ctx := context.Background()

	t.Run("demoting to observer removes their devices and keeps the membership", func(t *testing.T) {
		e := setupDepartureE2E(t)
		member := e.member(t, "member", authorizer.RoleOperator)
		owned := e.pair(t, member, "aa:bb:cc:dd:ee:01", "pk-1")

		require.NoError(t, e.svc.UpdateNamespaceMember(ctx, &requests.NamespaceUpdateMember{
			UserID: e.ownerID, TenantID: e.tenantID, MemberID: member, MemberRole: authorizer.RoleObserver,
		}))

		assert.Equal(t, models.DeviceStatusRemoved, e.device(t, owned.UID).Status)
		assert.Equal(t, []removedDevice{{tenantID: e.tenantID, uid: owned.UID}}, *e.removals)

		role, stillMember := e.isMember(t, member)
		assert.True(t, stillMember)
		assert.Equal(t, authorizer.RoleObserver, role)
	})

	t.Run("changing to a role that accepts devices keeps them", func(t *testing.T) {
		e := setupDepartureE2E(t)
		member := e.member(t, "member", authorizer.RoleOperator)
		owned := e.pair(t, member, "aa:bb:cc:dd:ee:01", "pk-1")

		require.NoError(t, e.svc.UpdateNamespaceMember(ctx, &requests.NamespaceUpdateMember{
			UserID: e.ownerID, TenantID: e.tenantID, MemberID: member, MemberRole: authorizer.RoleAdministrator,
		}))

		device := e.device(t, owned.UID)
		assert.Equal(t, models.DeviceStatusAccepted, device.Status)
		assert.Equal(t, member, device.OwnerID)
		assert.Empty(t, *e.removals)
	})
}

func TestKeepingADepartingMembersDeviceAsATeamDevice(t *testing.T) {
	ctx := context.Background()

	t.Run("a kept device stays as a team device", func(t *testing.T) {
		e := setupDepartureE2E(t)
		member := e.member(t, "member", authorizer.RoleOperator)
		kept := e.pair(t, member, "aa:bb:cc:dd:ee:01", "pk-1")
		gone := e.pair(t, member, "aa:bb:cc:dd:ee:02", "pk-2")

		_, err := e.svc.RemoveNamespaceMember(ctx, &requests.NamespaceRemoveMember{
			UserID: e.ownerID, TenantID: e.tenantID, MemberID: member, KeepDevices: []string{kept.UID},
		})
		require.NoError(t, err)

		device := e.device(t, kept.UID)
		assert.Equal(t, models.DeviceStatusAccepted, device.Status)
		assert.Empty(t, device.OwnerID)
		assert.Equal(t, []removedDevice{{tenantID: e.tenantID, uid: gone.UID}}, *e.removals)
	})

	t.Run("demoting keeps the listed devices too", func(t *testing.T) {
		e := setupDepartureE2E(t)
		member := e.member(t, "member", authorizer.RoleOperator)
		kept := e.pair(t, member, "aa:bb:cc:dd:ee:01", "pk-1")

		require.NoError(t, e.svc.UpdateNamespaceMember(ctx, &requests.NamespaceUpdateMember{
			UserID: e.ownerID, TenantID: e.tenantID, MemberID: member, MemberRole: authorizer.RoleObserver,
			KeepDevices: []string{kept.UID},
		}))

		assert.Equal(t, models.DeviceStatusAccepted, e.device(t, kept.UID).Status)
		assert.Empty(t, *e.removals)
	})

	t.Run("a device the member does not own is refused and nothing changes", func(t *testing.T) {
		e := setupDepartureE2E(t)
		member := e.member(t, "member", authorizer.RoleOperator)
		owned := e.pair(t, member, "aa:bb:cc:dd:ee:01", "pk-1")
		team := e.enrollKey(t, "aa:bb:cc:dd:ee:02", "pk-2")
		e.accept(t, team)

		_, err := e.svc.RemoveNamespaceMember(ctx, &requests.NamespaceRemoveMember{
			UserID: e.ownerID, TenantID: e.tenantID, MemberID: member, KeepDevices: []string{team},
		})
		require.ErrorIs(t, err, ErrDeviceNotOwned)

		assert.Equal(t, member, e.device(t, owned.UID).OwnerID)
		_, stillMember := e.isMember(t, member)
		assert.True(t, stillMember)
	})

	t.Run("a caller who cannot create provisioning keys cannot keep devices", func(t *testing.T) {
		e := setupDepartureE2E(t)
		operator := e.member(t, "operator", authorizer.RoleOperator)
		observer := e.member(t, "observer", authorizer.RoleObserver)

		_, err := e.svc.RemoveNamespaceMember(ctx, &requests.NamespaceRemoveMember{
			UserID: operator, TenantID: e.tenantID, MemberID: observer, KeepDevices: []string{"uid"},
		})
		require.ErrorIs(t, err, ErrRoleForbidden)

		_, stillMember := e.isMember(t, observer)
		assert.True(t, stillMember)
	})
}

func TestMakingAPairedDeviceATeamDevice(t *testing.T) {
	ctx := context.Background()

	t.Run("clears the owner and leaves the device accepted", func(t *testing.T) {
		e := setupDepartureE2E(t)
		member := e.member(t, "member", authorizer.RoleOperator)
		owned := e.pair(t, member, "aa:bb:cc:dd:ee:01", "pk-1")
		source := e.device(t, owned.UID).ProvisioningKeyID

		require.NoError(t, e.svc.MakeTeamDevice(ctx, e.tenantID, owned.UID))

		device := e.device(t, owned.UID)
		assert.Equal(t, models.DeviceStatusAccepted, device.Status)
		assert.Empty(t, device.OwnerID)
		assert.Equal(t, source, device.ProvisioningKeyID, "the device still records that it arrived by pairing")
		assert.Empty(t, *e.removals)

		_, err := e.svc.RemoveNamespaceMember(ctx, &requests.NamespaceRemoveMember{UserID: e.ownerID, TenantID: e.tenantID, MemberID: member})
		require.NoError(t, err)
		assert.Equal(t, models.DeviceStatusAccepted, e.device(t, owned.UID).Status)
	})

	t.Run("succeeds without change on a team device", func(t *testing.T) {
		e := setupDepartureE2E(t)
		team := e.enrollKey(t, "aa:bb:cc:dd:ee:01", "pk-1")
		e.accept(t, team)

		assert.NoError(t, e.svc.MakeTeamDevice(ctx, e.tenantID, team))
	})

	t.Run("is not found for a device that is not accepted", func(t *testing.T) {
		e := setupDepartureE2E(t)
		pending := e.enrollKey(t, "aa:bb:cc:dd:ee:01", "pk-1")

		require.ErrorIs(t, e.svc.MakeTeamDevice(ctx, e.tenantID, pending), ErrDeviceNotFound)
	})
}

type departsAfterRead struct {
	store.Store
	uid    string
	depart func()
}

func (s *departsAfterRead) DeviceResolve(ctx context.Context, sc scope.Scope, resolver store.DeviceResolver, value string, opts ...store.QueryOption) (*models.Device, error) {
	device, err := s.Store.DeviceResolve(ctx, sc, resolver, value, opts...)
	if resolver == store.DeviceUIDResolver && value == s.uid && s.depart != nil {
		depart := s.depart
		s.depart = nil
		depart()
	}

	return device, err
}

func TestADepartureDuringAuthenticationIsNotUndone(t *testing.T) {
	ctx := context.Background()

	for _, departure := range []store.MemberDeparture{{}, {KeepMembership: true}} {
		t.Run(fmt.Sprintf("keep membership %t", departure.KeepMembership), func(t *testing.T) {
			e := setupOwnershipE2E(t)
			member := e.member(t, "member", authorizer.RoleOperator)
			owned := e.pair(t, member, "aa:bb:cc:dd:ee:01", "pk-1")

			racing := &departsAfterRead{Store: e.st, uid: owned.UID, depart: func() {
				_, err := e.st.NamespaceDepartMember(ctx, scope.MustBounded(e.tenantID), member, departure)
				require.NoError(t, err)
			}}
			svc := NewService(racing, privateKey, publicKey, newMemoryCache())

			_, err := svc.AuthDevice(ctx, e.authRequest("aa:bb:cc:dd:ee:01", "pk-1"))
			require.Error(t, err, "the stale read must not be written back")

			device := e.device(t, owned.UID)
			assert.Equal(t, models.DeviceStatusRemoved, device.Status)
			assert.Empty(t, device.OwnerID)
		})
	}
}
