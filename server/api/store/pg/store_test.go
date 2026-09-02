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
	runSubSuite(t, "UserStore", func(t *testing.T, suite *storetest.Suite) {
		t.Helper()

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

	runSubSuite(t, "NamespaceStore", func(t *testing.T, suite *storetest.Suite) {
		t.Helper()

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

	runSubSuite(t, "DeviceStore", func(t *testing.T, suite *storetest.Suite) {
		t.Helper()

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

	runSubSuite(t, "SessionStore", func(t *testing.T, suite *storetest.Suite) {
		t.Helper()

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

	runSubSuite(t, "TagStore", func(t *testing.T, suite *storetest.Suite) {
		t.Helper()

		suite.TestTagCreate(t)
		suite.TestTagConflicts(t)
		suite.TestTagList(t)
		suite.TestTagResolve(t)
		suite.TestTagUpdate(t)
		suite.TestTagPushToTarget(t)
		suite.TestTagPullFromTarget(t)
		suite.TestTagDelete(t)
	})

	runSubSuite(t, "APIKeyStore", func(t *testing.T, suite *storetest.Suite) {
		t.Helper()

		suite.TestAPIKeyCreate(t)
		suite.TestAPIKeyConflicts(t)
		suite.TestAPIKeyResolve(t)
		suite.TestAPIKeyList(t)
		suite.TestAPIKeyUpdate(t)
		suite.TestAPIKeyDelete(t)
		suite.TestAPIKeyDeleteAllByCreator(t)
	})

	runSubSuite(t, "InstanceAPIKeyStore", func(t *testing.T, suite *storetest.Suite) {
		t.Helper()

		suite.TestInstanceAPIKeyCreate(t)
		suite.TestInstanceAPIKeyResolve(t)
		suite.TestInstanceAPIKeyList(t)
		suite.TestInstanceAPIKeyDelete(t)
	})

	runSubSuite(t, "InstallKeyStore", func(t *testing.T, suite *storetest.Suite) {
		t.Helper()

		suite.TestInstallKeyModeRoundTrip(t)
		suite.TestInstallKeyEventCreate(t)
		suite.TestInstallKeyEventList(t)
	})

	runSubSuite(t, "PublicKeyStore", func(t *testing.T, suite *storetest.Suite) {
		t.Helper()

		suite.TestPublicKeyResolve(t)
		suite.TestPublicKeyList(t)
		suite.TestPublicKeyCreate(t)
		suite.TestPublicKeyUpdate(t)
		suite.TestPublicKeyDelete(t)
	})

	runSubSuite(t, "StatsStore", func(t *testing.T, suite *storetest.Suite) {
		t.Helper()

		suite.TestGetStats(t)
		suite.TestGetStatsOnlineBoundary(t)
		suite.TestCountRegisteredDevices(t)
	})

	runSubSuite(t, "PrivateKeyStore", func(t *testing.T, suite *storetest.Suite) {
		t.Helper()

		suite.TestPrivateKeyCreate(t)
		suite.TestPrivateKeyGet(t)
	})

	runSubSuite(t, "MemberStore", func(t *testing.T, suite *storetest.Suite) {
		t.Helper()

		suite.TestNamespaceCreateMembership(t)
		suite.TestNamespaceUpdateMembership(t)
		suite.TestNamespaceDeleteMembership(t)
	})

	runSubSuite(t, "MembershipInvitationStore", func(t *testing.T, suite *storetest.Suite) {
		t.Helper()

		suite.TestMembershipInvitationCreate(t)
		suite.TestMembershipInvitationResolve(t)
		suite.TestMembershipInvitationResolveBySig(t)
		suite.TestMembershipInvitationUpdate(t)
		suite.TestMembershipInvitationDelete(t)
		suite.TestUserMembershipInvitationList(t)
		suite.TestNamespaceMembershipInvitationList(t)
		suite.TestNamespaceMembershipInvitationListWithStatusFilter(t)
	})

	runSubSuite(t, "UserInvitationStore", func(t *testing.T, suite *storetest.Suite) {
		t.Helper()

		suite.TestUserInvitationsUpsert(t)
		suite.TestUserInvitationGet(t)
		suite.TestUserInvitationUpdate(t)
		suite.TestUserInvitationUpdateDoesNotClobberInvitations(t)
	})

	runSubSuite(t, "SystemStore", func(t *testing.T, suite *storetest.Suite) {
		t.Helper()

		suite.TestSystemGetDefault(t)
		suite.TestSystemGet(t)
		suite.TestSystemSet(t)
	})

	runSubSuite(t, "TransactionStore", func(t *testing.T, suite *storetest.Suite) {
		t.Helper()

		suite.TestWithTransaction(t)
	})

	runSubSuite(t, "ScopeIsolation", func(t *testing.T, suite *storetest.Suite) {
		t.Helper()

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
		suite.TestScopeIsolationCountRegisteredDevices(t)
		suite.TestScopeIsolationAccessPolicyList(t)
		suite.TestScopeIsolationAccessPolicyResolve(t)
		suite.TestScopeIsolationSSHIdentityList(t)
		suite.TestScopeIsolationSSHIdentityResolve(t)
		suite.TestScopeRejectsUnconstructedScope(t)
	})
}

func runSubSuite(t *testing.T, name string, testFunc func(*testing.T, *storetest.Suite)) {
	t.Helper()

	t.Run(name, func(t *testing.T) {
		ctx := context.Background()
		provider, err := pgprovider.NewProvider(ctx)
		if err != nil {
			t.Fatalf("Failed to create PostgreSQL provider for %s: %v", name, err)
		}
		defer provider.Close(t) //nolint:errcheck

		suite := storetest.NewSuite(provider)
		testFunc(t, suite)
	})
}
