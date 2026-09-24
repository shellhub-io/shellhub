package pg_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

func TestProvisioningKeyRenameMigration(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 39)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	const (
		ownerID  = "11111111-1111-4111-8111-111111111111"
		tenant   = "22222222-2222-4222-8222-222222222222"
		deviceID = "33333333-3333-4333-8333-333333333333"
		keyID    = "4444444444444444444444444444444444444444444444444444444444444444"
		eventID  = "55555555-5555-4555-8555-555555555555"
	)

	seedUser(t, ctx, db, ownerID, "owner")
	seedNamespace(t, ctx, db, tenant, "ns", ownerID, fixtureTime)

	execSQL(t, ctx, db, `
		INSERT INTO install_keys (key_digest, namespace_id, name, reusable, mode, type, user_id, created_at, updated_at)
		VALUES (?, ?, 'fleet', true, 'automatic', 'user', ?, now(), now())
	`, keyID, tenant, ownerID)

	execSQL(t, ctx, db, `
		INSERT INTO devices (id, namespace_id, name, mac, public_key, status, custom_fields,
		                     remote_addr, ephemeral, ephemeral_timeout, created_at, updated_at, last_seen,
		                     install_key_id)
		VALUES (?, ?, 'dev', '00:00:00:00:00:01', '', 'accepted', '{}', '10.0.0.1', false, 0, now(), now(), now(), ?)
	`, deviceID, tenant, keyID)

	execSQL(t, ctx, db, `
		INSERT INTO install_key_events (id, install_key_id, namespace_id, device_uid, hostname, created_at)
		VALUES (?, ?, ?, ?, 'dev', now())
	`, eventID, keyID, tenant, deviceID)

	notNullBefore := notNullConstraintsContaining(t, ctx, db, "install_key")
	require.NotZero(t, notNullBefore)

	require.NoError(t, provider.ApplyNext(ctx), "040 must apply cleanly over release-shaped data")

	assert.Empty(t, catalogNamesContaining(t, ctx, db, "install_key"), "no install_key name may survive the rename")
	assert.Zero(t, notNullConstraintsContaining(t, ctx, db, "install_key"), "no NOT NULL constraint may keep an install_key name")
	assert.Equal(t, notNullBefore, notNullConstraintsContaining(t, ctx, db, "provisioning_key"))
	assert.ElementsMatch(t, []string{
		"table:provisioning_keys",
		"table:provisioning_key_events",
		"column:devices.provisioning_key_id",
		"column:provisioning_key_events.provisioning_key_id",
		"constraint:devices_provisioning_key_fkey",
		"constraint:provisioning_keys_pkey",
		"constraint:provisioning_keys_mode_check",
		"constraint:provisioning_keys_namespace_id_fkey",
		"constraint:provisioning_key_events_pkey",
		"constraint:provisioning_key_events_decided_status_check",
		"constraint:provisioning_key_events_namespace_id_fkey",
		"constraint:provisioning_key_events_provisioning_key_id_namespace_id_fkey",
		"index:devices_pending_by_provisioning_key",
		"index:provisioning_keys_pkey",
		"index:provisioning_keys_namespace_id_name_unique",
		"index:provisioning_keys_namespace_type_unique",
		"index:provisioning_keys_key_digest_unique",
		"index:provisioning_key_events_pkey",
		"index:provisioning_key_events_key_time_idx",
		"index:provisioning_key_events_device_time_idx",
	}, catalogNamesContaining(t, ctx, db, "provisioning_key"))

	var linked string
	require.NoError(t, db.NewRaw(`
		SELECT e.id FROM provisioning_key_events e
		JOIN provisioning_keys k ON k.key_digest = e.provisioning_key_id
		JOIN devices d ON d.provisioning_key_id = k.key_digest
		WHERE d.id = ?
	`, deviceID).Scan(ctx, &linked))
	assert.Equal(t, eventID, linked, "the rename must keep keys, events and devices linked")

	require.NoError(t, provider.Rollback(ctx), "040 must roll back")

	assert.Empty(t, catalogNamesContaining(t, ctx, db, "provisioning_key"), "the rollback must restore every old name")
	assert.Equal(t, notNullBefore, notNullConstraintsContaining(t, ctx, db, "install_key"))
	assert.Contains(t, catalogNamesContaining(t, ctx, db, "install_key"), "index:install_keys_key_digest_unique")
}

func catalogNamesContaining(t *testing.T, ctx context.Context, db *bun.DB, fragment string) []string {
	t.Helper()

	var names []string
	require.NoError(t, db.NewRaw(`
		SELECT 'table:' || table_name FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name LIKE '%' || ? || '%'
		UNION ALL
		SELECT 'column:' || table_name || '.' || column_name FROM information_schema.columns
		WHERE table_schema = 'public' AND column_name LIKE '%' || ? || '%'
		UNION ALL
		SELECT 'constraint:' || conname FROM pg_constraint
		WHERE connamespace = 'public'::regnamespace AND contype <> 'n' AND conname LIKE '%' || ? || '%'
		UNION ALL
		SELECT 'index:' || indexname FROM pg_indexes
		WHERE schemaname = 'public' AND indexname LIKE '%' || ? || '%'
	`, fragment, fragment, fragment, fragment).Scan(ctx, &names))

	return names
}

func notNullConstraintsContaining(t *testing.T, ctx context.Context, db *bun.DB, fragment string) int {
	t.Helper()

	var count int
	require.NoError(t, db.NewRaw(`
		SELECT count(*) FROM pg_constraint
		WHERE connamespace = 'public'::regnamespace AND contype = 'n' AND conname LIKE '%' || ? || '%'
	`, fragment).Scan(ctx, &count))

	return count
}
