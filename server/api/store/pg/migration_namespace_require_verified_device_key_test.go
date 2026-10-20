package pg_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

func TestNamespaceRequireVerifiedDeviceKeyMigration(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 49)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	const (
		ownerID  = "11111111-1111-4111-8111-111111111111"
		tenantID = "22222222-2222-4222-8222-222222222222"
		systemID = "33333333-3333-4333-8333-333333333333"
		eventID  = "44444444-4444-4444-8444-444444444444"
	)

	seedUser(t, ctx, db, ownerID, "owner")
	seedNamespace(t, ctx, db, tenantID, "existing", ownerID, fixtureTime)
	execSQL(t, ctx, db, `INSERT INTO systems (id, setup, authentication_device_require_verified_key) VALUES (?, true, true)`, systemID)
	execSQL(t, ctx, db, `
		INSERT INTO provisioning_keys (key_digest, namespace_id, name, reusable, mode, type, user_id, created_at, updated_at)
		VALUES ('digest', ?, 'tenant', true, 'automatic', 'system', ?, now(), now())`, tenantID, ownerID)
	execSQL(t, ctx, db, `
		INSERT INTO provisioning_key_events (id, provisioning_key_id, namespace_id, device_uid, hostname, created_at)
		VALUES (?, 'digest', ?, 'uid', 'host', ?)`, eventID, tenantID, fixtureTime)

	require.NoError(t, provider.ApplyNext(ctx), "050 must apply cleanly")

	var required bool
	require.NoError(t, db.NewRaw(`SELECT require_verified_device_key FROM namespaces WHERE id = ?`, tenantID).Scan(ctx, &required))
	assert.False(t, required, "a namespace that exists at the update keeps accepting unsigned agents")

	var refusal *string
	var attempts int
	require.NoError(t, db.NewRaw(`SELECT refusal, attempts FROM provisioning_key_events WHERE id = ?`, eventID).Scan(ctx, &refusal, &attempts))
	assert.Nil(t, refusal, "an event recorded before 050 was not a refusal")
	assert.Equal(t, 1, attempts)

	assert.Empty(t, systemsDeviceColumns(t, ctx, db), "the instance-wide rule is gone")

	require.NoError(t, provider.Rollback(ctx))

	assert.Equal(t, []string{"authentication_device_require_verified_key"}, systemsDeviceColumns(t, ctx, db))

	execSQL(t, ctx, db, `DELETE FROM systems`)
	execSQL(t, ctx, db, `INSERT INTO systems (id, setup) VALUES (?, true)`, systemID)
	require.NoError(t, db.NewRaw(`SELECT authentication_device_require_verified_key FROM systems WHERE id = ?`, systemID).Scan(ctx, &required))
	assert.True(t, required, "the down migration restores the default 046 set")
}

func systemsDeviceColumns(t *testing.T, ctx context.Context, db *bun.DB) []string {
	t.Helper()

	var columns []string
	require.NoError(t, db.NewRaw(`
		SELECT column_name FROM information_schema.columns
		WHERE table_name = 'systems' AND column_name LIKE 'authentication_device_%'`).Scan(ctx, &columns))

	return columns
}
