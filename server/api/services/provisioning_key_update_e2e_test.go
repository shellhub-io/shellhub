package services

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/require"
)

func (e *enrollmentE2E) storedProvisioningKey(t *testing.T, name string) *models.ProvisioningKey {
	t.Helper()

	key, err := e.st.ProvisioningKeyResolve(context.Background(), scope.MustBounded(e.tenantID), store.ProvisioningKeyNameResolver, name)
	require.NoError(t, err)

	return key
}

func (e *enrollmentE2E) changeProvisioningKey(name string, change func(*models.ProvisioningKey)) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		key, err := e.st.ProvisioningKeyResolve(ctx, scope.MustBounded(e.tenantID), store.ProvisioningKeyNameResolver, name)
		if err != nil {
			return err
		}

		change(key)

		return e.st.ProvisioningKeyUpdate(ctx, key)
	}
}

func (e *enrollmentE2E) updateProvisioningKeyWhileHeldOpen(t *testing.T, writes func(ctx context.Context) error, req *requests.UpdateProvisioningKey) error {
	t.Helper()

	return e.whileHeldOpen(t, writes, func() error { return e.svc.UpdateProvisioningKey(context.Background(), req) })
}

func TestUpdateProvisioningKeyE2E_AnEditWaitingBehindARevokeIsRefused(t *testing.T) {
	t.Parallel()
	e := setupEnrollmentE2E(t)
	e.provisioningKey(t, digest(1), "leaked", models.ProvisioningKeyModeAutomatic, models.ProvisioningKeyTypeUser, nil)

	revoke := e.changeProvisioningKey("leaked", func(key *models.ProvisioningKey) { key.Revoked = true })
	err := e.updateProvisioningKeyWhileHeldOpen(t, revoke, &requests.UpdateProvisioningKey{
		TenantID:    e.tenantID,
		CurrentName: "leaked",
		Name:        "renamed",
	})

	require.ErrorIs(t, err, ErrProvisioningKeyForbidden)
	key := e.storedProvisioningKey(t, "leaked")
	require.True(t, key.Revoked)
}

func TestUpdateProvisioningKeyE2E_ARevokeWaitingBehindAnEditKeepsBothChanges(t *testing.T) {
	t.Parallel()
	e := setupEnrollmentE2E(t)
	e.provisioningKey(t, digest(1), "fleet", models.ProvisioningKeyModeAutomatic, models.ProvisioningKeyTypeUser, nil)

	disable := e.changeProvisioningKey("fleet", func(key *models.ProvisioningKey) { key.Disabled = true })
	revoked := true
	err := e.updateProvisioningKeyWhileHeldOpen(t, disable, &requests.UpdateProvisioningKey{
		TenantID:    e.tenantID,
		CurrentName: "fleet",
		Revoked:     &revoked,
	})

	require.NoError(t, err)
	key := e.storedProvisioningKey(t, "fleet")
	require.True(t, key.Revoked)
	require.True(t, key.Disabled)
}

func TestUpdateProvisioningKeyE2E_ARefusedEditCreatesNoTag(t *testing.T) {
	t.Parallel()
	e := setupEnrollmentE2E(t)
	e.provisioningKey(t, digest(1), "leaked", models.ProvisioningKeyModeAutomatic, models.ProvisioningKeyTypeUser, nil)
	require.NoError(t, e.changeProvisioningKey("leaked", func(key *models.ProvisioningKey) { key.Revoked = true })(context.Background()))

	err := e.svc.UpdateProvisioningKey(context.Background(), &requests.UpdateProvisioningKey{
		TenantID:    e.tenantID,
		CurrentName: "leaked",
		Tags:        []string{"fresh"},
	})

	require.ErrorIs(t, err, ErrProvisioningKeyForbidden)
	_, err = e.st.TagResolve(context.Background(), scope.MustBounded(e.tenantID), store.TagNameResolver, "fresh")
	require.ErrorIs(t, err, store.ErrNoDocuments)
}
