package storetest

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testProvisioningKeyDigest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func (s *Suite) createProvisioningKey(t *testing.T, tenantID string) string {
	t.Helper()

	_, err := s.provider.Store().ProvisioningKeyCreate(context.Background(), &models.ProvisioningKey{
		ID:         testProvisioningKeyDigest,
		Name:       "ci",
		TenantID:   tenantID,
		Mode:       models.ProvisioningKeyModeAutomatic,
		Reusable:   true,
		UsageLimit: 0,
		Tags:       []string{},
		CreatedBy:  "00000000-0000-4000-0000-000000000009",
	})
	require.NoError(t, err)

	return testProvisioningKeyDigest
}

// TestProvisioningKeyEventCreate locks that a key's use is recorded, so its history is complete.
func (s *Suite) TestProvisioningKeyEventCreate(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("appends a row with the denormalized device facts and a stamped id and timestamp", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		digest := s.createProvisioningKey(t, tenantID)

		event := &models.ProvisioningKeyEvent{
			ProvisioningKeyID: digest,
			TenantID:          tenantID,
			DeviceUID:         "device-uid-1",
			Hostname:          "web-01",
			Identity:          "00:1a:2b:3c:4d:5e",
			Info:              &models.DeviceInfo{PrettyName: "Debian GNU/Linux 12", Version: "v0.18.0", Arch: "amd64", Platform: "docker"},
			SourceIP:          "203.0.113.7",
			Ephemeral:         true,
		}
		require.NoError(t, st.ProvisioningKeyEventCreate(ctx, event))

		events, count, err := st.ProvisioningKeyEventList(ctx, scope.MustBounded(tenantID), digest)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		require.Len(t, events, 1)

		got := events[0]
		assert.NotEmpty(t, got.ID)
		assert.False(t, got.Timestamp.IsZero())
		assert.Equal(t, "web-01", got.Hostname)
		assert.Equal(t, "00:1a:2b:3c:4d:5e", got.Identity)
		assert.Equal(t, "203.0.113.7", got.SourceIP)
		assert.True(t, got.Ephemeral)
		require.NotNil(t, got.Info)
		assert.Equal(t, "Debian GNU/Linux 12", got.Info.PrettyName)
		assert.Equal(t, "amd64", got.Info.Arch)
	})

	t.Run("fails when the referenced provisioning key does not exist", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)

		err := st.ProvisioningKeyEventCreate(ctx, &models.ProvisioningKeyEvent{
			ProvisioningKeyID: testProvisioningKeyDigest,
			TenantID:          tenantID,
			DeviceUID:         "device-uid-1",
			Hostname:          "web-01",
		})
		require.Error(t, err)
	})
}

// TestProvisioningKeyEventList locks the ordering and paging of a key's history.
func (s *Suite) TestProvisioningKeyEventList(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	appendEvent := func(t *testing.T, tenantID, digest, hostname string) {
		t.Helper()
		require.NoError(t, st.ProvisioningKeyEventCreate(ctx, &models.ProvisioningKeyEvent{
			ProvisioningKeyID: digest, TenantID: tenantID, DeviceUID: "d-" + hostname, Hostname: hostname,
		}))
	}

	t.Run("scopes by namespace and key and paginates", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		digest := s.createProvisioningKey(t, tenantID)

		otherTenant := s.CreateNamespace(t)
		_, err := st.ProvisioningKeyCreate(ctx, &models.ProvisioningKey{
			ID:   "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			Name: "other", TenantID: otherTenant, Mode: models.ProvisioningKeyModeAutomatic,
			Reusable: true, UsageLimit: 0,
			Tags: []string{}, CreatedBy: "00000000-0000-4000-0000-000000000009",
		})
		require.NoError(t, err)
		require.NoError(t, st.ProvisioningKeyEventCreate(ctx, &models.ProvisioningKeyEvent{
			ProvisioningKeyID: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			TenantID:          otherTenant, DeviceUID: "x", Hostname: "elsewhere",
		}))

		appendEvent(t, tenantID, digest, "a")
		appendEvent(t, tenantID, digest, "b")
		appendEvent(t, tenantID, digest, "c")

		events, count, err := st.ProvisioningKeyEventList(ctx, scope.MustBounded(tenantID), digest,
			st.Options().Paginate(&query.Paginator{Page: 1, PerPage: 2}))
		require.NoError(t, err)
		assert.Equal(t, 3, count)
		assert.Len(t, events, 2)
	})

	t.Run("is append-only: a further append grows the count and never mutates earlier rows", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		digest := s.createProvisioningKey(t, tenantID)

		appendEvent(t, tenantID, digest, "first")

		before, count, err := st.ProvisioningKeyEventList(ctx, scope.MustBounded(tenantID), digest)
		require.NoError(t, err)
		require.Equal(t, 1, count)
		original := before[0]

		appendEvent(t, tenantID, digest, "second")

		after, count, err := st.ProvisioningKeyEventList(ctx, scope.MustBounded(tenantID), digest)
		require.NoError(t, err)
		assert.Equal(t, 2, count)

		var still *models.ProvisioningKeyEvent
		for i := range after {
			if after[i].ID == original.ID {
				still = &after[i]
			}
		}
		require.NotNil(t, still)
		assert.Equal(t, original, *still)
	})

	t.Run("returns empty for a key with no enrollments", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		digest := s.createProvisioningKey(t, tenantID)

		events, count, err := st.ProvisioningKeyEventList(ctx, scope.MustBounded(tenantID), digest)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
		assert.Empty(t, events)
	})
}
