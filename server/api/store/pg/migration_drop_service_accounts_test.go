package pg_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDropServiceAccountsMigration covers migration 034 against data shaped like a release
// candidate's, which is the only place service accounts ever existed. It is destructive, so what
// it must not take with it matters as much as what it removes: the person keeps everything, and
// the bot's session history survives with its attribution dropped rather than the row.
func TestDropServiceAccountsMigration(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 33)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	const (
		tenant  = "22222222-2222-4222-8222-222222222222"
		ownerID = "11111111-1111-4111-8111-111111111111"
		botID   = "33333333-3333-4333-8333-333333333333"
		deviceI = "device-1"
	)

	seedUser(t, ctx, db, ownerID, "owner")
	seedNamespace(t, ctx, db, tenant, "ns", ownerID, fixtureTime)

	execSQL(t, ctx, db, `
		INSERT INTO memberships (user_id, namespace_id, created_at, updated_at, role)
		VALUES (?, ?, now(), now(), 'owner')
	`, ownerID, tenant)

	execSQL(t, ctx, db, `
		INSERT INTO users
		    (id, created_at, updated_at, origin, status, type, name, username, email,
		     password_digest, auth_methods, namespace_ownership_limit)
		VALUES (?, now(), now(), 'local', 'confirmed', 'service', 'ci-bot', 'ci-bot',
		        'svc-ci@service.local', '!', ARRAY['local']::user_auth_method[], -1)
	`, botID)
	execSQL(t, ctx, db, `
		INSERT INTO memberships (user_id, namespace_id, created_at, updated_at, role)
		VALUES (?, ?, now(), now(), 'service')
	`, botID, tenant)

	identity := func(owner, name string) {
		t.Helper()

		execSQL(t, ctx, db, `
			INSERT INTO ssh_identities (id, namespace_id, user_id, fingerprint, data, name, source, created_at)
			VALUES (gen_random_uuid(), ?, ?, ?, 'ssh-ed25519 AAAA key', ?, 'manual', now())
		`, tenant, owner, "SHA256:"+name, name)
	}

	identity(ownerID, "laptop")
	identity(botID, "bot-key-one")
	identity(botID, "bot-key-two")

	policy := func(name, subjectType string, userID, role *string) {
		t.Helper()

		execSQL(t, ctx, db, `
			INSERT INTO access_policies
			    (id, namespace_id, name, subject_type, subject_user_id, subject_role,
			     logins, source_ip, action, require_reauth, created_at, updated_at)
			VALUES (gen_random_uuid(), ?, ?, ?, ?, ?, '{root}', '{}', 'allow', false, now(), now())
		`, tenant, name, subjectType, userID, role)
	}

	owner, bot, serviceRole := ownerID, botID, "service"
	policy("for-the-owner", "user", &owner, nil)
	policy("for-the-bot", "user", &bot, nil)
	policy("for-every-bot", "role", nil, &serviceRole)

	execSQL(t, ctx, db, `
		INSERT INTO membership_invitations (id, tenant_id, user_id, invited_by, role, status, status_updated_at, expires_at, invitations, created_at, updated_at)
		VALUES (gen_random_uuid(), ?, ?, ?, 'service', 'pending', now(), now() + interval '7 days', 1, now(), now())
	`, tenant, botID, ownerID)

	execSQL(t, ctx, db, `
		INSERT INTO devices (id, namespace_id, name, mac, public_key, status, custom_fields,
		                     remote_addr, ephemeral, ephemeral_timeout, created_at, updated_at, last_seen)
		VALUES (?, ?, 'dev', '00:00:00:00:00:01', '', 'accepted', '{}', '10.0.0.1', false, 0, now(), now(), now())
	`, deviceI, tenant)
	execSQL(t, ctx, db, `
		INSERT INTO sessions (id, namespace_id, device_id, username, ip_address, user_id,
		                      started_at, seen_at, created_at, updated_at)
		VALUES ('session-of-the-bot', ?, ?, 'root', '10.0.0.1', ?, now(), now(), now(), now())
	`, tenant, deviceI, botID)

	require.NoError(t, provider.ApplyNext(ctx), "034 must apply cleanly over release-candidate data")

	var users []string
	require.NoError(t, db.NewRaw("SELECT name FROM users ORDER BY name").Scan(ctx, &users))
	assert.Equal(t, []string{"owner"}, users, "the bot's account is gone")

	var identities []string
	require.NoError(t, db.NewRaw("SELECT name FROM ssh_identities ORDER BY name").Scan(ctx, &identities))
	assert.Equal(t, []string{"laptop"}, identities, "both of the bot's keys go with it")

	var policies []string
	require.NoError(t, db.NewRaw("SELECT name FROM access_policies ORDER BY name").Scan(ctx, &policies))
	assert.Equal(t, []string{"for-the-owner"}, policies,
		"the rule naming the bot and the one naming every bot both go")

	var invitations int
	require.NoError(t, db.NewRaw("SELECT count(*) FROM membership_invitations").Scan(ctx, &invitations))
	assert.Zero(t, invitations, "a pending invite carrying the role would fail the enum rebuild")

	var attributed *string
	require.NoError(t, db.NewRaw("SELECT user_id::text FROM sessions WHERE id = 'session-of-the-bot'").Scan(ctx, &attributed))
	assert.Nil(t, attributed, "the session survives; only who ran it is dropped")

	var roles []string
	require.NoError(t, db.NewRaw("SELECT unnest(enum_range(NULL::membership_role))::text ORDER BY 1").Scan(ctx, &roles))
	assert.Equal(t, []string{"administrator", "observer", "operator", "owner"}, roles)

	var columns int
	require.NoError(t, db.NewRaw("SELECT count(*) FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'type'").Scan(ctx, &columns))
	assert.Zero(t, columns, "the column that held the species is gone")

	var types int
	require.NoError(t, db.NewRaw("SELECT count(*) FROM pg_type WHERE typname = 'user_type'").Scan(ctx, &types))
	assert.Zero(t, types)
}
