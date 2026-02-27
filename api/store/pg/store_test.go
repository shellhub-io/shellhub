package pg_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/store/storetest"
	"github.com/shellhub-io/shellhub/api/store/storetest/pgprovider"
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
		suite.TestUserConflicts(t)
		suite.TestUserUpdate(t)
		suite.TestUserDelete(t)
		suite.TestUserGetInfo(t)
	})

	runSubSuite(t, "NamespaceStore", func(suite *storetest.Suite, t *testing.T) {
		suite.TestNamespaceList(t)
		suite.TestNamespaceResolve(t)
		suite.TestNamespaceGetPreferred(t)
		suite.TestNamespaceCreate(t)
		suite.TestNamespaceConflicts(t)
		suite.TestNamespaceUpdate(t)
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
		suite.TestDeviceHeartbeat(t)
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
		suite.TestSessionEventsCreate(t)
		suite.TestSessionEventsList(t)
		suite.TestSessionEventsDelete(t)
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

	runSubSuite(t, "UserInvitationsStore", func(suite *storetest.Suite, t *testing.T) {
		suite.TestUserInvitationsUpsert(t)
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
		suite.TestMembershipInvitationUpdate(t)
	})

	runSubSuite(t, "SystemStore", func(suite *storetest.Suite, t *testing.T) {
		suite.TestSystemGet(t)
		suite.TestSystemSet(t)
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
