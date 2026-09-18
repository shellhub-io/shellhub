package pg_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAccessPolicyAPIKeySubjectMigration covers migration 031. What matters is what the shape
// check and the cascade decide: a policy naming a key is accepted only in that one form, and
// revoking the key takes the policy with it rather than leaving a rule that can never match.
func TestAccessPolicyAPIKeySubjectMigration(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 30)
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
		INSERT INTO access_policies
		    (id, namespace_id, name, subject_type, subject_user_id, logins, source_ip,
		     action, require_reauth, created_at, updated_at)
		VALUES (gen_random_uuid(), ?, 'the-owner', 'user', ?, '{root}', '{}', 'allow', false, now(), now())
	`, tenant, ownerID)

	require.NoError(t, provider.ApplyNext(ctx), "031 must apply cleanly")

	var keyID string
	require.NoError(t, db.NewRaw("SELECT id::text FROM api_keys WHERE name = 'ci'").Scan(ctx, &keyID))

	policy := func(name, subjectType string, userID, apiKeyID *string) error {
		_, err := db.ExecContext(ctx, `
			INSERT INTO access_policies
			    (id, namespace_id, name, subject_type, subject_user_id, subject_api_key_id,
			     logins, source_ip, action, require_reauth, created_at, updated_at)
			VALUES (gen_random_uuid(), ?, ?, ?, ?, ?, '{root}', '{}', 'allow', false, now(), now())
		`, tenant, name, subjectType, userID, apiKeyID)

		return err
	}

	require.NoError(t, policy("reaches-the-automation", "api-key", nil, &keyID))

	require.Error(t, policy("names-nothing", "api-key", nil, nil),
		"an api-key subject with no key is refused")
	owner := ownerID
	require.Error(t, policy("names-both", "api-key", &owner, &keyID),
		"a subject cannot be a person and a key at once")

	execSQL(t, ctx, db, "DELETE FROM api_keys WHERE name = 'ci'")

	var names []string
	require.NoError(t, db.NewRaw("SELECT name FROM access_policies ORDER BY name").Scan(ctx, &names))
	assert.Equal(t, []string{"the-owner"}, names,
		"revoking the key removes the rule naming it, and leaves the person's alone")
}
