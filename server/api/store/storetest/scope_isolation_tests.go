package storetest

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The isolation suite is the behavioural signal behind making namespace scope explicit: every
// converted operation seeds a row in one namespace, reads it bounded to a second, and expects
// nothing back. A converted query that stops applying its scope fails here rather than shipping.
//
// Each subtest also asserts the positive case against the owning namespace, so a query that returns
// nothing for an unrelated reason (a broken predicate, a missing fixture) cannot pass by accident.

func (s *Suite) TestScopeIsolationDeviceResolve(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)
	uid := s.CreateDevice(t, WithTenantID(owner), WithDeviceName("dev"))

	got, err := st.DeviceResolve(ctx, scope.MustBounded(owner), store.DeviceUIDResolver, string(uid))
	require.NoError(t, err)
	assert.Equal(t, string(uid), got.UID)

	got, err = st.DeviceResolve(ctx, scope.MustBounded(other), store.DeviceUIDResolver, string(uid))
	require.ErrorIs(t, err, store.ErrNoDocuments)
	assert.Nil(t, got)
}

func (s *Suite) TestScopeIsolationDeviceList(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)
	s.CreateDevice(t, WithTenantID(owner), WithDeviceName("dev"))

	devices, count, err := st.DeviceList(ctx, scope.MustBounded(owner), store.DeviceAcceptableAsFalse)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Len(t, devices, 1)

	devices, count, err = st.DeviceList(ctx, scope.MustBounded(other), store.DeviceAcceptableAsFalse)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Empty(t, devices)
}

func (s *Suite) TestScopeIsolationSessionResolve(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)
	deviceUID := s.CreateDevice(t, WithTenantID(owner), WithDeviceName("dev"))
	sessionUID := s.CreateSession(t, WithSessionDevice(deviceUID))

	got, err := st.SessionResolve(ctx, scope.MustBounded(owner), store.SessionUIDResolver, string(sessionUID))
	require.NoError(t, err)
	assert.Equal(t, string(sessionUID), got.UID)

	got, err = st.SessionResolve(ctx, scope.MustBounded(other), store.SessionUIDResolver, string(sessionUID))
	require.ErrorIs(t, err, store.ErrNoDocuments)
	assert.Nil(t, got)
}

func (s *Suite) TestScopeIsolationTagResolve(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)
	s.CreateTag(t, WithTagName("production"), WithTagTenant(owner))

	got, err := st.TagResolve(ctx, scope.MustBounded(owner), store.TagNameResolver, "production")
	require.NoError(t, err)
	assert.Equal(t, "production", got.Name)

	got, err = st.TagResolve(ctx, scope.MustBounded(other), store.TagNameResolver, "production")
	require.ErrorIs(t, err, store.ErrNoDocuments)
	assert.Nil(t, got)
}

func (s *Suite) TestScopeIsolationTagConflicts(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)
	s.CreateTag(t, WithTagName("production"), WithTagTenant(owner))

	conflicts, has, err := st.TagConflicts(ctx, scope.MustBounded(owner), &models.TagConflicts{Name: "production"})
	require.NoError(t, err)
	assert.True(t, has)
	assert.Equal(t, []string{"name"}, conflicts)

	conflicts, has, err = st.TagConflicts(ctx, scope.MustBounded(other), &models.TagConflicts{Name: "production"})
	require.NoError(t, err)
	assert.False(t, has)
	assert.Empty(t, conflicts)
}

func (s *Suite) TestScopeIsolationInstallKeyResolve(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)

	// The same digest in two namespaces: the install key's primary key is namespace-scoped, so the
	// digest alone does not identify a key. This is the collision an unbounded resolve would admit.
	const digest = "3333333333333333333333333333333333333333333333333333333333333333"
	for _, tenant := range []string{owner, other} {
		_, err := st.InstallKeyCreate(ctx, &models.InstallKey{
			ID:        digest,
			Name:      "shared-digest",
			TenantID:  tenant,
			Mode:      models.InstallKeyModeManual,
			Reusable:  true,
			Tags:      []string{},
			CreatedBy: "00000000-0000-4000-0000-000000000009",
		})
		require.NoError(t, err)
	}

	got, err := st.InstallKeyResolve(ctx, scope.MustBounded(owner), store.InstallKeyIDResolver, digest)
	require.NoError(t, err)
	assert.Equal(t, owner, got.TenantID)

	got, err = st.InstallKeyResolve(ctx, scope.MustBounded(other), store.InstallKeyIDResolver, digest)
	require.NoError(t, err)
	assert.Equal(t, other, got.TenantID)
}

func (s *Suite) TestScopeIsolationInstallKeyList(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)

	_, err := st.InstallKeyCreate(ctx, &models.InstallKey{
		ID:        "4444444444444444444444444444444444444444444444444444444444444444",
		Name:      "only-key",
		TenantID:  owner,
		Mode:      models.InstallKeyModeManual,
		Reusable:  true,
		Tags:      []string{},
		CreatedBy: "00000000-0000-4000-0000-000000000009",
	})
	require.NoError(t, err)

	// Every namespace is provisioned with its own system keys, so the assertion is not a count but
	// that nothing from another namespace is ever returned.
	keys, _, err := st.InstallKeyList(ctx, scope.MustBounded(owner))
	require.NoError(t, err)
	assert.Contains(t, installKeyNames(keys), "only-key")
	for _, k := range keys {
		assert.Equal(t, owner, k.TenantID)
	}

	keys, _, err = st.InstallKeyList(ctx, scope.MustBounded(other))
	require.NoError(t, err)
	assert.NotContains(t, installKeyNames(keys), "only-key")
	for _, k := range keys {
		assert.Equal(t, other, k.TenantID)
	}
}

func installKeyNames(keys []models.InstallKey) []string {
	names := make([]string, len(keys))
	for i, k := range keys {
		names[i] = k.Name
	}

	return names
}

func (s *Suite) TestScopeIsolationInstallKeyConflicts(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)

	_, err := st.InstallKeyCreate(ctx, &models.InstallKey{
		ID:        "5555555555555555555555555555555555555555555555555555555555555555",
		Name:      "taken",
		TenantID:  owner,
		Mode:      models.InstallKeyModeManual,
		Reusable:  true,
		Tags:      []string{},
		CreatedBy: "00000000-0000-4000-0000-000000000009",
	})
	require.NoError(t, err)

	_, has, err := st.InstallKeyConflicts(ctx, scope.MustBounded(owner), &models.InstallKeyConflicts{Name: "taken"})
	require.NoError(t, err)
	assert.True(t, has)

	_, has, err = st.InstallKeyConflicts(ctx, scope.MustBounded(other), &models.InstallKeyConflicts{Name: "taken"})
	require.NoError(t, err)
	assert.False(t, has)
}

func (s *Suite) TestScopeIsolationInstallKeyResolveSystem(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)

	// CreateNamespace provisions each namespace's system keys, so both exist; the assertion is that
	// each scope resolves its own.
	ownerKey, err := st.InstallKeyResolveSystem(ctx, scope.MustBounded(owner))
	require.NoError(t, err)
	assert.Equal(t, owner, ownerKey.TenantID)

	otherKey, err := st.InstallKeyResolveSystem(ctx, scope.MustBounded(other))
	require.NoError(t, err)
	assert.Equal(t, other, otherKey.TenantID)
	assert.NotEqual(t, ownerKey.TenantID, otherKey.TenantID)
}

func (s *Suite) TestScopeIsolationInstallKeyEventList(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)

	const digest = "6666666666666666666666666666666666666666666666666666666666666666"
	_, err := st.InstallKeyCreate(ctx, &models.InstallKey{
		ID:        digest,
		Name:      "with-history",
		TenantID:  owner,
		Mode:      models.InstallKeyModeManual,
		Reusable:  true,
		Tags:      []string{},
		CreatedBy: "00000000-0000-4000-0000-000000000009",
	})
	require.NoError(t, err)

	deviceUID := s.CreateDevice(t, WithTenantID(owner), WithDeviceName("enrolled"))
	require.NoError(t, st.InstallKeyEventCreate(ctx, &models.InstallKeyEvent{
		InstallKeyID: digest,
		TenantID:     owner,
		DeviceUID:    string(deviceUID),
		Hostname:     "enrolled",
	}))

	events, count, err := st.InstallKeyEventList(ctx, scope.MustBounded(owner), digest)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Len(t, events, 1)

	events, count, err = st.InstallKeyEventList(ctx, scope.MustBounded(other), digest)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Empty(t, events)
}

func (s *Suite) TestScopeIsolationAPIKeyList(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)
	s.CreateAPIKey(t, WithAPIKeyName("dev"), WithAPIKeyTenant(owner))

	keys, count, err := st.APIKeyList(ctx, scope.MustBounded(owner))
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Len(t, keys, 1)

	keys, count, err = st.APIKeyList(ctx, scope.MustBounded(other))
	require.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Empty(t, keys)
}

func (s *Suite) TestScopeIsolationPublicKeyList(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)
	s.CreatePublicKey(t, WithPublicKeyName("key"), WithPublicKeyTenant(owner))

	keys, count, err := st.PublicKeyList(ctx, scope.MustBounded(owner))
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Len(t, keys, 1)

	keys, count, err = st.PublicKeyList(ctx, scope.MustBounded(other))
	require.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Empty(t, keys)
}

func (s *Suite) TestScopeIsolationNamespaceGetMembers(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)
	memberID := s.CreateUser(t)

	require.NoError(t, st.NamespaceCreateMembership(ctx, scope.MustBounded(owner),
		&models.Member{ID: memberID, Role: authorizer.RoleObserver}))

	members, count, err := st.NamespaceGetMembers(ctx, scope.MustBounded(owner))
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	require.Len(t, members, 1)
	assert.Equal(t, memberID, members[0].ID)

	members, count, err = st.NamespaceGetMembers(ctx, scope.MustBounded(other))
	require.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Empty(t, members)
}

func (s *Suite) TestScopeIsolationMembershipInvitationResolve(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)
	userID := s.CreateUser(t)
	inviterID := s.CreateUser(t)

	require.NoError(t, st.MembershipInvitationCreate(ctx, &models.MembershipInvitation{
		UserID:    userID,
		TenantID:  owner,
		InvitedBy: inviterID,
		Status:    models.MembershipInvitationStatusPending,
		Role:      authorizer.RoleObserver,
	}))

	got, err := st.MembershipInvitationResolve(ctx, scope.MustBounded(owner), userID)
	require.NoError(t, err)
	assert.Equal(t, owner, got.TenantID)

	got, err = st.MembershipInvitationResolve(ctx, scope.MustBounded(other), userID)
	require.ErrorIs(t, err, store.ErrNoDocuments)
	assert.Nil(t, got)
}

func (s *Suite) TestScopeIsolationNamespaceMembershipInvitationList(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)
	userID := s.CreateUser(t)
	inviterID := s.CreateUser(t)

	require.NoError(t, st.MembershipInvitationCreate(ctx, &models.MembershipInvitation{
		UserID:    userID,
		TenantID:  owner,
		InvitedBy: inviterID,
		Status:    models.MembershipInvitationStatusPending,
		Role:      authorizer.RoleObserver,
	}))

	invitations, count, err := st.NamespaceMembershipInvitationList(ctx, scope.MustBounded(owner))
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
	assert.Len(t, invitations, 1)

	invitations, count, err = st.NamespaceMembershipInvitationList(ctx, scope.MustBounded(other))
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
	assert.Empty(t, invitations)
}

func (s *Suite) TestScopeIsolationGetStats(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)
	s.CreateDevice(t, WithTenantID(owner), WithDeviceName("dev"), WithDeviceStatus(models.DeviceStatusAccepted))

	ownerStats, err := st.GetStats(ctx, scope.MustBounded(owner))
	require.NoError(t, err)
	assert.Equal(t, 1, ownerStats.RegisteredDevices)

	otherStats, err := st.GetStats(ctx, scope.MustBounded(other))
	require.NoError(t, err)
	assert.Equal(t, 0, otherStats.RegisteredDevices)

	// The unbounded scope is what the instance-wide statistics ask for, and it does span both.
	allStats, err := st.GetStats(ctx, scope.NewUnbounded("test: instance-wide statistics span every namespace"))
	require.NoError(t, err)
	assert.Equal(t, 1, allStats.RegisteredDevices)
}

// TestScopeIsolationCountRegisteredDevices pins that the narrow count honours the scope it is
// given, so a bounded caller cannot read another namespace's devices through it.
func (s *Suite) TestScopeIsolationCountRegisteredDevices(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)
	s.CreateDevice(t, WithTenantID(owner), WithDeviceName("dev"), WithDeviceStatus(models.DeviceStatusAccepted))

	ownerCount, err := st.CountRegisteredDevices(ctx, scope.MustBounded(owner))
	require.NoError(t, err)
	assert.Equal(t, 1, ownerCount)

	otherCount, err := st.CountRegisteredDevices(ctx, scope.MustBounded(other))
	require.NoError(t, err)
	assert.Equal(t, 0, otherCount)

	// The unbounded scope is what the license evaluation asks for, and it does span both.
	allCount, err := st.CountRegisteredDevices(ctx, scope.NewUnbounded("test: instance-wide device count spans every namespace"))
	require.NoError(t, err)
	assert.Equal(t, 1, allCount)
}

// TestScopeRejectsUnconstructedScope pins the zero value: a [scope.Scope] that was never built is
// neither bounded nor deliberately unbounded, so the store refuses it instead of reading everything.
func (s *Suite) TestScopeRejectsUnconstructedScope(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	tenantID := s.CreateNamespace(t)
	uid := s.CreateDevice(t, WithTenantID(tenantID), WithDeviceName("dev"))

	got, err := st.DeviceResolve(ctx, scope.Scope{}, store.DeviceUIDResolver, string(uid))
	require.ErrorIs(t, err, store.ErrInvalidScope)
	assert.Nil(t, got)

	devices, _, err := st.DeviceList(ctx, scope.Scope{}, store.DeviceAcceptableAsFalse)
	require.ErrorIs(t, err, store.ErrInvalidScope)
	assert.Empty(t, devices)
}

func (s *Suite) TestScopeIsolationAPIKeyConflicts(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)
	s.CreateAPIKey(t, WithAPIKeyName("dev"), WithAPIKeyTenant(owner))

	_, has, err := st.APIKeyConflicts(ctx, scope.MustBounded(owner), &models.APIKeyConflicts{Name: "dev"})
	require.NoError(t, err)
	assert.True(t, has)

	_, has, err = st.APIKeyConflicts(ctx, scope.MustBounded(other), &models.APIKeyConflicts{Name: "dev"})
	require.NoError(t, err)
	assert.False(t, has)
}

func (s *Suite) TestScopeIsolationInstallKeyResolveSystemPairing(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)

	ownerKey, err := st.InstallKeyResolveSystemPairing(ctx, scope.MustBounded(owner))
	require.NoError(t, err)
	assert.Equal(t, owner, ownerKey.TenantID)

	otherKey, err := st.InstallKeyResolveSystemPairing(ctx, scope.MustBounded(other))
	require.NoError(t, err)
	assert.Equal(t, other, otherKey.TenantID)

	// An unbounded scope cannot name "the namespace's" system key, so it is refused outright rather
	// than returning whichever namespace's key happened to sort first.
	_, err = st.InstallKeyResolveSystemPairing(ctx, scope.NewUnbounded(reasonTestQueryMechanics))
	assert.ErrorIs(t, err, store.ErrInvalidScope)
}

func (s *Suite) TestScopeIsolationInstallKeyEventStampDecision(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)

	const digest = "7777777777777777777777777777777777777777777777777777777777777777"
	_, err := st.InstallKeyCreate(ctx, &models.InstallKey{
		ID:        digest,
		Name:      "stamped",
		TenantID:  owner,
		Mode:      models.InstallKeyModeManual,
		Reusable:  true,
		Tags:      []string{},
		CreatedBy: "00000000-0000-4000-0000-000000000009",
	})
	require.NoError(t, err)

	deviceUID := s.CreateDevice(t, WithTenantID(owner), WithDeviceName("enrolled"))
	require.NoError(t, st.InstallKeyEventCreate(ctx, &models.InstallKeyEvent{
		InstallKeyID: digest,
		TenantID:     owner,
		DeviceUID:    string(deviceUID),
		Hostname:     "enrolled",
	}))

	// Stamping from another namespace must not touch the event: this is an UPDATE, so an unbounded
	// or wrongly-bounded scope would silently write across namespaces.
	require.NoError(t, st.InstallKeyEventStampDecision(ctx, scope.MustBounded(other), string(deviceUID), models.DeviceStatusRejected, clock.Now()))

	events, _, err := st.InstallKeyEventList(ctx, scope.MustBounded(owner), digest)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Empty(t, events[0].DecidedStatus)

	require.NoError(t, st.InstallKeyEventStampDecision(ctx, scope.MustBounded(owner), string(deviceUID), models.DeviceStatusAccepted, clock.Now()))

	events, _, err = st.InstallKeyEventList(ctx, scope.MustBounded(owner), digest)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, models.DeviceStatusAccepted, events[0].DecidedStatus)

	// An unbounded scope is refused rather than stamping every namespace's newest event.
	assert.ErrorIs(t,
		st.InstallKeyEventStampDecision(ctx, scope.NewUnbounded(reasonTestQueryMechanics), string(deviceUID), models.DeviceStatusRejected, clock.Now()),
		store.ErrInvalidScope)
}

func (s *Suite) TestScopeIsolationMembershipWrites(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)
	memberID := s.CreateUser(t)

	require.NoError(t, st.NamespaceCreateMembership(ctx, scope.MustBounded(owner),
		&models.Member{ID: memberID, Role: authorizer.RoleObserver}))

	members, _, err := st.NamespaceGetMembers(ctx, scope.MustBounded(other))
	require.NoError(t, err)
	assert.Empty(t, members)

	// Updating and deleting it from another namespace finds nothing to act on.
	require.ErrorIs(t,
		st.NamespaceUpdateMembership(ctx, scope.MustBounded(other), &models.Member{ID: memberID, Role: authorizer.RoleAdministrator}),
		store.ErrNoDocuments)
	require.ErrorIs(t,
		st.NamespaceDeleteMembership(ctx, scope.MustBounded(other), &models.Member{ID: memberID}),
		store.ErrNoDocuments)

	// The membership survives both, and its own namespace can still act on it.
	require.NoError(t, st.NamespaceUpdateMembership(ctx, scope.MustBounded(owner),
		&models.Member{ID: memberID, Role: authorizer.RoleAdministrator}))
	require.NoError(t, st.NamespaceDeleteMembership(ctx, scope.MustBounded(owner), &models.Member{ID: memberID}))

	// Membership writes cannot be expressed unbounded at all.
	assert.ErrorIs(t,
		st.NamespaceCreateMembership(ctx, scope.NewUnbounded(reasonTestQueryMechanics), &models.Member{ID: memberID, Role: authorizer.RoleObserver}),
		store.ErrInvalidScope)
}

func (s *Suite) TestScopeIsolationNamespaceIncrementDeviceCount(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)

	require.NoError(t, st.NamespaceIncrementDeviceCount(ctx, scope.MustBounded(owner), models.DeviceStatusAccepted, 3))

	ownerNs, err := st.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, owner)
	require.NoError(t, err)
	assert.Equal(t, int64(3), ownerNs.DevicesAcceptedCount)

	otherNs, err := st.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, other)
	require.NoError(t, err)
	assert.Equal(t, int64(0), otherNs.DevicesAcceptedCount)

	// Counting devices across every namespace at once is not expressible.
	assert.ErrorIs(t,
		st.NamespaceIncrementDeviceCount(ctx, scope.NewUnbounded(reasonTestQueryMechanics), models.DeviceStatusAccepted, 1),
		store.ErrInvalidScope)
}

func (s *Suite) TestScopeIsolationDeviceConflicts(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)
	s.CreateDevice(t, WithTenantID(owner), WithDeviceName("shared-name"), WithDeviceStatus(models.DeviceStatusAccepted))

	_, has, err := st.DeviceConflicts(ctx, scope.MustBounded(owner), &models.DeviceConflicts{Name: "shared-name"})
	require.NoError(t, err)
	assert.True(t, has)

	_, has, err = st.DeviceConflicts(ctx, scope.MustBounded(other), &models.DeviceConflicts{Name: "shared-name"})
	require.NoError(t, err)
	assert.False(t, has)
}

func (s *Suite) TestScopeIsolationTagList(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)
	s.CreateTag(t, WithTagName("production"), WithTagTenant(owner))

	tags, count, err := st.TagList(ctx, scope.MustBounded(owner))
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Len(t, tags, 1)

	tags, count, err = st.TagList(ctx, scope.MustBounded(other))
	require.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Empty(t, tags)
}

func (s *Suite) TestScopeIsolationSessionList(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)
	deviceUID := s.CreateDevice(t, WithTenantID(owner), WithDeviceName("dev"))
	s.CreateSession(t, WithSessionDevice(deviceUID))

	sessions, count, err := st.SessionList(ctx, scope.MustBounded(owner))
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Len(t, sessions, 1)

	sessions, count, err = st.SessionList(ctx, scope.MustBounded(other))
	require.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Empty(t, sessions)
}

func (s *Suite) TestScopeIsolationAPIKeyResolve(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)
	keyID := s.CreateAPIKey(t, WithAPIKeyName("dev"), WithAPIKeyTenant(owner))

	got, err := st.APIKeyResolve(ctx, scope.MustBounded(owner), store.APIKeyIDResolver, keyID)
	require.NoError(t, err)
	assert.Equal(t, keyID, got.ID)

	got, err = st.APIKeyResolve(ctx, scope.MustBounded(other), store.APIKeyIDResolver, keyID)
	require.ErrorIs(t, err, store.ErrNoDocuments)
	assert.Nil(t, got)
}

func (s *Suite) TestScopeIsolationPublicKeyResolve(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)
	fingerprint := s.CreatePublicKey(t, WithPublicKeyName("key"), WithPublicKeyTenant(owner))

	got, err := st.PublicKeyResolve(ctx, scope.MustBounded(owner), store.PublicKeyFingerprintResolver, fingerprint)
	require.NoError(t, err)
	assert.Equal(t, fingerprint, got.Fingerprint)

	got, err = st.PublicKeyResolve(ctx, scope.MustBounded(other), store.PublicKeyFingerprintResolver, fingerprint)
	require.ErrorIs(t, err, store.ErrNoDocuments)
	assert.Nil(t, got)
}

func (s *Suite) TestScopeIsolationAccessPolicyList(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)

	_, err := st.AccessPolicyCreate(ctx, &models.AccessPolicy{
		TenantID: owner,
		Name:     "test-policy",
		Subject:  models.PolicySubject{Type: models.PolicySubjectAllMembers},
		Logins:   []string{"*"},
		SourceIP: []string{},
		Action:   models.PolicyActionAllow,
	})
	require.NoError(t, err)

	policies, _, err := st.AccessPolicyList(ctx, scope.MustBounded(owner))
	require.NoError(t, err)
	assert.True(t, accessPolicyContains(policies, "test-policy"))

	policies, _, err = st.AccessPolicyList(ctx, scope.MustBounded(other))
	require.NoError(t, err)
	assert.False(t, accessPolicyContains(policies, "test-policy"))
}

func accessPolicyContains(policies []models.AccessPolicy, name string) bool {
	for _, p := range policies {
		if p.Name == name {
			return true
		}
	}

	return false
}

func (s *Suite) TestScopeIsolationAccessPolicyResolve(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)

	id, err := st.AccessPolicyCreate(ctx, &models.AccessPolicy{
		TenantID: owner,
		Name:     "allow-all",
		Subject:  models.PolicySubject{Type: models.PolicySubjectAllMembers},
		Logins:   []string{"*"},
		SourceIP: []string{},
		Action:   models.PolicyActionAllow,
	})
	require.NoError(t, err)

	got, err := st.AccessPolicyResolve(ctx, scope.MustBounded(owner), store.AccessPolicyIDResolver, id)
	require.NoError(t, err)
	assert.Equal(t, id, got.ID)

	got, err = st.AccessPolicyResolve(ctx, scope.MustBounded(other), store.AccessPolicyIDResolver, id)
	require.ErrorIs(t, err, store.ErrNoDocuments)
	assert.Nil(t, got)
}

func (s *Suite) TestScopeIsolationSSHIdentityList(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)
	userID := s.CreateUser(t)

	_, err := st.SSHIdentityCreate(ctx, &models.SSHIdentity{
		TenantID:    owner,
		PrincipalID: userID,
		Fingerprint: "SHA256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa=",
		Data:        []byte("ssh-ed25519 AAAA fake-key"),
		Name:        "test-key",
		Source:      models.SSHIdentitySourceManual,
		CreatedAt:   clock.Now(),
	})
	require.NoError(t, err)

	identities, count, err := st.SSHIdentityList(ctx, scope.MustBounded(owner))
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Len(t, identities, 1)

	identities, count, err = st.SSHIdentityList(ctx, scope.MustBounded(other))
	require.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Empty(t, identities)
}

func (s *Suite) TestScopeIsolationSSHIdentityResolve(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	owner := s.CreateNamespace(t)
	other := s.CreateNamespace(t)
	userID := s.CreateUser(t)

	fingerprint := "SHA256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb="
	_, err := st.SSHIdentityCreate(ctx, &models.SSHIdentity{
		TenantID:    owner,
		PrincipalID: userID,
		Fingerprint: fingerprint,
		Data:        []byte("ssh-ed25519 BBBB fake-key"),
		Name:        "test-key",
		Source:      models.SSHIdentitySourceManual,
		CreatedAt:   clock.Now(),
	})
	require.NoError(t, err)

	got, err := st.SSHIdentityResolve(ctx, scope.MustBounded(owner), store.SSHIdentityFingerprintResolver, fingerprint)
	require.NoError(t, err)
	assert.Equal(t, fingerprint, got.Fingerprint)

	got, err = st.SSHIdentityResolve(ctx, scope.MustBounded(other), store.SSHIdentityFingerprintResolver, fingerprint)
	require.ErrorIs(t, err, store.ErrNoDocuments)
	assert.Nil(t, got)
}
