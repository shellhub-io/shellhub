package storetest

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testInstallKeyDigest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

// createInstallKey inserts a minimal, valid install key so enrollment events can reference it via
// the composite FK (install_key_id, namespace_id).
func (s *Suite) createInstallKey(t *testing.T, tenantID string) string {
	t.Helper()

	_, err := s.provider.Store().InstallKeyCreate(context.Background(), &models.InstallKey{
		ID:         testInstallKeyDigest,
		Name:       "ci",
		TenantID:   tenantID,
		Mode:       models.InstallKeyModeAutomatic,
		Reusable:   true,
		UsageLimit: 0,
		Tags:       []string{},
		CreatedBy:  "00000000-0000-4000-0000-000000000009",
	})
	require.NoError(t, err)

	return testInstallKeyDigest
}

func (s *Suite) TestInstallKeyEventCreate(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("appends a row with the denormalized device facts and a stamped id and timestamp", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		digest := s.createInstallKey(t, tenantID)

		event := &models.InstallKeyEvent{
			InstallKeyID: digest,
			TenantID:     tenantID,
			DeviceUID:    "device-uid-1",
			Hostname:     "web-01",
			MAC:          "00:1a:2b:3c:4d:5e",
			Info:         &models.DeviceInfo{PrettyName: "Debian GNU/Linux 12", Version: "v0.18.0", Arch: "amd64", Platform: "docker"},
			SourceIP:     "203.0.113.7",
			Ephemeral:    true,
		}
		require.NoError(t, st.InstallKeyEventCreate(ctx, event))

		events, count, err := st.InstallKeyEventList(ctx, tenantID, digest)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		require.Len(t, events, 1)

		got := events[0]
		assert.NotEmpty(t, got.ID)
		assert.False(t, got.Timestamp.IsZero())
		assert.Equal(t, "web-01", got.Hostname)
		assert.Equal(t, "00:1a:2b:3c:4d:5e", got.MAC)
		assert.Equal(t, "203.0.113.7", got.SourceIP)
		assert.True(t, got.Ephemeral)
		require.NotNil(t, got.Info)
		assert.Equal(t, "Debian GNU/Linux 12", got.Info.PrettyName)
		assert.Equal(t, "amd64", got.Info.Arch)
	})

	t.Run("fails when the referenced install key does not exist", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)

		err := st.InstallKeyEventCreate(ctx, &models.InstallKeyEvent{
			InstallKeyID: testInstallKeyDigest,
			TenantID:     tenantID,
			DeviceUID:    "device-uid-1",
			Hostname:     "web-01",
		})
		require.Error(t, err)
	})
}

func (s *Suite) TestInstallKeyEventList(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	appendEvent := func(t *testing.T, tenantID, digest, hostname string) {
		t.Helper()
		require.NoError(t, st.InstallKeyEventCreate(ctx, &models.InstallKeyEvent{
			InstallKeyID: digest, TenantID: tenantID, DeviceUID: "d-" + hostname, Hostname: hostname,
		}))
	}

	t.Run("scopes by namespace and key and paginates", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		digest := s.createInstallKey(t, tenantID)

		// A second namespace + key: its events must not leak into the first key's history.
		otherTenant := s.CreateNamespace(t)
		_, err := st.InstallKeyCreate(ctx, &models.InstallKey{
			ID:   "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			Name: "other", TenantID: otherTenant, Mode: models.InstallKeyModeAutomatic,
			Reusable: true, UsageLimit: 0,
			Tags: []string{}, CreatedBy: "00000000-0000-4000-0000-000000000009",
		})
		require.NoError(t, err)
		require.NoError(t, st.InstallKeyEventCreate(ctx, &models.InstallKeyEvent{
			InstallKeyID: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			TenantID:     otherTenant, DeviceUID: "x", Hostname: "elsewhere",
		}))

		appendEvent(t, tenantID, digest, "a")
		appendEvent(t, tenantID, digest, "b")
		appendEvent(t, tenantID, digest, "c")

		events, count, err := st.InstallKeyEventList(ctx, tenantID, digest,
			st.Options().Paginate(&query.Paginator{Page: 1, PerPage: 2}),
		)
		require.NoError(t, err)
		assert.Equal(t, 3, count)
		assert.Len(t, events, 2)
	})

	t.Run("is append-only: a further append grows the count and never mutates earlier rows", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		digest := s.createInstallKey(t, tenantID)

		appendEvent(t, tenantID, digest, "first")

		before, count, err := st.InstallKeyEventList(ctx, tenantID, digest)
		require.NoError(t, err)
		require.Equal(t, 1, count)
		original := before[0]

		appendEvent(t, tenantID, digest, "second")

		after, count, err := st.InstallKeyEventList(ctx, tenantID, digest)
		require.NoError(t, err)
		assert.Equal(t, 2, count)

		// The earlier row is byte-identical after the second append.
		var still *models.InstallKeyEvent
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
		digest := s.createInstallKey(t, tenantID)

		events, count, err := st.InstallKeyEventList(ctx, tenantID, digest)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
		assert.Empty(t, events)
	})
}
