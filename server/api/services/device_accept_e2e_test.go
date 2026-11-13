package services

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
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

func (e *enrollmentE2E) connect(t *testing.T, uid string) {
	t.Helper()

	gone, err := e.st.DeviceHeartbeat(context.Background(), scope.MustBounded(e.tenantID), []store.DeviceBeat{{UID: uid, TenantID: e.tenantID, At: e.clock.Now()}}, 0)
	require.NoError(t, err)
	require.Empty(t, gone)
}

func (e *enrollmentE2E) disconnect(t *testing.T, uid string) {
	t.Helper()

	require.NoError(t, e.st.DeviceOffline(context.Background(), scope.MustBounded(e.tenantID), uid, now))
}

func (e *enrollmentE2E) device(t *testing.T, uid string) *models.Device {
	t.Helper()

	device, err := e.st.DeviceResolve(context.Background(), scope.MustBounded(e.tenantID), store.DeviceUIDResolver, uid)
	require.NoError(t, err)

	return device
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

	return e.st.DeviceUpdate(ctx, sc, device)
}

func (e *enrollmentE2E) remove(ctx context.Context, sc scope.Scope, uid string) error {
	device, err := e.st.DeviceResolve(ctx, sc, store.DeviceUIDResolver, uid)
	if err != nil {
		return err
	}

	return e.st.DeviceDelete(ctx, sc, device)
}

func (e *enrollmentE2E) waitUntilBlockedOnLock(t *testing.T) {
	t.Helper()

	require.Eventually(t, func() bool {
		var waiting int
		err := e.db.QueryRowContext(context.Background(),
			"SELECT count(*) FROM pg_stat_activity WHERE datname = current_database() AND wait_event_type = 'Lock'",
		).Scan(&waiting)

		return err == nil && waiting > 0
	}, 10*time.Second, 10*time.Millisecond, "the call never blocked on a lock held by the open transaction")
}

func (e *enrollmentE2E) whileHeldOpen(t *testing.T, writes func(ctx context.Context) error, call func() error) error {
	t.Helper()

	commit := e.heldOpen(t, writes)

	waiting := make(chan error, 1)
	go func() { waiting <- call() }()

	e.waitUntilBlockedOnLock(t)
	commit()

	return <-waiting
}

func (e *enrollmentE2E) acceptWhileHeldOpen(t *testing.T, writes func(ctx context.Context) error, acceptedUID string) error {
	t.Helper()

	return e.whileHeldOpen(t, writes, func() error { return e.accept(context.Background(), acceptedUID) })
}

func (e *enrollmentE2E) acceptWhileMergeHeldOpen(t *testing.T, mac, oldUID, mergedUID, acceptedUID string) error {
	t.Helper()

	return e.acceptWhileHeldOpen(t, func(ctx context.Context) error {
		return e.merge(ctx, scope.MustBounded(e.tenantID), mac, oldUID, mergedUID)
	}, acceptedUID)
}

func TestDeviceAcceptE2E_ConcurrentAcceptOfTheSameDevice(t *testing.T) {
	t.Parallel()
	e := setupLegacyEnrollmentE2E(t)
	const mac = "aa:bb:cc:dd:51:01"

	oldUID := e.acceptedDevice(t, mac)
	newUID := e.enrollWithPublicKey(t, mac, "pk-new", "")

	err := e.acceptWhileMergeHeldOpen(t, mac, oldUID, newUID, newUID)

	require.ErrorIs(t, err, ErrDeviceStatusAccepted)
	require.Equal(t, models.DeviceStatusAccepted, e.status(t, newUID))
}

func TestDeviceAcceptE2E_ConcurrentAcceptOfAnotherDeviceWithTheSameMACMergesIntoTheWinner(t *testing.T) {
	t.Parallel()
	e := setupLegacyEnrollmentE2E(t)
	const mac = "aa:bb:cc:dd:51:02"

	oldUID := e.acceptedDevice(t, mac)
	mergedUID := e.enrollWithPublicKey(t, mac, "pk-merged", "")
	lateUID := e.enrollWithPublicKey(t, mac, "pk-late", "")
	e.disconnect(t, mergedUID)

	err := e.acceptWhileMergeHeldOpen(t, mac, oldUID, mergedUID, lateUID)

	require.NoError(t, err)
	require.Equal(t, models.DeviceStatusAccepted, e.status(t, lateUID))

	_, err = e.st.DeviceResolve(context.Background(), scope.MustBounded(e.tenantID), store.DeviceUIDResolver, mergedUID)
	require.ErrorIs(t, err, store.ErrNoDocuments)
}

func TestDeviceAcceptE2E_AcceptWhileTheOldDeviceIsRemoved(t *testing.T) {
	t.Parallel()
	e := setupLegacyEnrollmentE2E(t)
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
	t.Parallel()
	e := setupLegacyEnrollmentE2E(t)
	const mac = "aa:bb:cc:dd:51:03"

	e.disconnect(t, e.acceptedDevice(t, mac))
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

func TestDeviceAcceptE2E_RefusesToMergeAwayAConnectedDevice(t *testing.T) {
	t.Parallel()
	e := setupLegacyEnrollmentE2E(t)
	const mac = "aa:bb:cc:dd:51:05"

	oldUID := e.acceptedDevice(t, mac)
	e.connect(t, oldUID)
	oldName := e.device(t, oldUID).Name
	newUID := e.enrollWithPublicKey(t, mac, "pk-new", "")

	err := e.accept(context.Background(), newUID)

	require.ErrorIs(t, err, ErrDeviceMACConnected)
	require.Equal(t, models.DeviceStatusPending, e.status(t, newUID))
	require.Equal(t, models.DeviceStatusAccepted, e.status(t, oldUID))
	require.Equal(t, oldName, e.device(t, oldUID).Name)
}

func TestDeviceAcceptE2E_MergesAwayADisconnectedDevice(t *testing.T) {
	t.Parallel()
	e := setupLegacyEnrollmentE2E(t)
	const mac = "aa:bb:cc:dd:51:06"

	oldUID := e.acceptedDevice(t, mac)
	oldName := e.device(t, oldUID).Name
	newUID := e.enrollWithPublicKey(t, mac, "pk-new", "")
	e.disconnect(t, oldUID)

	require.NoError(t, e.accept(context.Background(), newUID))

	require.Equal(t, models.DeviceStatusAccepted, e.status(t, newUID))
	require.Equal(t, oldName, e.device(t, newUID).Name)

	_, err := e.st.DeviceResolve(context.Background(), scope.MustBounded(e.tenantID), store.DeviceUIDResolver, oldUID)
	require.ErrorIs(t, err, store.ErrNoDocuments)
}

const licenseDeviceCount = "test: the license counts devices across every namespace"

type registeredDeviceLicense struct {
	st    store.Store
	limit int64
}

func (l *registeredDeviceLicense) CanAcceptDevice(ctx context.Context) (bool, error) {
	registered, err := l.st.CountRegisteredDevices(ctx, scope.NewUnbounded(licenseDeviceCount))

	return l.limit > registered, err
}

func (l *registeredDeviceLicense) CanConnectDevice(ctx context.Context) (bool, error) {
	registered, err := l.st.CountRegisteredDevices(ctx, scope.NewUnbounded(licenseDeviceCount))

	return l.limit >= registered, err
}

var errLicenseUnreadable = errors.New("license unreadable")

type connectionSource interface {
	GetConnection(ctx context.Context) bun.IDB
}

type unreadableLicense struct {
	db connectionSource
}

func (l *unreadableLicense) readFrom(t *testing.T, st store.Store) {
	t.Helper()

	db, ok := st.(connectionSource)
	require.True(t, ok, "the store must expose the connection a transaction travels on")
	l.db = db
}

func (l *unreadableLicense) CanAcceptDevice(ctx context.Context) (bool, error) {
	if _, err := l.db.GetConnection(ctx).ExecContext(ctx, "SELECT id FROM a_license_table_that_does_not_exist"); err == nil {
		return false, errors.New("the failing license query succeeded")
	}

	return false, errLicenseUnreadable
}

func (l *unreadableLicense) CanConnectDevice(context.Context) (bool, error) {
	return false, errLicenseUnreadable
}

func (e *enrollmentE2E) secondNamespace(t *testing.T) *enrollmentE2E {
	t.Helper()

	second := *e
	second.tenantID = createOwnedNamespace(t, e.st, "second", "e2e-second")

	return &second
}

func (e *enrollmentE2E) legacySecondNamespace(t *testing.T) *enrollmentE2E {
	t.Helper()

	second := e.secondNamespace(t)
	second.legacy(t)

	return second
}

func (e *enrollmentE2E) limitDevices(t *testing.T, limit int) {
	t.Helper()

	_, err := e.db.ExecContext(context.Background(), "UPDATE namespaces SET max_devices = ? WHERE id = ?", limit, e.tenantID)
	require.NoError(t, err)
}

func (e *enrollmentE2E) acceptedCount(t *testing.T) int {
	t.Helper()

	var accepted int
	require.NoError(t, e.db.NewRaw("SELECT count(*) FROM devices WHERE namespace_id = ? AND status = 'accepted'", e.tenantID).
		Scan(context.Background(), &accepted))

	return accepted
}

func allAtOnce(calls ...func() error) []error {
	start := make(chan struct{})
	errs := make([]error, len(calls))

	var wg sync.WaitGroup
	for i, call := range calls {
		wg.Go(func() {
			<-start
			errs[i] = call()
		})
	}
	close(start)
	wg.Wait()

	return errs
}

func acceptAllAtOnce(t *testing.T, macPrefix string, racers int, namespaces ...*enrollmentE2E) []error {
	t.Helper()

	accepts := make([]func() error, racers)
	for i := range racers {
		namespace := namespaces[i%len(namespaces)]
		uid := namespace.enroll(t, fmt.Sprintf("%s:%02x", macPrefix, i), "")
		accepts[i] = func() error { return namespace.accept(context.Background(), uid) }
	}

	return allAtOnce(accepts...)
}

func withinDeadline(t *testing.T, deadline time.Duration, calls func() []error) []error {
	t.Helper()

	finished := make(chan []error, 1)
	go func() { finished <- calls() }()

	select {
	case errs := <-finished:
		return errs
	case <-time.After(deadline):
		require.FailNow(t, "the calls were still blocked when the deadline passed")

		return nil
	}
}

func requireRefusedBeyond(t *testing.T, limit int, errs []error, refusal error) {
	t.Helper()

	succeeded := 0
	for _, err := range errs {
		if err == nil {
			succeeded++

			continue
		}

		require.ErrorIs(t, err, refusal)
	}

	require.Equal(t, limit, succeeded)
}

func TestDeviceAcceptE2E_ConcurrentAcceptsStayWithinMaxDevices(t *testing.T) {
	t.Parallel()
	e := setupLegacyEnrollmentE2E(t)
	e.limitDevices(t, 1)

	errs := acceptAllAtOnce(t, "aa:bb:cc:dd:52", 8, e)

	requireRefusedBeyond(t, 1, errs, ErrMaxDeviceCountReached)
	require.Equal(t, 1, e.acceptedCount(t))
	require.Equal(t, int64(1), e.deviceCounts(t).DevicesAcceptedCount)
}

func TestDeviceAcceptE2E_ConcurrentAutomaticEnrollmentsStayWithinMaxDevices(t *testing.T) {
	t.Parallel()
	e := setupEnrollmentE2E(t)
	e.limitDevices(t, 1)
	e.provisioningKey(t, digest(0x54), "fleet", models.ProvisioningKeyModeAutomatic, models.ProvisioningKeyTypeUser, clearSecret)

	enrollments := make([]func() error, 8)
	for i := range enrollments {
		enrollments[i] = func() error {
			_, err := e.svc.AuthDevice(context.Background(), e.authRequest(fmt.Sprintf("aa:bb:cc:dd:54:%02x", i), plaintextFor(0x54)), "203.0.113.7")

			return err
		}
	}

	for _, err := range allAtOnce(enrollments...) {
		require.NoError(t, err)
	}

	require.Equal(t, 1, e.acceptedCount(t))
	require.Equal(t, int64(1), e.deviceCounts(t).DevicesAcceptedCount)
	require.Equal(t, int64(7), e.deviceCounts(t).DevicesPendingCount)
}

func TestDeviceAcceptE2E_ConcurrentAcceptsAcrossNamespacesStayWithinTheLicense(t *testing.T) {
	t.Parallel()
	license := &registeredDeviceLicense{limit: 2}
	e := setupEnrollmentE2EWith(t, WithEdition(envs.Enterprise), WithLicenseEvaluator(license))
	e.legacy(t)
	license.st = e.st

	errs := acceptAllAtOnce(t, "aa:bb:cc:dd:53", 8, e, e.legacySecondNamespace(t))

	requireRefusedBeyond(t, 2, errs, ErrDeviceLicenseLimit)

	registered, err := e.st.CountRegisteredDevices(context.Background(), scope.NewUnbounded(licenseDeviceCount))
	require.NoError(t, err)
	require.Equal(t, int64(2), registered)

	require.NoError(t, e.svc.EvaluateLicense(context.Background()))
}

func TestDeviceAcceptE2E_ConcurrentAutomaticEnrollmentsAcrossNamespacesStayWithinTheLicense(t *testing.T) {
	t.Parallel()
	license := &registeredDeviceLicense{limit: 2}
	e := setupEnrollmentE2EWith(t, WithEdition(envs.Enterprise), WithLicenseEvaluator(license))
	license.st = e.st
	second := e.secondNamespace(t)

	namespaces := []*enrollmentE2E{e, second}
	for i, namespace := range namespaces {
		namespace.provisioningKey(t, digest(byte(0x60+i)), "fleet", models.ProvisioningKeyModeAutomatic, models.ProvisioningKeyTypeUser, clearSecret)
	}

	enrollments := make([]func() error, 8)
	for i := range enrollments {
		namespace := namespaces[i%2]
		request := namespace.authRequest(fmt.Sprintf("aa:bb:cc:dd:60:%02x", i), plaintextFor(byte(0x60+i%2)))
		enrollments[i] = func() error {
			_, err := namespace.svc.AuthDevice(context.Background(), request, "203.0.113.7")

			return err
		}
	}

	for _, err := range allAtOnce(enrollments...) {
		require.NoError(t, err)
	}

	registered, err := e.st.CountRegisteredDevices(context.Background(), scope.NewUnbounded(licenseDeviceCount))
	require.NoError(t, err)
	require.Equal(t, int64(2), registered)
	require.Equal(t, int64(6), e.deviceCounts(t).DevicesPendingCount+second.deviceCounts(t).DevicesPendingCount)
	require.NoError(t, e.svc.EvaluateLicense(context.Background()))
}

func TestDeviceAcceptE2E_UnreadableLicense(t *testing.T) {
	t.Parallel()

	t.Run("fails open outside cloud", func(t *testing.T) {
		t.Parallel()
		license := &unreadableLicense{}
		e := setupEnrollmentE2EWith(t, WithEdition(envs.Enterprise), WithLicenseEvaluator(license))
		e.legacy(t)
		license.readFrom(t, e.st)
		uid := e.enroll(t, "aa:bb:cc:dd:55:01", "")

		require.NoError(t, e.accept(context.Background(), uid))
		require.Equal(t, models.DeviceStatusAccepted, e.status(t, uid))
	})

	t.Run("refuses on cloud", func(t *testing.T) {
		t.Parallel()
		license := &unreadableLicense{}
		e := setupEnrollmentE2EWith(t, WithEdition(envs.Cloud), WithLicenseEvaluator(license))
		e.legacy(t)
		license.readFrom(t, e.st)
		uid := e.enroll(t, "aa:bb:cc:dd:55:02", "")

		require.ErrorIs(t, e.accept(context.Background(), uid), errLicenseUnreadable)
		require.Equal(t, models.DeviceStatusPending, e.status(t, uid))
	})
}

func TestDeviceAcceptE2E_AcceptanceLockIsPerNamespace(t *testing.T) {
	t.Parallel()
	e := setupLegacyEnrollmentE2E(t)
	second := e.legacySecondNamespace(t)
	sameNamespaceUID := e.enroll(t, "aa:bb:cc:dd:56:01", "")
	otherNamespaceUID := second.enroll(t, "aa:bb:cc:dd:56:02", "")

	holdNamespaceLock := func(ctx context.Context) error {
		return e.svc.lockDeviceAcceptance(ctx, e.tenantID)
	}

	commit := e.heldOpen(t, holdNamespaceLock)
	errs := withinDeadline(t, 10*time.Second, func() []error {
		return []error{second.accept(context.Background(), otherNamespaceUID)}
	})
	commit()
	require.NoError(t, errs[0])

	require.NoError(t, e.acceptWhileHeldOpen(t, holdNamespaceLock, sameNamespaceUID))
}

func TestDeviceAcceptE2E_MixedBurstDoesNotDeadlock(t *testing.T) {
	t.Parallel()
	license := &registeredDeviceLicense{limit: 1000}
	e := setupEnrollmentE2EWith(t, WithEdition(envs.Enterprise), WithLicenseEvaluator(license))
	e.legacy(t)
	license.st = e.st
	second := e.legacySecondNamespace(t)

	var calls []func() error
	for i := range 8 {
		namespace := []*enrollmentE2E{e, second}[i%2]
		mac := fmt.Sprintf("aa:bb:cc:dd:57:%02x", i)
		replaced := namespace.acceptedDevice(t, mac)
		namespace.disconnect(t, replaced)
		replacement := namespace.enrollWithPublicKey(t, mac, "pk-new-"+mac, "")
		removed := namespace.acceptedDevice(t, fmt.Sprintf("aa:bb:cc:dd:58:%02x", i))
		fresh := namespace.enroll(t, fmt.Sprintf("aa:bb:cc:dd:59:%02x", i), "")

		calls = append(calls,
			func() error { return namespace.accept(context.Background(), replacement) },
			func() error {
				return namespace.svc.DeleteDevice(context.Background(), models.UID(removed), namespace.tenantID)
			},
			func() error { return namespace.accept(context.Background(), fresh) },
		)
	}

	for _, err := range withinDeadline(t, 30*time.Second, func() []error { return allAtOnce(calls...) }) {
		require.NoError(t, err)
	}

	for _, namespace := range []*enrollmentE2E{e, second} {
		require.Equal(t, 8, namespace.acceptedCount(t))
		require.Equal(t, int64(8), namespace.deviceCounts(t).DevicesAcceptedCount)
	}
}
