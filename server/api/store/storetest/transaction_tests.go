package storetest

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWithTransaction locks that a failing callback rolls back every write it made, and that a
// successful one commits them together.
func (s *Suite) TestWithTransaction(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("rolls back on error", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t, WithNamespaceName("tx-rollback"))

		errIntentional := errors.New("intentional error")
		err := st.WithTransaction(ctx, func(txCtx context.Context) error {
			device := &models.Device{
				UID:       uniqueHex(t, 64),
				Name:      "tx-device",
				TenantID:  tenantID,
				Identity:  &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:f1"},
				Info:      &models.DeviceInfo{},
				PublicKey: "-",
				Status:    models.DeviceStatusAccepted,
				CreatedAt: clock.Now(),
				LastSeen:  clock.Now(),
			}

			_, err := st.DeviceCreate(txCtx, device)
			if err != nil {
				return err
			}

			return errIntentional
		})
		require.ErrorIs(t, err, errIntentional)

		devices, count, err := st.DeviceList(ctx, scope.NewUnbounded(reasonTestQueryMechanics), store.DeviceAcceptableIfNotAccepted)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
		assert.Empty(t, devices)
	})

	t.Run("commits on success", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t, WithNamespaceName("tx-commit"))

		err := st.WithTransaction(ctx, func(txCtx context.Context) error {
			device := &models.Device{
				UID:       uniqueHex(t, 64),
				Name:      "tx-device",
				TenantID:  tenantID,
				Identity:  &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:f2"},
				Info:      &models.DeviceInfo{},
				PublicKey: "-",
				Status:    models.DeviceStatusAccepted,
				CreatedAt: clock.Now(),
				LastSeen:  clock.Now(),
			}

			_, err := st.DeviceCreate(txCtx, device)

			return err
		})
		require.NoError(t, err)

		devices, count, err := st.DeviceList(ctx, scope.NewUnbounded(reasonTestQueryMechanics), store.DeviceAcceptableIfNotAccepted)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		assert.Len(t, devices, 1)
	})
}
