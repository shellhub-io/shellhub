package pg_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSSHIdentityAPIKeyOwnerMigration covers migration 030. The three things it has to be true
// about are the ones a reader would doubt: a person's identity is untouched, an API key's
// identity is accepted although two foreign keys on user_id remain from 016 and 028, and the
// database itself removes what a revoked key owned.
func TestSSHIdentityAPIKeyOwnerMigration(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 29)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	const (
		tenant  = "22222222-2222-4222-8222-222222222222"
		ownerID = "11111111-1111-4111-8111-111111111111"
	)

	seedUser(t, ctx, db, ownerID, "owner")
	seedNamespace(t, ctx, db, tenant, "ns", ownerID, fixtureTime)

	execSQL(t, ctx, db, `
		INSERT INTO memberships (user_id, namespace_id, created_at, updated_at, role)
		VALUES (?, ?, now(), now(), 'owner')
	`, ownerID, tenant)

	execSQL(t, ctx, db, `
		INSERT INTO api_keys (key_digest, namespace_id, name, role, user_id, created_at, updated_at)
		VALUES ('1111111111111111111111111111111111111111111111111111111111111111', ?, 'ci', 'administrator', ?, now(), now())
	`, tenant, ownerID)

	execSQL(t, ctx, db, `
		INSERT INTO ssh_identities (id, namespace_id, user_id, fingerprint, data, name, source, created_at)
		VALUES (gen_random_uuid(), ?, ?, 'SHA256:person', 'ssh-ed25519 AAAA key', 'laptop', 'manual', now())
	`, tenant, ownerID)

	require.NoError(t, provider.ApplyNext(ctx), "030 must apply cleanly")

	var keyID string
	require.NoError(t, db.NewRaw("SELECT id::text FROM api_keys WHERE name = 'ci'").Scan(ctx, &keyID))

	enrol := func(name string) error {
		_, err := db.ExecContext(ctx, `
			INSERT INTO ssh_identities (id, namespace_id, api_key_id, fingerprint, data, name, source, created_at)
			VALUES (gen_random_uuid(), ?, ?, ?, 'ssh-ed25519 AAAA key', ?, 'manual', now())
		`, tenant, keyID, "SHA256:"+name, name)

		return err
	}

	require.NoError(t, enrol("robot"),
		"ssh_identities_member_fkey and ssh_identities_user_id_fkey are satisfied by a null user_id")

	_, err = db.ExecContext(ctx, `
		INSERT INTO ssh_identities (id, namespace_id, user_id, api_key_id, fingerprint, data, name, source, created_at)
		VALUES (gen_random_uuid(), ?, ?, ?, 'SHA256:both', 'ssh-ed25519 AAAA key', 'both', 'manual', now())
	`, tenant, ownerID, keyID)
	require.Error(t, err, "an identity owned by a person and a key at once is refused")

	_, err = db.ExecContext(ctx, `
		INSERT INTO ssh_identities (id, namespace_id, fingerprint, data, name, source, created_at)
		VALUES (gen_random_uuid(), ?, 'SHA256:orphan', 'ssh-ed25519 AAAA key', 'orphan', 'manual', now())
	`, tenant)
	require.Error(t, err, "an identity owned by nobody is refused")

	execSQL(t, ctx, db, "DELETE FROM api_keys WHERE name = 'ci'")

	var names []string
	require.NoError(t, db.NewRaw("SELECT name FROM ssh_identities ORDER BY name").Scan(ctx, &names))
	assert.Equal(t, []string{"laptop"}, names,
		"revoking the key takes its identities with it, and leaves the person's alone")
}
