package pg_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPolicySubjectTypedColumnsMigration covers migration 028. It converts a polymorphic
// subject_value into two typed columns and then constrains them, so it can only run once every row
// left is one the constraints accept: it drops the policies that name a user who is not a member,
// a malformed id, or a role nothing defines, and the identities enrolled for a non-member. Dropping
// too much silently revokes access; dropping too little leaves the constraint uncreatable and the
// upgrade dead at startup.
func TestPolicySubjectTypedColumnsMigration(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 27)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	const (
		tenant     = "22222222-2222-4222-8222-222222222222"
		memberID   = "11111111-1111-4111-8111-111111111111"
		strangerID = "44444444-4444-4444-8444-444444444444"
	)

	seedUser(t, ctx, db, memberID, "member")
	seedUser(t, ctx, db, strangerID, "stranger")
	seedNamespace(t, ctx, db, tenant, "ns", memberID, fixtureTime)

	execSQL(t, ctx, db, `
		INSERT INTO memberships (user_id, namespace_id, created_at, updated_at, role)
		VALUES (?, ?, now(), now(), 'owner')
	`, memberID, tenant)

	policy := func(name, subjectType, subjectValue string) {
		t.Helper()

		execSQL(t, ctx, db, `
			INSERT INTO access_policies
			    (id, namespace_id, name, subject_type, subject_value, logins, source_ip,
			     action, require_reauth, created_at, updated_at)
			VALUES (gen_random_uuid(), ?, ?, ?, ?, '{root}', '{}', 'deny', false, now(), now())
		`, tenant, name, subjectType, subjectValue)
	}

	policy("keeps-the-member", "user", memberID)
	policy("drops-the-stranger", "user", strangerID)
	policy("drops-the-malformed-id", "user", "not-a-uuid")
	policy("keeps-the-role", "role", "operator")
	policy("drops-the-invented-role", "role", "superadmin")
	policy("keeps-all-members", "all-members", "")

	identity := func(name, userID string) {
		t.Helper()

		execSQL(t, ctx, db, `
			INSERT INTO ssh_identities
			    (id, namespace_id, user_id, fingerprint, data, name, source, created_at)
			VALUES (gen_random_uuid(), ?, ?, ?, 'ssh-ed25519 AAAA key', ?, 'manual', now())
		`, tenant, userID, "SHA256:"+name, name)
	}

	identity("keeps-the-member-key", memberID)
	identity("drops-the-stranger-key", strangerID)

	require.NoError(t, provider.ApplyNext(ctx), "028 must apply cleanly")

	var policies []string
	require.NoError(t, db.NewRaw("SELECT name FROM access_policies ORDER BY name").Scan(ctx, &policies))
	assert.Equal(t, []string{"keeps-all-members", "keeps-the-member", "keeps-the-role"}, policies,
		"only the policies a member or a defined role can match survive")

	var identities []string
	require.NoError(t, db.NewRaw("SELECT name FROM ssh_identities ORDER BY name").Scan(ctx, &identities))
	assert.Equal(t, []string{"keeps-the-member-key"}, identities,
		"a key enrolled for someone who is not a member is revoked")

	var subject struct {
		UserID string `bun:"subject_user_id"`
		Role   string `bun:"subject_role"`
	}
	require.NoError(t, db.
		NewRaw(`SELECT coalesce(subject_user_id::text, '') AS subject_user_id,
		               coalesce(subject_role::text, '') AS subject_role
		        FROM access_policies WHERE name = 'keeps-the-member'`).
		Scan(ctx, &subject))
	assert.Equal(t, memberID, subject.UserID, "the id moves into the typed column")
	assert.Empty(t, subject.Role)
}
