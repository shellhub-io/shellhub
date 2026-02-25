package storetest

import (
	"testing"
)

// Suite runs all generic store tests against a provider
type Suite struct {
	provider StoreProvider
}

// NewSuite creates a new test suite with the given provider
func NewSuite(provider StoreProvider) *Suite {
	return &Suite{provider: provider}
}

// Run executes all store tests organized by interface
func (s *Suite) Run(t *testing.T) {
	t.Run("NamespaceStore", func(t *testing.T) {
		s.TestNamespaceList(t)
		s.TestNamespaceResolve(t)
		s.TestNamespaceGetPreferred(t)
		s.TestNamespaceCreate(t)
		s.TestNamespaceConflicts(t)
		s.TestNamespaceUpdate(t)
		s.TestNamespaceIncrementDeviceCount(t)
		s.TestNamespaceSyncDeviceCounts(t)
		s.TestNamespaceDelete(t)
		s.TestNamespaceDeleteMany(t)
	})

	t.Run("DeviceStore", func(t *testing.T) {
		s.TestDeviceList(t)
		s.TestDeviceResolve(t)
		s.TestDeviceCreate(t)
		s.TestDeviceConflicts(t)
		s.TestDeviceUpdate(t)
		s.TestDeviceHeartbeat(t)
		s.TestDeviceDelete(t)
		s.TestDeviceDeleteMany(t)
		s.TestDeviceStatusUpdatedAt(t)
	})

	t.Run("UserStore", func(t *testing.T) {
		s.TestUserList(t)
		s.TestUserResolve(t)
		s.TestUserCreate(t)
		s.TestUserConflicts(t)
		s.TestUserUpdate(t)
		s.TestUserDelete(t)
		s.TestUserGetInfo(t)
	})

	t.Run("SessionStore", func(t *testing.T) {
		s.TestSessionList(t)
		s.TestSessionResolve(t)
		s.TestSessionCreate(t)
		s.TestSessionUpdateDeviceUID(t)
		s.TestSessionUpdate(t)
		s.TestActiveSessionDelete(t)
		s.TestActiveSessionResolve(t)
		s.TestActiveSessionUpdate(t)
		s.TestSessionEventsCreate(t)
		s.TestSessionEventsList(t)
		s.TestSessionEventsDelete(t)
	})

	t.Run("TagStore", func(t *testing.T) {
		s.TestTagCreate(t)
		s.TestTagConflicts(t)
		s.TestTagList(t)
		s.TestTagResolve(t)
		s.TestTagUpdate(t)
		s.TestTagPushToTarget(t)
		s.TestTagPullFromTarget(t)
		s.TestTagDelete(t)
	})

	t.Run("APIKeyStore", func(t *testing.T) {
		s.TestAPIKeyCreate(t)
		s.TestAPIKeyConflicts(t)
		s.TestAPIKeyResolve(t)
		s.TestAPIKeyList(t)
		s.TestAPIKeyUpdate(t)
		s.TestAPIKeyDelete(t)
	})

	t.Run("PublicKeyStore", func(t *testing.T) {
		s.TestPublicKeyResolve(t)
		s.TestPublicKeyList(t)
		s.TestPublicKeyCreate(t)
		s.TestPublicKeyUpdate(t)
		s.TestPublicKeyDelete(t)
	})

	t.Run("StatsStore", func(t *testing.T) {
		s.TestGetStats(t)
	})

	t.Run("UserInvitationsStore", func(t *testing.T) {
		s.TestUserInvitationsUpsert(t)
	})

	t.Run("PrivateKeyStore", func(t *testing.T) {
		s.TestPrivateKeyCreate(t)
		s.TestPrivateKeyGet(t)
	})

	t.Run("MemberStore", func(t *testing.T) {
		s.TestNamespaceCreateMembership(t)
		s.TestNamespaceUpdateMembership(t)
		s.TestNamespaceDeleteMembership(t)
	})

	t.Run("MembershipInvitationStore", func(t *testing.T) {
		s.TestMembershipInvitationCreate(t)
		s.TestMembershipInvitationResolve(t)
		s.TestMembershipInvitationUpdate(t)
	})

	t.Run("SystemStore", func(t *testing.T) {
		s.TestSystemGet(t)
		s.TestSystemSet(t)
	})

	t.Run("TunnelStore", func(t *testing.T) {
		s.TestTunnelUpdateDeviceUID(t)
	})
}
