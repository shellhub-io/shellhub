package storetest

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *Suite) TestWithTransaction(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("rolls back on error", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t, WithNamespaceName("tx-rollback"))

		errIntentional := errors.New("intentional error")
		err := st.WithTransaction(ctx, func(txCtx context.Context) error {
			device := &models.Device{
				UID:       fmt.Sprintf("%064x", time.Now().UnixNano()),
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
		assert.ErrorIs(t, err, errIntentional)

		// Verify the device was NOT persisted (rolled back)
		devices, count, err := st.DeviceList(ctx, store.DeviceAcceptableIfNotAccepted)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
		assert.Empty(t, devices)
	})

	t.Run("commits on success", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t, WithNamespaceName("tx-commit"))

		err := st.WithTransaction(ctx, func(txCtx context.Context) error {
			device := &models.Device{
				UID:       fmt.Sprintf("%064x", time.Now().UnixNano()),
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

		// Verify the device WAS persisted
		devices, count, err := st.DeviceList(ctx, store.DeviceAcceptableIfNotAccepted)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		assert.Len(t, devices, 1)
	})
}
