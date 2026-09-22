package pg_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLegacyPermissionMigration covers migration 036 against the rows only a release candidate
// can hold: a namespace left in legacy mode without the permission legacy requires. Nothing
// creates that pair after this migration, so the backfill is reachable from here alone, and the
// constraint it installs is what keeps it unreachable.
func TestLegacyPermissionMigration(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 35)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	const (
		ownerID    = "11111111-1111-4111-8111-111111111111"
		brokenID   = "22222222-2222-4222-8222-222222222222"
		grandpaID  = "33333333-3333-4333-8333-333333333333"
		identityID = "44444444-4444-4444-8444-444444444444"
	)

	seedUser(t, ctx, db, ownerID, "owner")

	mode := func(id, name, accessMode string, legacyAllowed bool) {
		t.Helper()

		seedNamespace(t, ctx, db, id, name, ownerID, fixtureTime)
		execSQL(t, ctx, db, `
			UPDATE namespaces SET ssh_access_mode = ?, ssh_legacy_allowed = ? WHERE id = ?
		`, accessMode, legacyAllowed, id)
	}

	mode(brokenID, "born-legacy-by-cli", "legacy", false)
	mode(grandpaID, "grandfathered", "legacy", true)
	mode(identityID, "identity-first", "identity", false)

	require.NoError(t, provider.ApplyNext(ctx), "036 must apply over release-candidate data")

	allowed := func(id string) bool {
		t.Helper()

		var value bool
		require.NoError(t, db.NewRaw("SELECT ssh_legacy_allowed FROM namespaces WHERE id = ?", id).Scan(ctx, &value))

		return value
	}

	assert.True(t, allowed(brokenID), "the namespace stuck in legacy without the permission gets it")
	assert.True(t, allowed(grandpaID), "the backfill never flips a grandfathered permission off")
	assert.False(t, allowed(identityID), "identity mode is not granted a permission it never needed")

	_, err = db.ExecContext(ctx, `
		UPDATE namespaces SET ssh_legacy_allowed = false WHERE id = ?
	`, grandpaID)
	require.ErrorContains(t, err, "namespaces_legacy_requires_permission",
		"legacy without the permission is refused from here on")

	misspelledMode := "legcay"

	_, err = db.ExecContext(ctx, `
		UPDATE namespaces SET ssh_access_mode = ? WHERE id = ?
	`, misspelledMode, identityID)
	require.ErrorContains(t, err, "namespaces_ssh_access_mode_check",
		"a misspelled mode is refused rather than stored as one nothing recognises")
}
