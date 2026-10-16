package pg_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/pg/migrations"
	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

const (
	lastMigrationBeforeProvisioningKeys = 12
	legacyNamespaceFixturePath          = "../../../../tests/environment/legacy_namespace.sql"
)

func TestLegacyNamespaceFixtureMatchesAMigratedNamespace(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, lastMigrationBeforeProvisioningKeys)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	const (
		ownerID    = "11111111-1111-4111-8111-111111111111"
		migratedID = "22222222-2222-4222-8222-222222222222"
		fixtureID  = "33333333-3333-4333-8333-333333333333"
	)

	seedUser(t, ctx, db, ownerID, "owner")
	execSQL(t, ctx, db, `
		INSERT INTO namespaces (id, created_at, updated_at, scope, name, owner_id, max_devices, record_sessions, connection_announcement)
		VALUES (?, now(), now(), 'team', 'from-0-26', ?, -1, true, '')
	`, migratedID, ownerID)
	execSQL(t, ctx, db, `
		INSERT INTO memberships (user_id, namespace_id, created_at, updated_at, role) VALUES (?, ?, now(), now(), 'owner')
	`, ownerID, migratedID)

	for range len(migrations.All) - lastMigrationBeforeProvisioningKeys {
		require.NoError(t, provider.ApplyNext(ctx))
	}

	fixture, err := os.ReadFile(filepath.Clean(legacyNamespaceFixturePath))
	require.NoError(t, err)

	statement := strings.NewReplacer(":'tenant'", "?", ":'name'", "?", ":'owner'", "?", ":'max_devices'", "?", ":'announcement'", "?").Replace(string(fixture))
	execSQL(t, ctx, db, statement, fixtureID, "fixture", "owner", -1, "")

	namespaceShape := func(id string) []map[string]any {
		return rowsOf(t, ctx, db, `
			SELECT scope, owner_id, max_devices, record_sessions, connection_announcement, legacy,
			       devices_accepted_count, devices_pending_count, devices_rejected_count, devices_removed_count
			FROM namespaces WHERE id = ?
		`, id)
	}
	membershipShape := func(id string) []map[string]any {
		return rowsOf(t, ctx, db, `SELECT user_id, role FROM memberships WHERE namespace_id = ? ORDER BY role`, id)
	}
	keysShape := func(id string) []map[string]any {
		return rowsOf(t, ctx, db, `
			SELECT encode(sha256((CASE type WHEN 'legacy' THEN 'system:' ELSE 'system:pairing:' END || namespace_id)::bytea), 'hex') = key_digest AS digest_derived_from_tenant,
			       name, mode, webhook_url, webhook_secret, allowed_identities, webhook_timeout, webhook_callback_ttl,
			       reusable, usage_limit, used_times, last_used_at, ephemeral, ephemeral_timeout, tags, revoked, disabled,
			       type, key_encrypted, key_hint, user_id, expires_at
			FROM provisioning_keys WHERE namespace_id = ? ORDER BY type
		`, id)
	}
	policies := func(id string) int {
		var count int
		require.NoError(t, db.NewRaw(`SELECT count(*) FROM access_policies WHERE namespace_id = ?`, id).Scan(ctx, &count))

		return count
	}

	assert.Equal(t, namespaceShape(migratedID), namespaceShape(fixtureID))
	assert.Equal(t, membershipShape(migratedID), membershipShape(fixtureID))
	assert.Equal(t, keysShape(migratedID), keysShape(fixtureID))
	assert.Equal(t, 0, policies(migratedID))
	assert.Equal(t, 0, policies(fixtureID))
}

func rowsOf(t *testing.T, ctx context.Context, db *bun.DB, query string, args ...any) []map[string]any {
	t.Helper()

	var rows []map[string]any
	require.NoError(t, db.NewRaw(query, args...).Scan(ctx, &rows))
	require.NotEmpty(t, rows)

	return rows
}
