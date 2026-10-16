package pg_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

func TestNamespaceLegacyMigration(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 48)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	const (
		ownerID    = "11111111-1111-4111-8111-111111111111"
		legacyID   = "22222222-2222-4222-8222-222222222222"
		identityID = "33333333-3333-4333-8333-333333333333"
	)

	seedUser(t, ctx, db, ownerID, "owner")
	seedNamespace(t, ctx, db, legacyID, "from-0-26", ownerID, fixtureTime)
	seedNamespace(t, ctx, db, identityID, "born-identity", ownerID, fixtureTime)
	execSQL(t, ctx, db, `UPDATE namespaces SET ssh_access_mode = 'legacy', ssh_legacy_allowed = true WHERE id = ?`, legacyID)
	execSQL(t, ctx, db, `UPDATE namespaces SET ssh_access_mode = 'identity', ssh_legacy_allowed = false WHERE id = ?`, identityID)

	require.NoError(t, provider.ApplyNext(ctx), "049 must apply over release-candidate data")

	legacy := func(id string) bool {
		t.Helper()

		var value bool
		require.NoError(t, db.NewRaw("SELECT legacy FROM namespaces WHERE id = ?", id).Scan(ctx, &value))

		return value
	}

	assert.True(t, legacy(legacyID), "a namespace in legacy mode is marked legacy")
	assert.False(t, legacy(identityID), "a namespace in identity mode is new")
	assert.Equal(t, []string{"legacy"}, namespaceModeColumns(t, ctx, db), "the mode columns of the release candidates are gone")

	require.NoError(t, provider.Rollback(ctx))

	assert.Equal(t, []string{"ssh_access_mode", "ssh_legacy_allowed"}, namespaceModeColumns(t, ctx, db))

	mode := func(id string) (string, bool) {
		t.Helper()

		var accessMode string
		var allowed bool
		require.NoError(t, db.NewRaw("SELECT ssh_access_mode, ssh_legacy_allowed FROM namespaces WHERE id = ?", id).Scan(ctx, &accessMode, &allowed))

		return accessMode, allowed
	}

	accessMode, allowed := mode(legacyID)
	assert.Equal(t, "legacy", accessMode)
	assert.True(t, allowed)

	accessMode, allowed = mode(identityID)
	assert.Equal(t, "identity", accessMode)
	assert.False(t, allowed)

	_, err = db.ExecContext(ctx, `UPDATE namespaces SET ssh_legacy_allowed = false WHERE id = ?`, legacyID)
	require.ErrorContains(t, err, "namespaces_legacy_requires_permission", "the down migration recreates the constraints 036 added")
}

func namespaceModeColumns(t *testing.T, ctx context.Context, db *bun.DB) []string {
	t.Helper()

	var columns []string
	require.NoError(t, db.NewRaw(`
		SELECT column_name FROM information_schema.columns
		WHERE table_name = 'namespaces' AND column_name IN ('legacy', 'ssh_access_mode', 'ssh_legacy_allowed')
		ORDER BY column_name
	`).Scan(ctx, &columns))

	return columns
}
