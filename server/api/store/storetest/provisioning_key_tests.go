package storetest

import (
	"context"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProvisioningKeyModeRoundTrip verifies the enrollment-mode columns (mode, webhook_url,
// webhook_secret, allowed_identities) persist and read back intact through the store.
func (s *Suite) TestProvisioningKeyModeRoundTrip(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	require.NoError(t, s.provider.CleanDatabase(t))
	tenantID := s.CreateNamespace(t)

	const owner = "00000000-0000-4000-0000-000000000009"

	webhookDigest := "1111111111111111111111111111111111111111111111111111111111111111"
	_, err := st.ProvisioningKeyCreate(ctx, &models.ProvisioningKey{
		ID:            webhookDigest,
		Name:          "hook",
		TenantID:      tenantID,
		Mode:          models.ProvisioningKeyModeWebhook,
		WebhookURL:    "https://hook.example/enroll",
		WebhookSecret: "s3cr3t",
		Reusable:      true,
		CreatedBy:     owner,
	})
	require.NoError(t, err)

	allowlistDigest := "2222222222222222222222222222222222222222222222222222222222222222"
	_, err = st.ProvisioningKeyCreate(ctx, &models.ProvisioningKey{
		ID:                allowlistDigest,
		Name:              "allow",
		TenantID:          tenantID,
		Mode:              models.ProvisioningKeyModeAllowlist,
		AllowedIdentities: []string{"aa:bb:cc:dd:ee:ff", "11:22:33:44:55:66"},
		Reusable:          true,
		CreatedBy:         owner,
	})
	require.NoError(t, err)

	t.Run("persists webhook mode config", func(t *testing.T) {
		got, err := st.ProvisioningKeyResolve(ctx, scope.MustBounded(tenantID), store.ProvisioningKeyNameResolver, "hook")
		require.NoError(t, err)
		assert.Equal(t, models.ProvisioningKeyModeWebhook, got.Mode)
		assert.Equal(t, "https://hook.example/enroll", got.WebhookURL)
		assert.Equal(t, "s3cr3t", got.WebhookSecret)
		assert.Empty(t, got.AllowedIdentities)
	})

	t.Run("persists allowlist mode config", func(t *testing.T) {
		got, err := st.ProvisioningKeyResolve(ctx, scope.MustBounded(tenantID), store.ProvisioningKeyNameResolver, "allow")
		require.NoError(t, err)
		assert.Equal(t, models.ProvisioningKeyModeAllowlist, got.Mode)
		assert.Equal(t, []string{"aa:bb:cc:dd:ee:ff", "11:22:33:44:55:66"}, got.AllowedIdentities)
		assert.Empty(t, got.WebhookURL)
	})
}

// TestProvisioningKeyTags verifies a key holds its tags themselves: every read reports a renamed tag
// under its new name, and an update replaces the tags the key holds.
func (s *Suite) TestProvisioningKeyTags(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	require.NoError(t, s.provider.CleanDatabase(t))
	tenantID := s.CreateNamespace(t)
	sc := scope.MustBounded(tenantID)

	fleet := s.CreateTag(t, WithTagTenant(tenantID), WithTagName("fleet"))
	edge := s.CreateTag(t, WithTagTenant(tenantID), WithTagName("edge"))

	createKey := func(t *testing.T, digest, name string, tagIDs []string) {
		t.Helper()

		_, err := st.ProvisioningKeyCreate(ctx, &models.ProvisioningKey{
			ID:        strings.Repeat(digest, 64),
			Name:      name,
			TenantID:  tenantID,
			Mode:      models.ProvisioningKeyModeAutomatic,
			Reusable:  true,
			TagIDs:    tagIDs,
			CreatedBy: "00000000-0000-4000-0000-000000000009",
		})
		require.NoError(t, err)
	}

	createKey(t, "5", "tagged", []string{fleet, edge})

	renamed, err := st.TagResolve(ctx, sc, store.TagIDResolver, fleet)
	require.NoError(t, err)
	renamed.Name = "fleetwide"
	require.NoError(t, st.TagUpdate(ctx, renamed))

	t.Run("resolving reports a renamed tag under its new name", func(t *testing.T) {
		got, err := st.ProvisioningKeyResolve(ctx, sc, store.ProvisioningKeyNameResolver, "tagged")
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{fleet, edge}, got.TagIDs)
		assert.ElementsMatch(t, []string{"fleetwide", "edge"}, got.Tags)
	})

	t.Run("listing reports a renamed tag under its new name", func(t *testing.T) {
		keys, _, err := st.ProvisioningKeyList(ctx, sc)
		require.NoError(t, err)

		for _, key := range keys {
			if key.Name == "tagged" {
				assert.ElementsMatch(t, []string{"fleetwide", "edge"}, key.Tags)

				return
			}
		}

		t.Fatal("the key is not listed")
	})

	t.Run("an update replaces the tags the key holds", func(t *testing.T) {
		createKey(t, "6", "replaced", []string{fleet, edge})

		key, err := st.ProvisioningKeyResolve(ctx, sc, store.ProvisioningKeyNameResolver, "replaced")
		require.NoError(t, err)

		key.TagIDs = []string{edge}
		require.NoError(t, st.ProvisioningKeyUpdate(ctx, key))

		got, err := st.ProvisioningKeyResolve(ctx, sc, store.ProvisioningKeyNameResolver, "replaced")
		require.NoError(t, err)
		assert.Equal(t, []string{edge}, got.TagIDs)
		assert.Equal(t, []string{"edge"}, got.Tags)
	})

	t.Run("an update without TagIDs leaves the tags the key holds", func(t *testing.T) {
		createKey(t, "7", "kept", []string{fleet, edge})

		key, err := st.ProvisioningKeyResolve(ctx, sc, store.ProvisioningKeyNameResolver, "kept")
		require.NoError(t, err)

		key.TagIDs = nil
		key.Disabled = true
		require.NoError(t, st.ProvisioningKeyUpdate(ctx, key))

		got, err := st.ProvisioningKeyResolve(ctx, sc, store.ProvisioningKeyNameResolver, "kept")
		require.NoError(t, err)
		assert.True(t, got.Disabled)
		assert.ElementsMatch(t, []string{fleet, edge}, got.TagIDs)
	})

	t.Run("an update with empty TagIDs drops every tag the key holds", func(t *testing.T) {
		createKey(t, "8", "emptied", []string{fleet, edge})

		key, err := st.ProvisioningKeyResolve(ctx, sc, store.ProvisioningKeyNameResolver, "emptied")
		require.NoError(t, err)

		key.TagIDs = []string{}
		require.NoError(t, st.ProvisioningKeyUpdate(ctx, key))

		got, err := st.ProvisioningKeyResolve(ctx, sc, store.ProvisioningKeyNameResolver, "emptied")
		require.NoError(t, err)
		assert.Empty(t, got.TagIDs)
	})
}

// TestProvisioningKeyListPendingDevices verifies the per-key count of enrollments awaiting a decision,
// which is what tells the keys list which key has something to review.
func (s *Suite) TestProvisioningKeyListPendingDevices(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	require.NoError(t, s.provider.CleanDatabase(t))
	tenantID := s.CreateNamespace(t)

	const owner = "00000000-0000-4000-0000-000000000009"

	waitingDigest := "3333333333333333333333333333333333333333333333333333333333333333"
	_, err := st.ProvisioningKeyCreate(ctx, &models.ProvisioningKey{
		ID:        waitingDigest,
		Name:      "waiting",
		TenantID:  tenantID,
		Mode:      models.ProvisioningKeyModeManual,
		Reusable:  true,
		CreatedBy: owner,
	})
	require.NoError(t, err)

	settledDigest := "4444444444444444444444444444444444444444444444444444444444444444"
	_, err = st.ProvisioningKeyCreate(ctx, &models.ProvisioningKey{
		ID:        settledDigest,
		Name:      "settled",
		TenantID:  tenantID,
		Mode:      models.ProvisioningKeyModeAutomatic,
		Reusable:  true,
		CreatedBy: owner,
	})
	require.NoError(t, err)

	s.CreateDevice(t, WithTenantID(tenantID), WithDeviceProvisioningKey(waitingDigest), WithDeviceStatus(models.DeviceStatusPending))
	s.CreateDevice(t, WithTenantID(tenantID), WithDeviceProvisioningKey(waitingDigest), WithDeviceStatus(models.DeviceStatusPending))
	s.CreateDevice(t, WithTenantID(tenantID), WithDeviceProvisioningKey(waitingDigest), WithDeviceStatus(models.DeviceStatusAccepted))
	s.CreateDevice(t, WithTenantID(tenantID), WithDeviceProvisioningKey(settledDigest), WithDeviceStatus(models.DeviceStatusAccepted))
	s.CreateDevice(t, WithTenantID(tenantID), WithDeviceProvisioningKey(settledDigest), WithDeviceStatus(models.DeviceStatusRejected))

	t.Run("counts only the devices a key still owes a decision", func(t *testing.T) {
		provisioningKeys, _, err := st.ProvisioningKeyList(ctx, scope.MustBounded(tenantID))
		require.NoError(t, err)

		counts := make(map[string]int, len(provisioningKeys))
		for _, key := range provisioningKeys {
			counts[key.Name] = key.PendingDevices
		}

		assert.Equal(t, 2, counts["waiting"])
		assert.Equal(t, 0, counts["settled"])
	})
}
