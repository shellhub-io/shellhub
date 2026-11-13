package storetest

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
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
				CreatedAt: s.Clock().Now(),
				LastSeen:  s.Clock().Now(),
			}

			_, err := st.DeviceCreate(txCtx, scope.MustBounded(device.TenantID), device)
			if err != nil {
				return err
			}

			return errIntentional
		})
		require.ErrorIs(t, err, errIntentional)

		devices, count, err := st.DeviceList(ctx, scope.NewUnbounded(reasonTestQueryMechanics), store.DeviceAcceptableIfNotAccepted)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
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
				CreatedAt: s.Clock().Now(),
				LastSeen:  s.Clock().Now(),
			}

			_, err := st.DeviceCreate(txCtx, scope.MustBounded(device.TenantID), device)

			return err
		})
		require.NoError(t, err)

		devices, count, err := st.DeviceList(ctx, scope.NewUnbounded(reasonTestQueryMechanics), store.DeviceAcceptableIfNotAccepted)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
		assert.Len(t, devices, 1)
	})
}

func (s *Suite) pendingDevice(t *testing.T, tenantID, mac string) *models.Device {
	t.Helper()

	return &models.Device{
		UID:       uniqueHex(t, 64),
		Name:      "tx-" + uniqueHex(t, 8),
		TenantID:  tenantID,
		Identity:  &models.DeviceIdentity{MAC: mac},
		Info:      &models.DeviceInfo{},
		PublicKey: "-",
		Status:    models.DeviceStatusPending,
		CreatedAt: s.Clock().Now(),
		LastSeen:  s.Clock().Now(),
	}
}

func (s *Suite) devicesIn(t *testing.T) int64 {
	t.Helper()

	_, count, err := s.provider.Store().DeviceList(context.Background(), scope.NewUnbounded(reasonTestQueryMechanics), store.DeviceAcceptableIfNotAccepted)
	require.NoError(t, err)

	return count
}

// TestWithSavepoint locks that a statement failing inside a savepoint rolls back only the
// savepoint, leaving the enclosing transaction able to write and commit, and that a savepoint whose
// callback succeeds commits with that transaction.
func (s *Suite) TestWithSavepoint(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("a failed statement rolls back only the savepoint", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t, WithNamespaceName("savepoint-failure"))
		sc := scope.MustBounded(tenantID)

		var duplicate error
		err := st.WithTransaction(ctx, func(txCtx context.Context) error {
			first := s.pendingDevice(t, tenantID, "aa:bb:cc:dd:ee:f3")
			if _, err := st.DeviceCreate(txCtx, sc, first); err != nil {
				return err
			}

			duplicate = st.WithSavepoint(txCtx, func(savepointCtx context.Context) error {
				_, err := st.DeviceCreate(savepointCtx, sc, first)

				return err
			})

			_, err := st.DeviceCreate(txCtx, sc, s.pendingDevice(t, tenantID, "aa:bb:cc:dd:ee:f4"))

			return err
		})
		require.ErrorIs(t, duplicate, store.ErrDuplicate)
		require.NoError(t, err)

		assert.Equal(t, int64(2), s.devicesIn(t))
	})

	t.Run("a successful savepoint commits with the transaction", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t, WithNamespaceName("savepoint-success"))
		sc := scope.MustBounded(tenantID)

		err := st.WithTransaction(ctx, func(txCtx context.Context) error {
			return st.WithSavepoint(txCtx, func(savepointCtx context.Context) error {
				_, err := st.DeviceCreate(savepointCtx, sc, s.pendingDevice(t, tenantID, "aa:bb:cc:dd:ee:f5"))

				return err
			})
		})
		require.NoError(t, err)

		assert.Equal(t, int64(1), s.devicesIn(t))
	})

	t.Run("outside a transaction it commits on its own", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t, WithNamespaceName("savepoint-alone"))

		err := st.WithSavepoint(ctx, func(savepointCtx context.Context) error {
			_, err := st.DeviceCreate(savepointCtx, scope.MustBounded(tenantID), s.pendingDevice(t, tenantID, "aa:bb:cc:dd:ee:f6"))

			return err
		})
		require.NoError(t, err)

		assert.Equal(t, int64(1), s.devicesIn(t))
	})
}

func (s *Suite) holdAcceptanceLock(t *testing.T, sc scope.Scope) {
	t.Helper()

	held := make(chan error, 1)
	release := make(chan struct{})
	committed := make(chan error, 1)

	go func() {
		committed <- s.provider.Store().WithTransaction(context.Background(), func(txCtx context.Context) error {
			err := s.provider.Store().DeviceLockAcceptance(txCtx, sc)
			held <- err
			<-release

			return err
		})
	}()

	require.NoError(t, <-held)
	t.Cleanup(func() {
		close(release)
		require.NoError(t, <-committed)
	})
}

const acceptanceLockWait = 500 * time.Millisecond

func (s *Suite) lockAcceptanceWithin(sc scope.Scope) error {
	ctx, cancel := context.WithTimeout(context.Background(), acceptanceLockWait)
	defer cancel()

	return s.provider.Store().WithTransaction(ctx, func(txCtx context.Context) error {
		return s.provider.Store().DeviceLockAcceptance(txCtx, sc)
	})
}

// TestDeviceLockAcceptance locks that an acceptance lock waits only for a holder of the same lock:
// a namespace's lock for that namespace's holder, and the instance-wide lock for its own. It also
// locks the errors returned outside a transaction and for a zero scope.
func (s *Suite) TestDeviceLockAcceptance(t *testing.T) {
	st := s.provider.Store()

	tenant := scope.MustBounded("00000000-0000-4000-8000-000000000001")
	otherTenant := scope.MustBounded("00000000-0000-4000-8000-000000000002")
	instance := scope.NewUnbounded("test: the instance-wide acceptance lock")

	t.Run("refuses outside a transaction", func(t *testing.T) {
		require.ErrorIs(t, st.DeviceLockAcceptance(context.Background(), tenant), store.ErrLockOutsideTransaction)
	})

	t.Run("refuses a zero scope", func(t *testing.T) {
		require.ErrorIs(t, s.lockAcceptanceWithin(scope.Scope{}), store.ErrInvalidScope)
	})

	t.Run("a namespace lock waits only for its own namespace", func(t *testing.T) {
		s.holdAcceptanceLock(t, tenant)

		require.ErrorIs(t, s.lockAcceptanceWithin(tenant), context.DeadlineExceeded)
		require.NoError(t, s.lockAcceptanceWithin(otherTenant))
		require.NoError(t, s.lockAcceptanceWithin(instance))
	})

	t.Run("the instance lock waits only for the instance lock", func(t *testing.T) {
		s.holdAcceptanceLock(t, instance)

		require.ErrorIs(t, s.lockAcceptanceWithin(instance), context.DeadlineExceeded)
		require.NoError(t, s.lockAcceptanceWithin(tenant))
	})
}
