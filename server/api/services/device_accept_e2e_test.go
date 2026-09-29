package services

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/require"
)

func (e *enrollmentE2E) accept(ctx context.Context, uid string) error {
	return e.svc.UpdateDeviceStatus(ctx, &requests.DeviceUpdateStatus{
		TenantID: e.tenantID, UID: uid, Status: string(models.DeviceStatusAccepted),
	})
}

func (e *enrollmentE2E) acceptedDevice(t *testing.T, mac string) string {
	t.Helper()

	uid := e.enrollWithPublicKey(t, mac, "pk-old-"+mac, "")
	require.NoError(t, e.accept(context.Background(), uid))

	return uid
}

func (e *enrollmentE2E) heldOpen(t *testing.T, writes func(ctx context.Context) error) (commit func()) {
	t.Helper()

	written := make(chan error, 1)
	release := make(chan struct{})
	committed := make(chan error, 1)

	var releaseOnce sync.Once
	releaseWrites := func() { releaseOnce.Do(func() { close(release) }) }
	t.Cleanup(releaseWrites)

	go func() {
		committed <- e.st.WithTransaction(context.Background(), func(ctx context.Context) error {
			err := writes(ctx)
			written <- err
			if err != nil {
				return err
			}

			<-release

			return nil
		})
	}()

	require.NoError(t, <-written)

	return func() {
		releaseWrites()
		require.NoError(t, <-committed)
	}
}

func (e *enrollmentE2E) merge(ctx context.Context, sc scope.Scope, mac, oldUID, mergedUID string) error {
	if err := e.st.DeviceLockMAC(ctx, sc, mac); err != nil {
		return err
	}

	if err := e.remove(ctx, sc, oldUID); err != nil {
		return err
	}

	device, err := e.st.DeviceResolve(ctx, sc, store.DeviceUIDResolver, mergedUID)
	if err != nil {
		return err
	}

	device.Status = models.DeviceStatusAccepted

	return e.st.DeviceUpdate(ctx, device)
}

func (e *enrollmentE2E) remove(ctx context.Context, sc scope.Scope, uid string) error {
	device, err := e.st.DeviceResolve(ctx, sc, store.DeviceUIDResolver, uid)
	if err != nil {
		return err
	}

	return e.st.DeviceDelete(ctx, device)
}

func (e *enrollmentE2E) waitUntilBlockedOnLock(t *testing.T) {
	t.Helper()

	require.Eventually(t, func() bool {
		var waiting int
		err := e.db.QueryRowContext(context.Background(),
			"SELECT count(*) FROM pg_stat_activity WHERE datname = current_database() AND wait_event_type = 'Lock'",
		).Scan(&waiting)

		return err == nil && waiting > 0
	}, 10*time.Second, 10*time.Millisecond, "the accept never blocked on a lock held by the concurrent merge")
}

func (e *enrollmentE2E) acceptWhileHeldOpen(t *testing.T, writes func(ctx context.Context) error, acceptedUID string) error {
	t.Helper()

	commit := e.heldOpen(t, writes)

	loser := make(chan error, 1)
	go func() { loser <- e.accept(context.Background(), acceptedUID) }()

	e.waitUntilBlockedOnLock(t)
	commit()

	return <-loser
}

func (e *enrollmentE2E) acceptWhileMergeHeldOpen(t *testing.T, mac, oldUID, mergedUID, acceptedUID string) error {
	t.Helper()

	return e.acceptWhileHeldOpen(t, func(ctx context.Context) error {
		return e.merge(ctx, scope.MustBounded(e.tenantID), mac, oldUID, mergedUID)
	}, acceptedUID)
}

func TestDeviceAcceptE2E_ConcurrentAcceptOfTheSameDevice(t *testing.T) {
	e := setupEnrollmentE2E(t)
	const mac = "aa:bb:cc:dd:51:01"

	oldUID := e.acceptedDevice(t, mac)
	newUID := e.enrollWithPublicKey(t, mac, "pk-new", "")

	err := e.acceptWhileMergeHeldOpen(t, mac, oldUID, newUID, newUID)

	require.ErrorIs(t, err, ErrDeviceStatusAccepted)
	require.Equal(t, models.DeviceStatusAccepted, e.status(t, newUID))
}

func TestDeviceAcceptE2E_ConcurrentAcceptOfAnotherDeviceWithTheSameMACMergesIntoTheWinner(t *testing.T) {
	e := setupEnrollmentE2E(t)
	const mac = "aa:bb:cc:dd:51:02"

	oldUID := e.acceptedDevice(t, mac)
	mergedUID := e.enrollWithPublicKey(t, mac, "pk-merged", "")
	lateUID := e.enrollWithPublicKey(t, mac, "pk-late", "")

	err := e.acceptWhileMergeHeldOpen(t, mac, oldUID, mergedUID, lateUID)

	require.NoError(t, err)
	require.Equal(t, models.DeviceStatusAccepted, e.status(t, lateUID))

	_, err = e.st.DeviceResolve(context.Background(), scope.MustBounded(e.tenantID), store.DeviceUIDResolver, mergedUID)
	require.ErrorIs(t, err, store.ErrNoDocuments)
}

func TestDeviceAcceptE2E_AcceptWhileTheOldDeviceIsRemoved(t *testing.T) {
	e := setupEnrollmentE2E(t)
	const mac = "aa:bb:cc:dd:51:04"

	oldUID := e.acceptedDevice(t, mac)
	newUID := e.enrollWithPublicKey(t, mac, "pk-new", "")

	err := e.acceptWhileHeldOpen(t, func(ctx context.Context) error {
		return e.remove(ctx, scope.MustBounded(e.tenantID), oldUID)
	}, newUID)

	require.NoError(t, err)
	require.Equal(t, models.DeviceStatusAccepted, e.status(t, newUID))
}

func TestDeviceAcceptE2E_ManyConcurrentAcceptsOfTheSameDevice(t *testing.T) {
	e := setupEnrollmentE2E(t)
	const mac = "aa:bb:cc:dd:51:03"

	e.acceptedDevice(t, mac)
	newUID := e.enrollWithPublicKey(t, mac, "pk-new", "")

	const racers = 8
	start := make(chan struct{})
	errs := make([]error, racers)

	var wg sync.WaitGroup
	for i := range racers {
		wg.Go(func() {
			<-start
			errs[i] = e.accept(context.Background(), newUID)
		})
	}
	close(start)
	wg.Wait()

	succeeded := 0
	for _, err := range errs {
		if err == nil {
			succeeded++

			continue
		}

		require.ErrorIs(t, err, ErrDeviceStatusAccepted)
	}

	require.Equal(t, 1, succeeded)
	require.Equal(t, models.DeviceStatusAccepted, e.status(t, newUID))

	counts := e.deviceCounts(t)
	require.Equal(t, int64(1), counts.DevicesAcceptedCount)
	require.Equal(t, int64(0), counts.DevicesPendingCount)
}
