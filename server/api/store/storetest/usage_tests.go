package storetest

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type usageFixture struct {
	since   time.Time
	tenantA string
	tenantB string
}

func (s *Suite) seedUsage(t *testing.T) usageFixture {
	t.Helper()

	require.NoError(t, s.provider.CleanDatabase(t))

	now := s.Clock().Now()

	tenantA := s.CreateNamespace(t)
	tenantB := s.CreateNamespace(t)
	s.CreateMembership(t, tenantB, s.CreateUser(t), "observer")

	accepted := s.CreateDevice(t, WithTenantID(tenantA), WithDeviceStatus(models.DeviceStatusAccepted))
	s.CreateDevice(t, WithTenantID(tenantA), WithDeviceStatus(models.DeviceStatusAccepted))
	s.CreateDevice(t, WithTenantID(tenantA), WithDeviceStatus(models.DeviceStatusPending))
	s.CreateDevice(t, WithTenantID(tenantB), WithDeviceStatus(models.DeviceStatusRejected))

	clk := s.pinClock(t, now.Add(-48*time.Hour))
	s.CreateSession(t, WithSessionDevice(accepted), WithSessionActive(false))
	clk.Set(now)
	s.CreateSession(t, WithSessionDevice(accepted), WithSessionActive(false))

	return usageFixture{since: now.Add(-24 * time.Hour), tenantA: tenantA, tenantB: tenantB}
}

// TestInstanceUsage locks the instance-wide counts the telemetry heartbeat reports.
func (s *Suite) TestInstanceUsage(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("counts the whole instance, sessions only since the cutoff", func(t *testing.T) {
		fixture := s.seedUsage(t)

		usage, err := st.InstanceUsage(ctx, fixture.since)
		require.NoError(t, err)

		assert.Equal(t, &models.InstanceUsage{
			Namespaces: 2,
			Users:      3,
			Devices:    models.DeviceUsage{Accepted: 2, Pending: 1, Rejected: 1},
			Sessions:   1,
		}, usage)
	})

	t.Run("an empty instance counts zero", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		usage, err := st.InstanceUsage(ctx, s.Clock().Now())
		require.NoError(t, err)

		assert.Equal(t, &models.InstanceUsage{}, usage)
	})
}

// TestNamespaceUsages locks the per-namespace counts the cloud telemetry heartbeat reports.
func (s *Suite) TestNamespaceUsages(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("counts each namespace on its own", func(t *testing.T) {
		fixture := s.seedUsage(t)

		usages, err := st.NamespaceUsages(ctx, fixture.since)
		require.NoError(t, err)

		assert.ElementsMatch(t, []models.NamespaceUsage{
			{
				TenantID: fixture.tenantA,
				Members:  1,
				Devices:  models.DeviceUsage{Accepted: 2, Pending: 1},
				Sessions: 1,
			},
			{
				TenantID: fixture.tenantB,
				Members:  2,
				Devices:  models.DeviceUsage{Rejected: 1},
			},
		}, usages)
	})
}
