package storetest

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
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
		got, err := st.InstallKeyResolve(ctx, store.InstallKeyNameResolver, "hook", st.Options().InNamespace(tenantID))
		require.NoError(t, err)
		assert.Equal(t, models.InstallKeyModeWebhook, got.Mode)
		assert.Equal(t, "https://hook.example/enroll", got.WebhookURL)
		assert.Equal(t, "s3cr3t", got.WebhookSecret)
		assert.Empty(t, got.AllowedMACs)
	})

	t.Run("persists allowlist mode config", func(t *testing.T) {
		got, err := st.InstallKeyResolve(ctx, store.InstallKeyNameResolver, "allow", st.Options().InNamespace(tenantID))
		require.NoError(t, err)
		assert.Equal(t, models.InstallKeyModeAllowlist, got.Mode)
		assert.Equal(t, []string{"aa:bb:cc:dd:ee:ff", "11:22:33:44:55:66"}, got.AllowedMACs)
		assert.Empty(t, got.WebhookURL)
	})
}
