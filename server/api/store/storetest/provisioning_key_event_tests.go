package storetest

import (
	"context"
	"sync"
	"testing"
	"time"

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
		assert.Equal(t, int64(1), count)
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

	refusal := func(tenantID, digest, sourceIP string) *models.ProvisioningKeyEvent {
		return &models.ProvisioningKeyEvent{
			ProvisioningKeyID: digest,
			TenantID:          tenantID,
			DeviceUID:         "refused-uid",
			Hostname:          "old-box",
			SourceIP:          sourceIP,
			Refusal:           models.ProvisioningKeyEventRefusalUnverifiedDeviceKey,
		}
	}

	history := func(t *testing.T, tenantID, digest string) []models.ProvisioningKeyEvent {
		t.Helper()

		events, _, err := st.ProvisioningKeyEventList(ctx, scope.MustBounded(tenantID), digest)
		require.NoError(t, err)

		return events
	}

	t.Run("refreshes a repeated refusal instead of appending one", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		digest := s.createProvisioningKey(t, tenantID)

		require.NoError(t, st.ProvisioningKeyEventCreate(ctx, refusal(tenantID, digest, "192.0.2.1")))
		first := history(t, tenantID, digest)[0].Timestamp

		s.Clock().Set(first.Add(time.Minute))
		require.NoError(t, st.ProvisioningKeyEventCreate(ctx, refusal(tenantID, digest, "192.0.2.2")))

		events := history(t, tenantID, digest)
		require.Len(t, events, 1)
		assert.Equal(t, 2, events[0].Attempts)
		assert.Equal(t, "192.0.2.2", events[0].SourceIP)
		assert.WithinDuration(t, first.Add(time.Minute), events[0].Timestamp, time.Millisecond)
	})

	t.Run("starts a new refusal after the device enrolled, leaving the old one as it was", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		digest := s.createProvisioningKey(t, tenantID)
		enrollment := refusal(tenantID, digest, "192.0.2.1")
		enrollment.Refusal = ""

		require.NoError(t, st.ProvisioningKeyEventCreate(ctx, refusal(tenantID, digest, "192.0.2.1")))
		s.Clock().Set(s.Clock().Now().Add(time.Minute))
		require.NoError(t, st.ProvisioningKeyEventCreate(ctx, enrollment))
		s.Clock().Set(s.Clock().Now().Add(time.Minute))
		require.NoError(t, st.ProvisioningKeyEventCreate(ctx, refusal(tenantID, digest, "192.0.2.3")))

		events := history(t, tenantID, digest)
		require.Len(t, events, 3)
		for _, event := range events {
			assert.Equal(t, 1, event.Attempts)
		}
	})

	t.Run("stamps a decision on the enrollment, not on a newer refusal", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		digest := s.createProvisioningKey(t, tenantID)
		enrollment := refusal(tenantID, digest, "192.0.2.1")
		enrollment.Refusal = ""

		require.NoError(t, st.ProvisioningKeyEventCreate(ctx, enrollment))
		s.Clock().Set(s.Clock().Now().Add(time.Minute))
		require.NoError(t, st.ProvisioningKeyEventCreate(ctx, refusal(tenantID, digest, "192.0.2.2")))

		decidedAt := s.Clock().Now().Add(time.Minute)
		require.NoError(t, st.ProvisioningKeyEventStampDecision(ctx, scope.MustBounded(tenantID), "refused-uid", models.DeviceStatusAccepted, decidedAt))

		events := history(t, tenantID, digest)
		require.Len(t, events, 2)
		for _, event := range events {
			if event.Refusal != "" {
				assert.Empty(t, event.DecidedStatus, "a refusal carries no decision")
				assert.Nil(t, event.DecidedAt)

				continue
			}

			assert.Equal(t, models.DeviceStatusAccepted, event.DecidedStatus)
			require.NotNil(t, event.DecidedAt)
			assert.WithinDuration(t, decidedAt, *event.DecidedAt, time.Millisecond)
		}
	})

	t.Run("records the same device's refusal on another key as its own", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		digest := s.createProvisioningKey(t, tenantID)
		tenantKey, err := st.ProvisioningKeyResolveSystem(ctx, scope.MustBounded(tenantID))
		require.NoError(t, err)

		require.NoError(t, st.ProvisioningKeyEventCreate(ctx, refusal(tenantID, digest, "192.0.2.1")))
		s.Clock().Set(s.Clock().Now().Add(time.Minute))
		require.NoError(t, st.ProvisioningKeyEventCreate(ctx, refusal(tenantID, tenantKey.ID, "192.0.2.1")))

		assert.Len(t, history(t, tenantID, digest), 1)
		assert.Len(t, history(t, tenantID, tenantKey.ID), 1)
	})

	t.Run("keeps one row for refusals of the same device that arrive together", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		digest := s.createProvisioningKey(t, tenantID)

		const attempts = 64

		var wg sync.WaitGroup
		errs := make(chan error, attempts)
		for range attempts {
			wg.Go(func() {
				errs <- st.ProvisioningKeyEventCreate(ctx, refusal(tenantID, digest, "192.0.2.1"))
			})
		}
		wg.Wait()
		close(errs)

		for err := range errs {
			require.NoError(t, err)
		}

		events := history(t, tenantID, digest)
		require.Len(t, events, 1)
		assert.Equal(t, attempts, events[0].Attempts)
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
			CreatedBy: "00000000-0000-4000-0000-000000000009",
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
		assert.Equal(t, int64(3), count)
		assert.Len(t, events, 2)
	})

	t.Run("is append-only: a further append grows the count and never mutates earlier rows", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		digest := s.createProvisioningKey(t, tenantID)

		appendEvent(t, tenantID, digest, "first")

		before, count, err := st.ProvisioningKeyEventList(ctx, scope.MustBounded(tenantID), digest)
		require.NoError(t, err)
		require.Equal(t, int64(1), count)
		original := before[0]

		appendEvent(t, tenantID, digest, "second")

		after, count, err := st.ProvisioningKeyEventList(ctx, scope.MustBounded(tenantID), digest)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)

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
		assert.Equal(t, int64(0), count)
		assert.Empty(t, events)
	})
}
