package storetest

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDeviceListAfter tests reading a namespace's devices page after page by cursor.
func (s *Suite) TestDeviceListAfter(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	base := time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC)
	minute := func(n int) time.Time { return base.Add(time.Duration(n) * time.Minute) }

	createAt := func(t *testing.T, createdAt time.Time, opts ...DeviceOption) string {
		t.Helper()

		now := s.Clock().Now()
		s.Clock().Set(createdAt)
		defer s.Clock().Set(now)

		return string(s.CreateDevice(t, opts...))
	}

	readAll := func(t *testing.T, sc scope.Scope, size int64, opts ...store.QueryOption) []models.Device {
		t.Helper()

		var all []models.Device
		var after *query.Cursor

		for range 100 {
			page, next, err := st.DeviceListAfter(ctx, sc, store.DeviceAcceptableIfNotAccepted, after, size, opts...)
			require.NoError(t, err)

			all = append(all, page...)
			if next == nil {
				return all
			}

			require.Len(t, page, int(size), "a page with a next cursor is full")
			after = next
		}

		t.Fatal("the read did not end")

		return nil
	}

	uids := func(devices []models.Device) []string {
		out := make([]string, len(devices))
		for i, d := range devices {
			out[i] = d.UID
		}

		return out
	}

	t.Run("returns every device once in creation order", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		offsets := []int{5, 1, 6, 0, 3, 2, 4}
		want := make([]string, len(offsets))
		for _, offset := range offsets {
			want[offset] = createAt(t, minute(offset), WithTenantID(tenantID))
		}

		got := readAll(t, scope.MustBounded(tenantID), 3)

		assert.Equal(t, want, uids(got))
	})

	t.Run("returns devices sharing a creation time once each, by id", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		want := make([]string, 5)
		for i := range want {
			want[i] = createAt(t, base, WithTenantID(tenantID))
		}

		got := readAll(t, scope.MustBounded(tenantID), 2)

		assert.Equal(t, want, uids(got))
	})

	t.Run("keeps creation times a microsecond apart in order across pages", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		want := make([]string, 4)
		for i := range want {
			want[len(want)-1-i] = createAt(t, base.Add(time.Duration(len(want)-1-i)*time.Microsecond), WithTenantID(tenantID))
		}

		got := readAll(t, scope.MustBounded(tenantID), 1)

		assert.Equal(t, want, uids(got))
	})

	t.Run("ends on a full page when the namespace is a multiple of the page size", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		for i := range 4 {
			createAt(t, minute(i), WithTenantID(tenantID))
		}

		sc := scope.MustBounded(tenantID)
		first, next, err := st.DeviceListAfter(ctx, sc, store.DeviceAcceptableIfNotAccepted, nil, 2)
		require.NoError(t, err)
		require.Len(t, first, 2)
		require.NotNil(t, next)

		last, next, err := st.DeviceListAfter(ctx, sc, store.DeviceAcceptableIfNotAccepted, next, 2)
		require.NoError(t, err)
		assert.Len(t, last, 2)
		assert.Nil(t, next)
	})

	t.Run("returns no devices and no cursor for an empty namespace", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)

		devices, next, err := st.DeviceListAfter(ctx, scope.MustBounded(tenantID), store.DeviceAcceptableIfNotAccepted, nil, 10)
		require.NoError(t, err)
		assert.Empty(t, devices)
		assert.Nil(t, next)
	})

	t.Run("leaves out the devices of another namespace", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		mine := s.CreateNamespace(t)
		other := s.CreateNamespace(t)
		want := []string{
			createAt(t, minute(0), WithTenantID(mine)),
			createAt(t, minute(2), WithTenantID(mine)),
		}
		createAt(t, minute(1), WithTenantID(other))
		createAt(t, minute(3), WithTenantID(other))

		got := readAll(t, scope.MustBounded(mine), 1)

		assert.Equal(t, want, uids(got))
	})

	t.Run("combines status, filters and the connector filter with the cursor", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		create := func(n int, name string, status models.DeviceStatus, platform string) string {
			return createAt(t, minute(n),
				WithTenantID(tenantID),
				WithDeviceName(name),
				WithDeviceStatus(status),
				WithDevicePlatform(platform),
			)
		}

		web1 := create(0, "web-1", models.DeviceStatusAccepted, "docker")
		create(1, "db-1", models.DeviceStatusAccepted, "docker")
		create(2, "web-2", models.DeviceStatusPending, "docker")
		web3 := create(3, "web-3", models.DeviceStatusAccepted, "native")
		create(4, "web-4", models.DeviceStatusAccepted, "connector")
		web5 := create(5, "web-5", models.DeviceStatusAccepted, "docker")

		got := readAll(t, scope.MustBounded(tenantID), 1,
			st.Options().WithDeviceStatus(models.DeviceStatusAccepted),
			st.Options().Match(&query.Filters{Data: []query.Filter{{
				Type:   query.FilterTypeProperty,
				Params: &query.FilterProperty{Name: "name", Operator: "contains", Value: "web"},
			}}}),
			st.Options().Match(&query.Filters{Data: []query.Filter{{
				Type:   query.FilterTypeProperty,
				Params: &query.FilterProperty{Name: "platform", Operator: "ne", Value: "connector"},
			}}}),
		)

		assert.Equal(t, []string{web1, web3, web5}, uids(got))
	})

	t.Run("computes acceptable as the list does", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		s.CreateDevice(t, WithTenantID(tenantID), WithDeviceStatus(models.DeviceStatusPending))

		devices, _, err := st.DeviceListAfter(ctx, scope.MustBounded(tenantID), store.DeviceAcceptableAsFalse, nil, 10)
		require.NoError(t, err)
		require.Len(t, devices, 1)
		assert.False(t, devices[0].Acceptable)

		devices, _, err = st.DeviceListAfter(ctx, scope.MustBounded(tenantID), store.DeviceAcceptableIfNotAccepted, nil, 10)
		require.NoError(t, err)
		require.Len(t, devices, 1)
		assert.True(t, devices[0].Acceptable)
	})

	t.Run("returns the namespace and tags of each device", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t, WithNamespaceName("cursor-ns"))
		tag := s.CreateTag(t, WithTagName("production"), WithTagTenant(tenantID))
		uid := s.CreateDevice(t, WithTenantID(tenantID))
		require.NoError(t, st.TagPushToTarget(ctx, tag, store.TagTargetDevice, string(uid)))

		devices, _, err := st.DeviceListAfter(ctx, scope.MustBounded(tenantID), store.DeviceAcceptableIfNotAccepted, nil, 10)
		require.NoError(t, err)
		require.Len(t, devices, 1)
		assert.Equal(t, "cursor-ns", devices[0].Namespace)
		require.Len(t, devices[0].Tags, 1)
		assert.Equal(t, "production", devices[0].Tags[0].Name)
	})

	t.Run("returns a device created during the read once, and one changed during it once", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		sc := scope.MustBounded(tenantID)
		d0 := createAt(t, minute(0), WithTenantID(tenantID))
		d1 := createAt(t, minute(1), WithTenantID(tenantID))
		d2 := createAt(t, minute(2), WithTenantID(tenantID))

		first, next, err := st.DeviceListAfter(ctx, sc, store.DeviceAcceptableIfNotAccepted, nil, 2)
		require.NoError(t, err)
		require.Equal(t, []string{d0, d1}, uids(first))

		for _, uid := range []string{d0, d2} {
			device, err := st.DeviceResolve(ctx, sc, store.DeviceUIDResolver, uid)
			require.NoError(t, err)
			device.LastSeen = minute(60)
			device.Name += "-renamed"
			require.NoError(t, st.DeviceUpdate(ctx, device))
		}

		d3 := createAt(t, minute(3), WithTenantID(tenantID))

		var rest []models.Device
		for next != nil {
			var page []models.Device
			page, next, err = st.DeviceListAfter(ctx, sc, store.DeviceAcceptableIfNotAccepted, next, 2)
			require.NoError(t, err)
			rest = append(rest, page...)
		}

		assert.Equal(t, []string{d2, d3}, uids(rest))
	})

	t.Run("leaves a device created behind the cursor out of the pages after it", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		sc := scope.MustBounded(tenantID)
		d0 := createAt(t, minute(0), WithTenantID(tenantID))
		d1 := createAt(t, minute(1), WithTenantID(tenantID))
		d2 := createAt(t, minute(2), WithTenantID(tenantID))

		first, next, err := st.DeviceListAfter(ctx, sc, store.DeviceAcceptableIfNotAccepted, nil, 2)
		require.NoError(t, err)
		require.Equal(t, []string{d0, d1}, uids(first))

		behind := createAt(t, minute(0).Add(30*time.Second), WithTenantID(tenantID))

		rest, next, err := st.DeviceListAfter(ctx, sc, store.DeviceAcceptableIfNotAccepted, next, 2)
		require.NoError(t, err)
		assert.Equal(t, []string{d2}, uids(rest))
		assert.Nil(t, next)

		assert.Equal(t, []string{d0, behind, d1, d2}, uids(readAll(t, sc, 2)))
	})

	t.Run("keeps a device removed during the read with status removed", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		sc := scope.MustBounded(tenantID)
		devices := make([]string, 4)
		for i := range devices {
			devices[i] = createAt(t, minute(i), WithTenantID(tenantID))
		}

		accepted := st.Options().WithDeviceStatus(models.DeviceStatusAccepted)
		first, next, err := st.DeviceListAfter(ctx, sc, store.DeviceAcceptableIfNotAccepted, nil, 2, accepted)
		require.NoError(t, err)
		require.Equal(t, devices[:2], uids(first))

		removed, err := st.DeviceResolve(ctx, sc, store.DeviceUIDResolver, devices[2])
		require.NoError(t, err)
		removedAt := minute(60)
		removed.Status = models.DeviceStatusRemoved
		removed.RemovedAt = &removedAt
		require.NoError(t, st.DeviceUpdate(ctx, removed))

		rest, next, err := st.DeviceListAfter(ctx, sc, store.DeviceAcceptableIfNotAccepted, next, 2, accepted)
		require.NoError(t, err)
		assert.Equal(t, devices[3:], uids(rest))
		assert.Nil(t, next)

		unfiltered := readAll(t, sc, 10)
		require.Equal(t, devices, uids(unfiltered))
		assert.Equal(t, models.DeviceStatusRemoved, unfiltered[2].Status)
	})
}
