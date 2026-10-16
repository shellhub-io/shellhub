package services

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/require"
)

func (e *enrollmentE2E) legacy(t *testing.T) {
	t.Helper()

	_, err := e.db.NewRaw("UPDATE namespaces SET legacy = true WHERE id = ?", e.tenantID).Exec(context.Background())
	require.NoError(t, err)
}

func (e *enrollmentE2E) deviceCount(t *testing.T) int64 {
	t.Helper()

	var count int64
	require.NoError(t, e.db.NewRaw("SELECT count(*) FROM devices WHERE namespace_id = ?", e.tenantID).Scan(context.Background(), &count))

	return count
}

func (e *enrollmentE2E) authRequest(mac, provisioningKey string) requests.DeviceAuth {
	return requests.DeviceAuth{
		TenantID:        e.tenantID,
		Hostname:        "host-" + mac,
		Identity:        &requests.DeviceIdentity{MAC: mac},
		Info:            &requests.DeviceInfo{ID: "debian", PrettyName: "Debian", Version: "v0.1.0", Arch: "amd64", Platform: "docker"},
		PublicKey:       "pk-" + mac,
		ProvisioningKey: provisioningKey,
	}
}

func TestAuthDeviceMode(t *testing.T) {
	ctx := context.Background()

	t.Run("new namespace refuses a bare tenant for an unknown device with 403 and creates nothing", func(t *testing.T) {
		e := setupEnrollmentE2E(t)

		_, err := e.svc.AuthDevice(ctx, e.authRequest("aa:bb:cc:dd:ee:01", ""))
		require.ErrorIs(t, err, ErrAuthNamespaceRequiresCredential)
		require.Equal(t, int64(0), e.deviceCount(t))
	})

	t.Run("new namespace enrolls by provisioning key", func(t *testing.T) {
		e := setupEnrollmentE2E(t)
		e.provisioningKey(t, digest(0x77), "manual", models.ProvisioningKeyModeManual, models.ProvisioningKeyTypeUser, func(k *models.ProvisioningKey) { k.KeyEncrypted, k.KeyHint = "", "" })

		uid := e.enroll(t, "aa:bb:cc:dd:ee:02", plaintextFor(0x77))
		require.Equal(t, models.DeviceStatusPending, e.status(t, uid))
	})

	t.Run("legacy namespace lands a bare tenant pending under the built-in legacy key", func(t *testing.T) {
		e := setupLegacyEnrollmentE2E(t)

		uid := e.enroll(t, "aa:bb:cc:dd:ee:03", "")
		require.Equal(t, models.DeviceStatusPending, e.status(t, uid))

		events := e.events(t, "legacy")
		require.Len(t, events, 1)
		require.Equal(t, uid, events[0].DeviceUID)
	})

	t.Run("legacy namespace refuses a presented provisioning key with 403 and creates nothing", func(t *testing.T) {
		e := setupEnrollmentE2E(t)
		e.provisioningKey(t, digest(0x77), "manual", models.ProvisioningKeyModeManual, models.ProvisioningKeyTypeUser, func(k *models.ProvisioningKey) { k.KeyEncrypted, k.KeyHint = "", "" })
		e.legacy(t)

		_, err := e.svc.AuthDevice(ctx, e.authRequest("aa:bb:cc:dd:ee:04", plaintextFor(0x77)))
		require.ErrorIs(t, err, ErrAuthNamespaceLegacy)
		require.Equal(t, int64(0), e.deviceCount(t))
	})

	t.Run("legacy namespace refuses a pairing at the accept step", func(t *testing.T) {
		e := setupOwnershipE2E(t)
		e.legacy(t)

		pairing, err := e.svc.CreateDevicePairing(ctx, pairingRequest("aa:bb:cc:dd:ee:05", "pk-05"))
		require.NoError(t, err)

		_, err = e.svc.AcceptDevicePairing(ctx, e.ownerID, &requests.DevicePairingAccept{Code: pairing.Code, TenantID: e.tenantID})
		require.ErrorIs(t, err, ErrNamespaceLegacy)
		require.Equal(t, int64(0), e.deviceCount(t))
	})

	t.Run("a device accepted by tenant before the upgrade re-authenticates after it", func(t *testing.T) {
		e := setupLegacyEnrollmentE2E(t)

		uid := e.enroll(t, "aa:bb:cc:dd:ee:06", "")
		require.NoError(t, e.svc.UpdateDeviceStatus(ctx, &requests.DeviceUpdateStatus{TenantID: e.tenantID, UID: uid, Status: string(models.DeviceStatusAccepted)}))

		require.NoError(t, e.svc.UpgradeNamespace(ctx, e.tenantID))

		res, err := e.svc.AuthDevice(ctx, e.authRequest("aa:bb:cc:dd:ee:06", ""))
		require.NoError(t, err)
		require.Equal(t, uid, res.UID)
		require.Equal(t, models.DeviceStatusAccepted, e.status(t, uid))
	})

	t.Run("a device pending by tenant before the upgrade stays pending and acceptable after it", func(t *testing.T) {
		e := setupLegacyEnrollmentE2E(t)

		uid := e.enroll(t, "aa:bb:cc:dd:ee:07", "")
		require.NoError(t, e.svc.UpgradeNamespace(ctx, e.tenantID))

		res, err := e.svc.AuthDevice(ctx, e.authRequest("aa:bb:cc:dd:ee:07", ""))
		require.NoError(t, err)
		require.Equal(t, uid, res.UID)
		require.Equal(t, models.DeviceStatusPending, e.status(t, uid))

		require.NoError(t, e.svc.UpdateDeviceStatus(ctx, &requests.DeviceUpdateStatus{TenantID: e.tenantID, UID: uid, Status: string(models.DeviceStatusAccepted)}))
		require.Equal(t, models.DeviceStatusAccepted, e.status(t, uid))
	})

	t.Run("a paired device re-authenticates with the tenant it persisted", func(t *testing.T) {
		e := setupOwnershipE2E(t)

		accepted := e.pair(t, e.ownerID, "aa:bb:cc:dd:ee:08", "pk-08")

		req := e.enrollmentE2E.authRequest("aa:bb:cc:dd:ee:08", "")
		req.PublicKey = "pk-08"
		res, err := e.svc.AuthDevice(ctx, req)
		require.NoError(t, err)
		require.Equal(t, accepted.UID, res.UID)
	})

	t.Run("a removed device returning by tenant alone to a new namespace is refused", func(t *testing.T) {
		e := setupLegacyEnrollmentE2E(t)

		uid := e.enroll(t, "aa:bb:cc:dd:ee:09", "")
		require.NoError(t, e.svc.UpdateDeviceStatus(ctx, &requests.DeviceUpdateStatus{TenantID: e.tenantID, UID: uid, Status: string(models.DeviceStatusAccepted)}))
		require.NoError(t, e.svc.DeleteDevice(ctx, models.UID(uid), e.tenantID))

		require.NoError(t, e.svc.UpgradeNamespace(ctx, e.tenantID))

		_, err := e.svc.AuthDevice(ctx, e.authRequest("aa:bb:cc:dd:ee:09", ""))
		require.ErrorIs(t, err, ErrAuthNamespaceRequiresCredential)

		device, err := e.st.DeviceResolve(ctx, scope.MustBounded(e.tenantID), store.DeviceUIDResolver, uid)
		require.NoError(t, err)
		require.NotNil(t, device.RemovedAt, "the refusal revives nothing")
	})
}

func TestProvisioningKeysInLegacyNamespace(t *testing.T) {
	ctx := context.Background()
	e := setupLegacyEnrollmentE2E(t)

	t.Run("create refuses", func(t *testing.T) {
		_, err := e.svc.CreateProvisioningKey(ctx, &requests.CreateProvisioningKey{TenantID: e.tenantID, Name: "fleet", Mode: string(models.ProvisioningKeyModeManual), UserID: "00000000-0000-4000-0000-000000000009"})
		require.ErrorIs(t, err, ErrNamespaceLegacy)
	})

	t.Run("update refuses", func(t *testing.T) {
		name := "renamed"
		err := e.svc.UpdateProvisioningKey(ctx, &requests.UpdateProvisioningKey{TenantID: e.tenantID, CurrentName: "legacy", Name: name})
		require.ErrorIs(t, err, ErrNamespaceLegacy)
	})

	t.Run("reveal refuses", func(t *testing.T) {
		_, err := e.svc.RevealProvisioningKey(ctx, &requests.RevealProvisioningKey{TenantID: e.tenantID, Name: "legacy"})
		require.ErrorIs(t, err, ErrNamespaceLegacy)
	})

	t.Run("list serves the system keys", func(t *testing.T) {
		keys, count, err := e.svc.ListProvisioningKeys(ctx, &requests.ListProvisioningKey{TenantID: e.tenantID})
		require.NoError(t, err)
		require.Equal(t, int64(2), count)

		types := make([]models.ProvisioningKeyType, 0, len(keys))
		for _, key := range keys {
			types = append(types, key.Type)
		}

		require.ElementsMatch(t, []models.ProvisioningKeyType{models.ProvisioningKeyTypeLegacy, models.ProvisioningKeyTypePairing}, types)
	})

	t.Run("history serves the legacy key's enrollments", func(t *testing.T) {
		uid := e.enroll(t, "aa:bb:cc:dd:ee:20", "")

		legacy, err := e.st.ProvisioningKeyResolveSystem(ctx, scope.MustBounded(e.tenantID))
		require.NoError(t, err)

		events, count, err := e.svc.ListProvisioningKeyEvents(ctx, &requests.ListProvisioningKeyEvents{
			TenantID:  e.tenantID,
			ID:        legacy.ID,
			Paginator: query.Paginator{Page: 1, PerPage: 10},
		})
		require.NoError(t, err)
		require.Equal(t, int64(1), count)
		require.Equal(t, uid, events[0].DeviceUID)
	})
}

func TestSystemProvisioningKeyIsFixed(t *testing.T) {
	ctx := context.Background()
	e := setupEnrollmentE2E(t)

	disabled := true
	mode := string(models.ProvisioningKeyModeAutomatic)

	for _, req := range []*requests.UpdateProvisioningKey{
		{TenantID: e.tenantID, CurrentName: "legacy", Disabled: &disabled},
		{TenantID: e.tenantID, CurrentName: "legacy", Mode: &mode},
		{TenantID: e.tenantID, CurrentName: "pairing", Disabled: &disabled},
	} {
		require.ErrorIs(t, e.svc.UpdateProvisioningKey(ctx, req), ErrProvisioningKeyForbidden)
	}
}

func (e *enrollmentE2E) ownerPolicies(t *testing.T) []string {
	t.Helper()

	var ids []string
	require.NoError(t, e.db.NewRaw(
		"SELECT p.id::text FROM access_policies p JOIN namespaces n ON n.id = p.namespace_id WHERE n.id = ? AND p.subject_user_id = n.owner_id ORDER BY p.id",
		e.tenantID,
	).Scan(context.Background(), &ids))

	return ids
}

func (e *enrollmentE2E) policyCount(t *testing.T) int64 {
	t.Helper()

	var count int64
	require.NoError(t, e.db.NewRaw("SELECT count(*) FROM access_policies WHERE namespace_id = ?", e.tenantID).Scan(context.Background(), &count))

	return count
}

func (e *enrollmentE2E) dropPolicies(t *testing.T) {
	t.Helper()

	_, err := e.db.NewRaw("DELETE FROM access_policies WHERE namespace_id = ?", e.tenantID).Exec(context.Background())
	require.NoError(t, err)
}

func (e *enrollmentE2E) isLegacy(t *testing.T) bool {
	t.Helper()

	namespace, err := e.st.NamespaceResolve(context.Background(), store.NamespaceTenantIDResolver, e.tenantID)
	require.NoError(t, err)

	return namespace.Settings.IsLegacy()
}

func TestUpgradeNamespaceAgainstDatabase(t *testing.T) {
	ctx := context.Background()

	t.Run("seeds the owner's policy once in a namespace that has none", func(t *testing.T) {
		e := setupLegacyEnrollmentE2E(t)
		e.dropPolicies(t)

		require.NoError(t, e.svc.UpgradeNamespace(ctx, e.tenantID))
		require.False(t, e.isLegacy(t))
		require.Len(t, e.ownerPolicies(t), 1)

		require.ErrorIs(t, e.svc.UpgradeNamespace(ctx, e.tenantID), ErrNamespaceNotLegacy)
		require.Equal(t, int64(1), e.policyCount(t))
	})

	t.Run("leaves the policies a namespace already has alone", func(t *testing.T) {
		e := setupLegacyEnrollmentE2E(t)
		before := e.ownerPolicies(t)
		require.NotEmpty(t, before)

		require.NoError(t, e.svc.UpgradeNamespace(ctx, e.tenantID))
		require.False(t, e.isLegacy(t))
		require.Equal(t, before, e.ownerPolicies(t))
		require.Equal(t, int64(len(before)), e.policyCount(t))
	})

	t.Run("keeps the namespace legacy when the owner's policy cannot be written", func(t *testing.T) {
		e := setupLegacyEnrollmentE2E(t)
		e.dropPolicies(t)
		_, err := e.db.NewRaw("DELETE FROM memberships WHERE namespace_id = ?", e.tenantID).Exec(ctx)
		require.NoError(t, err)

		require.Error(t, e.svc.UpgradeNamespace(ctx, e.tenantID))
		require.True(t, e.isLegacy(t))
		require.Equal(t, int64(0), e.policyCount(t))
	})
}
