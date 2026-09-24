package storetest

import (
	"context"
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
		Tags:          []string{},
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
		Tags:              []string{},
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
		Tags:      []string{},
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
		Tags:      []string{},
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
