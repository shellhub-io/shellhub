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

// TestInstallKeyModeRoundTrip verifies the enrollment-mode columns (mode, webhook_url,
// webhook_secret, allowed_macs) persist and read back intact through the store.
func (s *Suite) TestInstallKeyModeRoundTrip(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	require.NoError(t, s.provider.CleanDatabase(t))
	tenantID := s.CreateNamespace(t)

	const owner = "00000000-0000-4000-0000-000000000009"

	webhookDigest := "1111111111111111111111111111111111111111111111111111111111111111"
	_, err := st.InstallKeyCreate(ctx, &models.InstallKey{
		ID:            webhookDigest,
		Name:          "hook",
		TenantID:      tenantID,
		Mode:          models.InstallKeyModeWebhook,
		WebhookURL:    "https://hook.example/enroll",
		WebhookSecret: "s3cr3t",
		Reusable:      true,
		Tags:          []string{},
		CreatedBy:     owner,
	})
	require.NoError(t, err)

	allowlistDigest := "2222222222222222222222222222222222222222222222222222222222222222"
	_, err = st.InstallKeyCreate(ctx, &models.InstallKey{
		ID:          allowlistDigest,
		Name:        "allow",
		TenantID:    tenantID,
		Mode:        models.InstallKeyModeAllowlist,
		AllowedMACs: []string{"aa:bb:cc:dd:ee:ff", "11:22:33:44:55:66"},
		Reusable:    true,
		Tags:        []string{},
		CreatedBy:   owner,
	})
	require.NoError(t, err)

	t.Run("persists webhook mode config", func(t *testing.T) {
		got, err := st.InstallKeyResolve(ctx, scope.MustBounded(tenantID), store.InstallKeyNameResolver, "hook")
		require.NoError(t, err)
		assert.Equal(t, models.InstallKeyModeWebhook, got.Mode)
		assert.Equal(t, "https://hook.example/enroll", got.WebhookURL)
		assert.Equal(t, "s3cr3t", got.WebhookSecret)
		assert.Empty(t, got.AllowedMACs)
	})

	t.Run("persists allowlist mode config", func(t *testing.T) {
		got, err := st.InstallKeyResolve(ctx, scope.MustBounded(tenantID), store.InstallKeyNameResolver, "allow")
		require.NoError(t, err)
		assert.Equal(t, models.InstallKeyModeAllowlist, got.Mode)
		assert.Equal(t, []string{"aa:bb:cc:dd:ee:ff", "11:22:33:44:55:66"}, got.AllowedMACs)
		assert.Empty(t, got.WebhookURL)
	})
}

// TestInstallKeyListPendingDevices verifies the per-key count of enrollments awaiting a decision,
// which is what tells the keys list which key has something to review.
func (s *Suite) TestInstallKeyListPendingDevices(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	require.NoError(t, s.provider.CleanDatabase(t))
	tenantID := s.CreateNamespace(t)

	const owner = "00000000-0000-4000-0000-000000000009"

	waitingDigest := "3333333333333333333333333333333333333333333333333333333333333333"
	_, err := st.InstallKeyCreate(ctx, &models.InstallKey{
		ID:        waitingDigest,
		Name:      "waiting",
		TenantID:  tenantID,
		Mode:      models.InstallKeyModeManual,
		Reusable:  true,
		Tags:      []string{},
		CreatedBy: owner,
	})
	require.NoError(t, err)

	settledDigest := "4444444444444444444444444444444444444444444444444444444444444444"
	_, err = st.InstallKeyCreate(ctx, &models.InstallKey{
		ID:        settledDigest,
		Name:      "settled",
		TenantID:  tenantID,
		Mode:      models.InstallKeyModeAutomatic,
		Reusable:  true,
		Tags:      []string{},
		CreatedBy: owner,
	})
	require.NoError(t, err)

	s.CreateDevice(t, WithTenantID(tenantID), WithDeviceInstallKey(waitingDigest), WithDeviceStatus(models.DeviceStatusPending))
	s.CreateDevice(t, WithTenantID(tenantID), WithDeviceInstallKey(waitingDigest), WithDeviceStatus(models.DeviceStatusPending))
	s.CreateDevice(t, WithTenantID(tenantID), WithDeviceInstallKey(waitingDigest), WithDeviceStatus(models.DeviceStatusAccepted))
	s.CreateDevice(t, WithTenantID(tenantID), WithDeviceInstallKey(settledDigest), WithDeviceStatus(models.DeviceStatusAccepted))
	s.CreateDevice(t, WithTenantID(tenantID), WithDeviceInstallKey(settledDigest), WithDeviceStatus(models.DeviceStatusRejected))

	t.Run("counts only the devices a key still owes a decision", func(t *testing.T) {
		installKeys, _, err := st.InstallKeyList(ctx, scope.MustBounded(tenantID))
		require.NoError(t, err)

		counts := make(map[string]int, len(installKeys))
		for _, key := range installKeys {
			counts[key.Name] = key.PendingDevices
		}

		assert.Equal(t, 2, counts["waiting"])
		assert.Equal(t, 0, counts["settled"])
	})
}
