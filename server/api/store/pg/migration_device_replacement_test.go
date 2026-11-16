package pg_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

func TestDeviceReplacementMigration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 61)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	const (
		ownerID  = "11111111-1111-4111-8111-111111111111"
		tenantID = "22222222-2222-4222-8222-222222222222"
	)

	seedUser(t, ctx, db, ownerID, "owner")
	seedNamespace(t, ctx, db, tenantID, "existing", ownerID, fixtureTime)
	execSQL(t, ctx, db, `
		INSERT INTO provisioning_keys (key_digest, namespace_id, name, reusable, mode, type, created_at, updated_at)
		VALUES ('digest', ?, 'fleet', true, 'automatic', 'user', now(), now())`, tenantID)

	require.NoError(t, provider.ApplyNext(ctx), "062 must apply cleanly")

	var replaces bool
	require.NoError(t, db.NewRaw(`SELECT replace_offline_devices FROM provisioning_keys WHERE key_digest = 'digest'`).Scan(ctx, &replaces))
	assert.True(t, replaces, "a key that exists at the update keeps replacing offline devices")

	assert.ElementsMatch(t, []string{"replaced_device_uid", "replaced_device_name"}, replacementColumns(t, ctx, db, "provisioning_key_events"))

	execSQL(t, ctx, db, `
		INSERT INTO provisioning_key_events (id, provisioning_key_id, namespace_id, device_uid, hostname, created_at, replaced_device_uid, replaced_device_name)
		VALUES
			('33333333-3333-4333-8333-333333333331', 'digest', ?, 'newcomer', 'evilbox', now(), NULL, NULL),
			('33333333-3333-4333-8333-333333333332', 'digest', ?, 'newcomer', 'evilbox', now(), 'old', 'field-box')`, tenantID, tenantID)

	require.NoError(t, provider.Rollback(ctx))

	assert.Empty(t, replacementColumns(t, ctx, db, "provisioning_key_events"))

	var ids []string
	require.NoError(t, db.NewRaw(`SELECT id::text FROM provisioning_key_events`).Scan(ctx, &ids))
	assert.Equal(t, []string{"33333333-3333-4333-8333-333333333331"}, ids, "the rollback drops the replacement rows, which would read as enrollments without their columns, and keeps the enrollment")
}

func replacementColumns(t *testing.T, ctx context.Context, db *bun.DB, table string) []string {
	t.Helper()

	var columns []string
	require.NoError(t, db.NewRaw(`
		SELECT column_name FROM information_schema.columns
		WHERE table_name = ? AND column_name LIKE 'replaced_%'`, table).Scan(ctx, &columns))

	return columns
}
