package services

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (e *enrollmentE2E) requireGone(t *testing.T, uid string) {
	t.Helper()

	_, err := e.st.DeviceResolve(context.Background(), scope.MustBounded(e.tenantID), store.DeviceUIDResolver, uid)
	require.ErrorIs(t, err, store.ErrNoDocuments)
}

func (e *enrollmentE2E) requireReplaced(t *testing.T, uid, replacedUID, replacedName string) {
	t.Helper()

	device := e.device(t, uid)
	require.Equal(t, replacedName, device.Name, "the newcomer takes over the replaced device's name")
	require.NotNil(t, device.Replaced, "the device records what it replaced")
	assert.Equal(t, replacedUID, device.Replaced.UID)
	assert.Equal(t, replacedName, device.Replaced.Name)
	assert.WithinDuration(t, e.clock.Now(), device.Replaced.At, time.Millisecond)
	e.requireGone(t, replacedUID)
}

func (e *enrollmentE2E) replacementEvents(t *testing.T, keyName, uid string) []models.ProvisioningKeyEvent {
	t.Helper()

	var replacements []models.ProvisioningKeyEvent
	for _, event := range e.events(t, keyName) {
		if event.DeviceUID == uid && event.Replaced != nil {
			replacements = append(replacements, event)
		}
	}

	return replacements
}

func (e *enrollmentE2E) requireEventReplaced(t *testing.T, keyName, uid, replacedUID, replacedName string) {
	t.Helper()

	replacements := e.replacementEvents(t, keyName, uid)
	require.Len(t, replacements, 1, "the replacement is an event of its own in the key's history")
	assert.Equal(t, replacedUID, replacements[0].Replaced.UID)
	assert.Equal(t, replacedName, replacements[0].Replaced.Name)
	assert.False(t, replacements[0].IsCurrent, "a replacement is never the device's current enrollment")
	assert.Empty(t, replacements[0].DecidedStatus, "a replacement is never stamped with a decision")

	for _, event := range e.events(t, keyName) {
		if event.DeviceUID == uid && event.IsCurrent {
			assert.Equal(t, models.DeviceStatusAccepted, event.DecidedStatus, "the decision lands on the enrollment, not on the replacement")

			return
		}
	}

	t.Fatalf("no current enrollment event for %s on %s", uid, keyName)
}

func TestDeviceReplaceE2E_ManualAccept(t *testing.T) { //nolint:tparallel // the subtests run in order against the one database this test sets up
	t.Parallel()
	ctx := context.Background()
	e := setupEnrollmentE2E(t)
	const mac = "aa:bb:cc:dd:ee:a1"

	e.provisioningKey(t, digest(0xa1), "manual-replace", models.ProvisioningKeyModeManual, models.ProvisioningKeyTypeUser, clearSecret)

	old := e.enrollAs(t, "field-box", mac, "pk-old", plaintextFor(0xa1))
	require.NoError(t, e.accept(ctx, old))
	e.disconnect(t, old)

	newcomer := e.enrollAs(t, "evilbox", mac, "pk-new", plaintextFor(0xa1))
	require.Equal(t, models.DeviceStatusPending, e.status(t, newcomer))

	t.Run("an accept that may not remove devices refuses to replace one", func(t *testing.T) {
		require.ErrorIs(t, e.acceptWithoutRemoval(ctx, newcomer), ErrDeviceReplaceForbidden)

		require.Equal(t, models.DeviceStatusPending, e.status(t, newcomer))
		require.Equal(t, models.DeviceStatusAccepted, e.status(t, old))
		require.Equal(t, "field-box", e.device(t, old).Name)
		require.Nil(t, e.device(t, newcomer).Replaced)
	})

	t.Run("an accept that does not confirm the replacement is refused naming the device", func(t *testing.T) {
		err := e.acceptUnconfirmed(ctx, newcomer)
		require.ErrorIs(t, err, ErrDeviceReplaceUnconfirmed)

		var refusal errors.Error
		require.True(t, errors.As(err, &refusal))
		assert.Equal(t, errors.InvalidFields{Fields: map[string]string{"replace": "field-box"}}, refusal.Data)

		require.Equal(t, models.DeviceStatusPending, e.status(t, newcomer))
		require.Equal(t, models.DeviceStatusAccepted, e.status(t, old))
	})

	t.Run("an accept that confirms the replacement replaces it and records it", func(t *testing.T) {
		e.clock.Advance(time.Minute)

		require.NoError(t, e.accept(ctx, newcomer))

		require.Equal(t, models.DeviceStatusAccepted, e.status(t, newcomer))
		e.requireReplaced(t, newcomer, old, "field-box")
		e.requireEventReplaced(t, "manual-replace", newcomer, old, "field-box")
	})
}

func TestDeviceReplaceE2E_RemovalKeepsTheHistory(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	e := setupEnrollmentE2E(t)
	const mac = "aa:bb:cc:dd:ee:a9"

	e.provisioningKey(t, digest(0xa9), "manual-history", models.ProvisioningKeyModeManual, models.ProvisioningKeyTypeUser, clearSecret)

	old := e.enrollAs(t, "field-box", mac, "pk-old", plaintextFor(0xa9))
	require.NoError(t, e.accept(ctx, old))
	e.disconnect(t, old)

	newcomer := e.enrollAs(t, "evilbox", mac, "pk-new", plaintextFor(0xa9))
	require.NoError(t, e.accept(ctx, newcomer))
	e.requireReplaced(t, newcomer, old, "field-box")

	require.NoError(t, e.svc.DeleteDevice(ctx, models.UID(newcomer), e.tenantID))

	revived := e.enrollAs(t, "evilbox", mac, "pk-new", plaintextFor(0xa9))
	require.Equal(t, newcomer, revived, "the same identity comes back as the same device")
	require.NoError(t, e.accept(ctx, revived))

	replaced := e.device(t, revived).Replaced
	require.NotNil(t, replaced, "the device's history still holds the replacement it made")
	assert.Equal(t, old, replaced.UID)
	require.Len(t, e.replacementEvents(t, "manual-history", revived), 1, "an accept with nothing to replace records nothing")
}

func TestDeviceReplaceE2E_AChainKeepsEveryReplacement(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	e := setupEnrollmentE2E(t)
	const mac = "aa:bb:cc:dd:ee:aa"

	e.provisioningKey(t, digest(0xaa), "manual-chain", models.ProvisioningKeyModeManual, models.ProvisioningKeyTypeUser, clearSecret)

	original := e.enrollAs(t, "field-box", mac, "pk-original", plaintextFor(0xaa))
	require.NoError(t, e.accept(ctx, original))
	e.disconnect(t, original)

	first := e.enrollAs(t, "evilbox", mac, "pk-first", plaintextFor(0xaa))
	require.NoError(t, e.accept(ctx, first))
	e.disconnect(t, first)

	second := e.enrollAs(t, "evilbox-again", mac, "pk-second", plaintextFor(0xaa))
	require.NoError(t, e.accept(ctx, second))

	e.requireEventReplaced(t, "manual-chain", first, original, "field-box")
	e.requireEventReplaced(t, "manual-chain", second, first, "field-box")
	e.requireReplaced(t, second, first, "field-box")

	e.clock.Advance(time.Minute)
	returned := e.enrollAs(t, "evilbox", mac, "pk-first", plaintextFor(0xaa))
	require.Equal(t, first, returned, "the replaced device's agent comes back under its old UID")
	require.Nil(t, e.device(t, returned).Replaced, "a device recreated after its record was deleted does not read the replacements of the record before it")
}

func TestDeviceReplaceE2E_ADeviceWithNoEnrollmentEventStillRecordsIt(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	e := setupEnrollmentE2E(t)
	const mac = "aa:bb:cc:dd:ee:ab"

	e.provisioningKey(t, digest(0xab), "manual-no-event", models.ProvisioningKeyModeManual, models.ProvisioningKeyTypeUser, clearSecret)

	old := e.enrollAs(t, "field-box", mac, "pk-old", plaintextFor(0xab))
	require.NoError(t, e.accept(ctx, old))
	e.disconnect(t, old)

	newcomer := e.enrollAs(t, "evilbox", mac, "pk-new", plaintextFor(0xab))
	_, err := e.db.ExecContext(ctx, "DELETE FROM provisioning_key_events WHERE device_uid = ?", newcomer)
	require.NoError(t, err)

	require.NoError(t, e.accept(ctx, newcomer))

	e.requireReplaced(t, newcomer, old, "field-box")
	replacements := e.replacementEvents(t, "manual-no-event", newcomer)
	require.Len(t, replacements, 1, "the replacement is recorded from the device itself")
	assert.Equal(t, "evilbox", replacements[0].Hostname, "the row names the newcomer as it enrolled, before it takes the old name")
}

func TestDeviceReplaceE2E_AcceptWithoutATwinRecordsNothing(t *testing.T) {
	t.Parallel()
	e := setupEnrollmentE2E(t)

	uid := e.acceptedDevice(t, "aa:bb:cc:dd:ee:a2")

	require.Equal(t, models.DeviceStatusAccepted, e.status(t, uid))
	require.Nil(t, e.device(t, uid).Replaced)
}

func TestDeviceReplaceE2E_ProvisioningKey(t *testing.T) {
	t.Parallel()

	t.Run("a key replaces an offline device by default and records it", func(t *testing.T) {
		t.Parallel()
		e := setupEnrollmentE2E(t)
		const mac = "aa:bb:cc:dd:ee:a3"

		e.provisioningKey(t, digest(0xa3), "auto-replace", models.ProvisioningKeyModeAutomatic, models.ProvisioningKeyTypeUser, clearSecret)

		old := e.enrollAs(t, "field-box", mac, "pk-old", plaintextFor(0xa3))
		e.disconnect(t, old)

		reflashed := e.enrollAs(t, "fresh-image", mac, "pk-reflashed", plaintextFor(0xa3))

		require.Equal(t, models.DeviceStatusAccepted, e.status(t, reflashed))
		e.requireReplaced(t, reflashed, old, "field-box")
		e.requireEventReplaced(t, "auto-replace", reflashed, old, "field-box")
	})

	t.Run("a key that does not replace offline devices leaves the newcomer pending", func(t *testing.T) {
		t.Parallel()
		e := setupEnrollmentE2E(t)
		const mac = "aa:bb:cc:dd:ee:a4"

		e.provisioningKey(t, digest(0xa4), "auto-keep", models.ProvisioningKeyModeAutomatic, models.ProvisioningKeyTypeUser, func(k *models.ProvisioningKey) {
			clearSecret(k)
			k.ReplaceOfflineDevices = false
		})

		old := e.enrollAs(t, "field-box", mac, "pk-old", plaintextFor(0xa4))
		require.Equal(t, models.DeviceStatusAccepted, e.status(t, old))
		e.disconnect(t, old)

		newcomer := e.enrollAs(t, "evilbox", mac, "pk-new", plaintextFor(0xa4))
		e.clock.Advance(models.EnrollmentReconcileInterval + time.Second)
		e.enrollAs(t, "evilbox", mac, "pk-new", plaintextFor(0xa4))

		require.Equal(t, models.DeviceStatusPending, e.status(t, newcomer))
		require.Equal(t, models.DeviceStatusAccepted, e.status(t, old))
		require.Equal(t, "field-box", e.device(t, old).Name)
		require.Equal(t, 1, e.usedTimes(t, digest(0xa4)), "a refused replacement spends no key use")

		t.Run("an administrator still replaces it by hand", func(t *testing.T) {
			require.NoError(t, e.accept(context.Background(), newcomer))

			e.requireReplaced(t, newcomer, old, "field-box")
		})
	})

	t.Run("the option round-trips through the key's create and update", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		e := setupEnrollmentE2E(t)
		replaces := func(name string) bool {
			key, err := e.st.ProvisioningKeyResolve(ctx, scope.MustBounded(e.tenantID), store.ProvisioningKeyNameResolver, name)
			require.NoError(t, err)

			return key.ReplaceOfflineDevices
		}
		off, on := false, true

		_, err := e.svc.CreateProvisioningKey(ctx, &requests.CreateProvisioningKey{TenantID: e.tenantID, Name: "omitted"})
		require.NoError(t, err)
		require.True(t, replaces("omitted"), "a key created without the option replaces offline devices")

		_, err = e.svc.CreateProvisioningKey(ctx, &requests.CreateProvisioningKey{TenantID: e.tenantID, Name: "off", ReplaceOfflineDevices: &off})
		require.NoError(t, err)
		require.False(t, replaces("off"))

		require.NoError(t, e.svc.UpdateProvisioningKey(ctx, &requests.UpdateProvisioningKey{
			TenantID: e.tenantID, CurrentName: "off", ReplaceOfflineDevices: &on,
		}))
		require.True(t, replaces("off"))
	})
}

func TestDeviceReplaceE2E_Pairing(t *testing.T) {
	t.Parallel()

	for _, mac := range []string{"aa:bb:cc:dd:ee:a5", ""} {
		t.Run("identity "+strconv.Quote(mac), func(t *testing.T) {
			t.Parallel()
			pairingReplacement(t, mac)
		})
	}
}

func pairingReplacement(t *testing.T, mac string) {
	t.Helper()
	ctx := context.Background()
	e := setupOwnershipE2E(t)

	operator := e.member(t, "operator", authorizer.RoleOperator)
	administrator := e.member(t, "administrator", authorizer.RoleAdministrator)

	old := e.pair(t, operator, mac, "pk-old").UID
	e.disconnect(t, old)

	pairing, err := e.svc.CreateDevicePairing(ctx, pairingRequest(mac, "pk-new"))
	require.NoError(t, err)

	accept := func(userID string, replace bool) (*models.DevicePairingAccepted, error) {
		return e.svc.AcceptDevicePairing(ctx, userID, &requests.DevicePairingAccept{Code: pairing.Code, TenantID: e.tenantID, Replace: replace})
	}

	t.Run("a member who may not remove devices cannot replace one", func(t *testing.T) {
		_, err := accept(operator, true)
		require.ErrorIs(t, err, ErrDeviceReplaceForbidden)

		require.Equal(t, models.DeviceStatusAccepted, e.status(t, old))
		require.Equal(t, int64(0), e.pendingCount(t), "the refusal comes before the device is created")
	})

	t.Run("an unconfirmed replacement names the device it would replace", func(t *testing.T) {
		_, err := accept(administrator, false)
		require.ErrorIs(t, err, ErrDeviceReplaceUnconfirmed)

		var refusal errors.Error
		require.True(t, errors.As(err, &refusal))
		assert.Equal(t, errors.InvalidFields{Fields: map[string]string{"replace": "host-" + mac}}, refusal.Data)
		require.Equal(t, models.DeviceStatusAccepted, e.status(t, old))
	})

	t.Run("a confirmed replacement by an administrator replaces it", func(t *testing.T) {
		accepted, err := accept(administrator, true)
		require.NoError(t, err)

		e.requireReplaced(t, accepted.UID, old, "host-"+mac)
		assert.Equal(t, administrator, accepted.OwnerID)
	})
}

func TestDeviceReplaceE2E_APairingCodeForADeviceAcceptedMeanwhileReplacesNothing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	e := setupOwnershipE2E(t)
	const mac = "aa:bb:cc:dd:ee:a7"
	operator := e.member(t, "operator", authorizer.RoleOperator)

	pairing, err := e.svc.CreateDevicePairing(ctx, pairingRequest(mac, "pk-same"))
	require.NoError(t, err)

	uid := e.enrollKey(t, mac, "pk-same")
	e.accept(t, uid)
	e.disconnect(t, uid)

	accepted, err := e.svc.AcceptDevicePairing(ctx, operator, &requests.DevicePairingAccept{Code: pairing.Code, TenantID: e.tenantID})
	require.NoError(t, err, "the code's device is the accepted one, not a twin to replace")
	assert.Equal(t, uid, accepted.UID)
}
