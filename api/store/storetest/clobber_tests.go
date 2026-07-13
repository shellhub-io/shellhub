package storetest

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

// These tests lock the fix for issue shellhub-io/shellhub#6637: full-model *Update methods must
// not clobber columns that are maintained by separate targeted writes (atomic counters, the
// heartbeat, preferred-namespace clears, the invitation counter). Each test interleaves a targeted
// write with a full-model update built from a stale resolve-time snapshot and asserts the targeted
// column survives. This class of bug is invisible to the mocked-store service tests.

// dbAccessor is implemented by the PostgreSQL provider to expose the underlying Bun driver for
// assertions on columns not surfaced through the store models (e.g. active_sessions.created_at).
type dbAccessor interface {
	DB() *bun.DB
}

// TestNamespaceUpdateDoesNotClobberDeviceCounts ensures a full-model NamespaceUpdate carrying a
// stale device-counter snapshot cannot roll back a concurrent NamespaceIncrementDeviceCount.
func (s *Suite) TestNamespaceUpdateDoesNotClobberDeviceCounts(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	require.NoError(t, s.provider.CleanDatabase(t))

	tenantID := s.CreateNamespace(t, WithNamespaceName("counters"))
	require.NoError(t, st.NamespaceIncrementDeviceCount(ctx, tenantID, models.DeviceStatusAccepted, 5))

	snapshot, err := st.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenantID)
	require.NoError(t, err)
	require.Equal(t, int64(5), snapshot.DevicesAcceptedCount)

	// Concurrent increment lands between the resolve and the save.
	require.NoError(t, st.NamespaceIncrementDeviceCount(ctx, tenantID, models.DeviceStatusAccepted, 1))

	snapshot.Name = "counters-renamed"
	require.NoError(t, st.NamespaceUpdate(ctx, snapshot))

	updated, err := st.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenantID)
	require.NoError(t, err)
	assert.Equal(t, "counters-renamed", updated.Name, "the intended field must still be written")
	assert.Equal(t, int64(6), updated.DevicesAcceptedCount, "concurrent counter increment must not be clobbered")
}

// TestDeviceUpdateDoesNotClobberCustomFields ensures a full-model DeviceUpdate carrying a stale
// custom_fields snapshot cannot drop a concurrently-added custom field.
func (s *Suite) TestDeviceUpdateDoesNotClobberCustomFields(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	require.NoError(t, s.provider.CleanDatabase(t))

	uid := s.CreateDevice(t)
	require.NoError(t, st.DeviceSetCustomField(ctx, string(uid), "first", "1"))

	snapshot, err := st.DeviceResolve(ctx, store.DeviceUIDResolver, string(uid))
	require.NoError(t, err)
	require.Equal(t, map[string]string{"first": "1"}, snapshot.CustomFields)

	// Concurrent set adds a second field between the resolve and the save.
	require.NoError(t, st.DeviceSetCustomField(ctx, string(uid), "second", "2"))

	snapshot.Name = "device-renamed"
	require.NoError(t, st.DeviceUpdate(ctx, snapshot))

	updated, err := st.DeviceResolve(ctx, store.DeviceUIDResolver, string(uid))
	require.NoError(t, err)
	assert.Equal(t, "device-renamed", updated.Name, "the intended field must still be written")
	assert.Equal(t, map[string]string{"first": "1", "second": "2"}, updated.CustomFields,
		"concurrently-added custom field must not be clobbered")
}

// TestDeviceUpdateDoesNotClobberHeartbeat ensures a full-model DeviceUpdate carrying a stale
// last_seen/disconnected_at snapshot cannot roll a device offline by overwriting a fresher
// DeviceHeartbeat.
func (s *Suite) TestDeviceUpdateDoesNotClobberHeartbeat(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	require.NoError(t, s.provider.CleanDatabase(t))

	uid := s.CreateDevice(t)

	snapshot, err := st.DeviceResolve(ctx, store.DeviceUIDResolver, string(uid))
	require.NoError(t, err)

	// Concurrent heartbeat bumps last_seen and clears disconnected_at after the snapshot.
	heartbeat := clock.Now().UTC().Add(2 * time.Hour).Truncate(time.Second)
	modified, err := st.DeviceHeartbeat(ctx, []string{string(uid)}, heartbeat)
	require.NoError(t, err)
	require.Equal(t, int64(1), modified)

	snapshot.Name = "device-renamed"
	require.NoError(t, st.DeviceUpdate(ctx, snapshot))

	updated, err := st.DeviceResolve(ctx, store.DeviceUIDResolver, string(uid))
	require.NoError(t, err)
	assert.Equal(t, "device-renamed", updated.Name, "the intended field must still be written")
	assert.WithinDuration(t, heartbeat, updated.LastSeen, time.Second,
		"concurrent heartbeat last_seen must not be clobbered")
	assert.Nil(t, updated.DisconnectedAt, "device must remain online after the stale update")
}

// TestDeviceOffline covers the targeted disconnected_at write that replaced the full-model update
// in the OfflineDevice service path.
func (s *Suite) TestDeviceOffline(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("stamps disconnected_at without touching last_seen", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		uid := s.CreateDevice(t)

		before, err := st.DeviceResolve(ctx, store.DeviceUIDResolver, string(uid))
		require.NoError(t, err)
		require.Nil(t, before.DisconnectedAt)

		disconnectedAt := clock.Now().UTC().Add(-time.Minute).Truncate(time.Second)
		require.NoError(t, st.DeviceOffline(ctx, string(uid), disconnectedAt))

		updated, err := st.DeviceResolve(ctx, store.DeviceUIDResolver, string(uid))
		require.NoError(t, err)
		require.NotNil(t, updated.DisconnectedAt)
		assert.WithinDuration(t, disconnectedAt, *updated.DisconnectedAt, time.Second)
		assert.WithinDuration(t, before.LastSeen, updated.LastSeen, time.Second, "last_seen must be untouched")
	})

	t.Run("fails for non-existent device", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		err := st.DeviceOffline(ctx, "0000000000000000000000000000000000000000000000000000000000000000", clock.Now())
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})
}

// TestUserUpdateDoesNotClobberPreferredNamespace ensures a full-model UserUpdate carrying a stale
// preferred_namespace_id cannot restore a preference a concurrent removal just cleared.
func (s *Suite) TestUserUpdateDoesNotClobberPreferredNamespace(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	require.NoError(t, s.provider.CleanDatabase(t))

	userID := s.CreateUser(t)
	tenantID := s.CreateNamespace(t, WithOwner(userID))
	require.NoError(t, st.UserUpdatePreferredNamespace(ctx, userID, tenantID))

	snapshot, err := st.UserResolve(ctx, store.UserIDResolver, userID)
	require.NoError(t, err)
	require.Equal(t, tenantID, snapshot.Preferences.PreferredNamespace)

	// Concurrent membership/namespace removal clears the preference after the snapshot.
	require.NoError(t, st.UserUpdatePreferredNamespace(ctx, userID, ""))

	snapshot.Name = "renamed"
	require.NoError(t, st.UserUpdate(ctx, snapshot))

	updated, err := st.UserResolve(ctx, store.UserIDResolver, userID)
	require.NoError(t, err)
	assert.Equal(t, "renamed", updated.Name, "the intended field must still be written")
	assert.Empty(t, updated.Preferences.PreferredNamespace, "concurrent preferred-namespace clear must not be clobbered")
}

// TestUserUpdatePreferredNamespace covers the targeted preferred_namespace_id write used by login.
func (s *Suite) TestUserUpdatePreferredNamespace(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("sets and clears the preferred namespace", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		userID := s.CreateUser(t)
		tenantID := s.CreateNamespace(t, WithOwner(userID))

		require.NoError(t, st.UserUpdatePreferredNamespace(ctx, userID, tenantID))
		user, err := st.UserResolve(ctx, store.UserIDResolver, userID)
		require.NoError(t, err)
		assert.Equal(t, tenantID, user.Preferences.PreferredNamespace)

		require.NoError(t, st.UserUpdatePreferredNamespace(ctx, userID, ""))
		user, err = st.UserResolve(ctx, store.UserIDResolver, userID)
		require.NoError(t, err)
		assert.Empty(t, user.Preferences.PreferredNamespace)
	})

	t.Run("fails for non-existent user", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		err := st.UserUpdatePreferredNamespace(ctx, "00000000-0000-4000-8000-000000000000", "")
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})
}

// TestUserInvitationUpdateDoesNotClobberInvitations ensures a status-change UserInvitationUpdate
// carrying a stale count cannot clobber a concurrent re-invite increment.
func (s *Suite) TestUserInvitationUpdateDoesNotClobberInvitations(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	require.NoError(t, s.provider.CleanDatabase(t))

	id, err := st.UserInvitationsUpsert(ctx, "invitee@test.com")
	require.NoError(t, err)

	snapshot, err := st.UserInvitationGet(ctx, store.UserInvitationIDResolver, id)
	require.NoError(t, err)
	require.Equal(t, 1, snapshot.Invitations)

	// Concurrent re-invite increments the counter between the resolve and the save.
	_, err = st.UserInvitationsUpsert(ctx, "invitee@test.com")
	require.NoError(t, err)

	snapshot.Status = models.UserInvitationStatusAccepted
	require.NoError(t, st.UserInvitationUpdate(ctx, snapshot))

	updated, err := st.UserInvitationGet(ctx, store.UserInvitationIDResolver, id)
	require.NoError(t, err)
	assert.Equal(t, models.UserInvitationStatusAccepted, updated.Status, "the intended field must still be written")
	assert.Equal(t, 2, updated.Invitations, "concurrent re-invite increment must not be clobbered")
}

// TestActiveSessionUpdatePreservesCreatedAt ensures ActiveSessionUpdate no longer resets the row's
// creation timestamp to "now" (ActiveSessionFromModel always stamps clock.Now()).
func (s *Suite) TestActiveSessionUpdatePreservesCreatedAt(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	provider, ok := s.provider.(dbAccessor)
	if !ok {
		t.Skip("provider does not expose the raw DB; created_at cannot be asserted")
	}
	db := provider.DB()

	require.NoError(t, s.provider.CleanDatabase(t))

	sessionUID := s.CreateSession(t, WithSessionActive(true))

	// Pin created_at to a fixed point in the past so a reset to "now" is unambiguous.
	createdAt := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := db.NewUpdate().
		Table("active_sessions").
		Set("created_at = ?", createdAt).
		Where("session_id = ?", string(sessionUID)).
		Exec(ctx)
	require.NoError(t, err)

	require.NoError(t, st.ActiveSessionUpdate(ctx, &models.ActiveSession{UID: sessionUID, LastSeen: clock.Now()}))

	var got time.Time
	require.NoError(t, db.NewSelect().
		Table("active_sessions").
		Column("created_at").
		Where("session_id = ?", string(sessionUID)).
		Scan(ctx, &got))

	assert.WithinDuration(t, createdAt, got, 24*time.Hour, "created_at must not be reset on update")
	assert.True(t, got.Before(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)),
		"created_at must not be reset to the update time")
}
