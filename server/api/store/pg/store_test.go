package pg_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/storetest"
	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
)

// TestPgStore runs all store tests against PostgreSQL
// Each sub-suite gets a fresh database with migrations to prevent test pollution
func TestPgStore(t *testing.T) {
	// Run each store interface test suite with its own isolated database
	// This prevents data leakage between test suites and ensures clean state

	runSubSuite(t, "UserStore", func(suite *storetest.Suite, t *testing.T) {
		suite.TestUserList(t)
		suite.TestUserResolve(t)
		suite.TestUserCreate(t)
		suite.TestUserCreatePasswordRoundTrip(t)
		suite.TestUserCreateDuplicate(t)
		suite.TestUserUpdate(t)
		suite.TestUserUpdatePreferredNamespace(t)
		suite.TestUserUpdateDoesNotClobberPreferredNamespace(t)
		suite.TestUserDelete(t)
		suite.TestUserGetInfo(t)
		suite.TestUserConflictsRemoved(t)
	})

	runSubSuite(t, "NamespaceStore", func(suite *storetest.Suite, t *testing.T) {
		suite.TestNamespaceList(t)
		suite.TestNamespaceResolve(t)
		suite.TestNamespaceGetDeviceLimit(t)
		suite.TestNamespaceGetPreferred(t)
		suite.TestNamespaceCreate(t)
		suite.TestNamespaceCreateDuplicate(t)
		suite.TestNamespaceConflicts(t)
		suite.TestNamespaceUpdate(t)
		suite.TestNamespaceUpdateDoesNotClobberDeviceCounts(t)
		suite.TestNamespaceIncrementDeviceCount(t)
		suite.TestNamespaceSyncDeviceCounts(t)
		suite.TestNamespaceDelete(t)
		suite.TestNamespaceDeleteMany(t)
	})

	runSubSuite(t, "DeviceStore", func(suite *storetest.Suite, t *testing.T) {
		suite.TestDeviceList(t)
		suite.TestDeviceResolve(t)
		suite.TestDeviceCreate(t)
		suite.TestDeviceConflicts(t)
		suite.TestDeviceUpdate(t)
		suite.TestDeviceUpdateDoesNotClobberCustomFields(t)
		suite.TestDeviceUpdateDoesNotClobberHeartbeat(t)
		suite.TestDeviceHeartbeat(t)
		suite.TestDeviceOffline(t)
		suite.TestDeviceDelete(t)
		suite.TestDeviceDeleteMany(t)
	})

	runSubSuite(t, "SessionStore", func(suite *storetest.Suite, t *testing.T) {
		suite.TestSessionList(t)
		suite.TestSessionResolve(t)
		suite.TestSessionCreate(t)
		suite.TestSessionUpdateDeviceUID(t)
		suite.TestSessionUpdate(t)
		suite.TestActiveSessionDelete(t)
		suite.TestActiveSessionResolve(t)
		suite.TestActiveSessionUpdate(t)
		suite.TestActiveSessionUpdatePreservesCreatedAt(t)
		suite.TestSessionEventsCreate(t)
		suite.TestSessionEventsList(t)
		suite.TestSessionEventsDelete(t)
		suite.TestSessionCleanup(t)
	})

	runSubSuite(t, "TagStore", func(suite *storetest.Suite, t *testing.T) {
		suite.TestTagCreate(t)
		suite.TestTagConflicts(t)
		suite.TestTagList(t)
		suite.TestTagResolve(t)
		suite.TestTagUpdate(t)
		suite.TestTagPushToTarget(t)
		suite.TestTagPullFromTarget(t)
		suite.TestTagDelete(t)
	})

	runSubSuite(t, "APIKeyStore", func(suite *storetest.Suite, t *testing.T) {
		suite.TestAPIKeyCreate(t)
		suite.TestAPIKeyConflicts(t)
		suite.TestAPIKeyResolve(t)
		suite.TestAPIKeyList(t)
		suite.TestAPIKeyUpdate(t)
		suite.TestAPIKeyDelete(t)
		suite.TestAPIKeyDeleteAllByCreator(t)
	})

	runSubSuite(t, "InstallKeyStore", func(suite *storetest.Suite, t *testing.T) {
		suite.TestInstallKeyModeRoundTrip(t)
		suite.TestInstallKeyEventCreate(t)
		suite.TestInstallKeyEventList(t)
	})

	runSubSuite(t, "PublicKeyStore", func(suite *storetest.Suite, t *testing.T) {
		suite.TestPublicKeyResolve(t)
		suite.TestPublicKeyList(t)
		suite.TestPublicKeyCreate(t)
		suite.TestPublicKeyUpdate(t)
		suite.TestPublicKeyDelete(t)
	})

	runSubSuite(t, "StatsStore", func(suite *storetest.Suite, t *testing.T) {
		suite.TestGetStats(t)
	})

	runSubSuite(t, "PrivateKeyStore", func(suite *storetest.Suite, t *testing.T) {
		suite.TestPrivateKeyCreate(t)
		suite.TestPrivateKeyGet(t)
	})

	runSubSuite(t, "MemberStore", func(suite *storetest.Suite, t *testing.T) {
		suite.TestNamespaceCreateMembership(t)
		suite.TestNamespaceUpdateMembership(t)
		suite.TestNamespaceDeleteMembership(t)
	})

	runSubSuite(t, "MembershipInvitationStore", func(suite *storetest.Suite, t *testing.T) {
		suite.TestMembershipInvitationCreate(t)
		suite.TestMembershipInvitationResolve(t)
		suite.TestMembershipInvitationResolveBySig(t)
		suite.TestMembershipInvitationUpdate(t)
		suite.TestMembershipInvitationDelete(t)
		suite.TestUserMembershipInvitationList(t)
		suite.TestNamespaceMembershipInvitationList(t)
		suite.TestNamespaceMembershipInvitationListWithStatusFilter(t)
	})

	runSubSuite(t, "UserInvitationStore", func(suite *storetest.Suite, t *testing.T) {
		suite.TestUserInvitationsUpsert(t)
		suite.TestUserInvitationGet(t)
		suite.TestUserInvitationUpdate(t)
		suite.TestUserInvitationUpdateDoesNotClobberInvitations(t)
	})

	runSubSuite(t, "SystemStore", func(suite *storetest.Suite, t *testing.T) {
		suite.TestSystemGetDefault(t)
		suite.TestSystemGet(t)
		suite.TestSystemSet(t)
	})

	runSubSuite(t, "TransactionStore", func(suite *storetest.Suite, t *testing.T) {
		suite.TestWithTransaction(t)
	})

	// Cross-namespace isolation for every operation that takes a namespace scope. These are the
	// tests that fail when a converted query stops applying its bound.
	runSubSuite(t, "ScopeIsolation", func(suite *storetest.Suite, t *testing.T) {
		suite.TestScopeIsolationDeviceResolve(t)
		suite.TestScopeIsolationDeviceList(t)
		suite.TestScopeIsolationDeviceConflicts(t)
		suite.TestScopeIsolationSessionResolve(t)
		suite.TestScopeIsolationSessionList(t)
		suite.TestScopeIsolationTagResolve(t)
		suite.TestScopeIsolationTagList(t)
		suite.TestScopeIsolationTagConflicts(t)
		suite.TestScopeIsolationInstallKeyResolve(t)
		suite.TestScopeIsolationInstallKeyList(t)
		suite.TestScopeIsolationInstallKeyConflicts(t)
		suite.TestScopeIsolationInstallKeyResolveSystem(t)
		suite.TestScopeIsolationInstallKeyResolveSystemPairing(t)
		suite.TestScopeIsolationInstallKeyEventList(t)
		suite.TestScopeIsolationInstallKeyEventStampDecision(t)
		suite.TestScopeIsolationAPIKeyResolve(t)
		suite.TestScopeIsolationAPIKeyList(t)
		suite.TestScopeIsolationAPIKeyConflicts(t)
		suite.TestScopeIsolationPublicKeyResolve(t)
		suite.TestScopeIsolationPublicKeyList(t)
		suite.TestScopeIsolationNamespaceGetMembers(t)
		suite.TestScopeIsolationNamespaceIncrementDeviceCount(t)
		suite.TestScopeIsolationMembershipWrites(t)
		suite.TestScopeIsolationMembershipInvitationResolve(t)
		suite.TestScopeIsolationNamespaceMembershipInvitationList(t)
		suite.TestScopeIsolationGetStats(t)
		suite.TestScopeIsolationAccessPolicyList(t)
		suite.TestScopeIsolationAccessPolicyResolve(t)
		suite.TestScopeIsolationSSHIdentityList(t)
		suite.TestScopeIsolationSSHIdentityResolve(t)
		suite.TestScopeRejectsUnconstructedScope(t)
	})
}

// runSubSuite creates a fresh PostgreSQL database for each sub-suite
// This ensures complete isolation between test suites
func runSubSuite(t *testing.T, name string, testFunc func(*storetest.Suite, *testing.T)) {
	t.Run(name, func(t *testing.T) {
		// Create fresh provider with new database + migrations
		ctx := context.Background()
		provider, err := pgprovider.NewProvider(ctx)
		if err != nil {
			t.Fatalf("Failed to create PostgreSQL provider for %s: %v", name, err)
		}
		defer provider.Close(t)

		// Create suite and run tests
		suite := storetest.NewSuite(provider)
		testFunc(suite, t)
	})
}
