package pg_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccessPolicyAllDevicesMigration(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 46)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	const (
		tenant  = "22222222-2222-4222-8222-222222222222"
		ownerID = "11111111-1111-4111-8111-111111111111"
		tagID   = "33333333-3333-4333-8333-333333333333"
	)

	seedUser(t, ctx, db, ownerID, "owner")
	seedNamespace(t, ctx, db, tenant, "ns", ownerID, fixtureTime)

	execSQL(t, ctx, db, `
		INSERT INTO memberships (user_id, namespace_id, created_at, updated_at, role)
		VALUES (?, ?, now(), now(), 'owner')
	`, ownerID, tenant)

	execSQL(t, ctx, db, `
		INSERT INTO tags (id, namespace_id, name, created_at, updated_at)
		VALUES (?, ?, 'production', now(), now())
	`, tagID, tenant)

	policy := func(name string) {
		execSQL(t, ctx, db, `
			INSERT INTO access_policies
			    (id, namespace_id, name, subject_type, subject_user_id, logins, source_ip,
			     action, require_reauth, created_at, updated_at)
			VALUES (gen_random_uuid(), ?, ?, 'user', ?, '{root}', '{}', 'allow', false, now(), now())
		`, tenant, name, ownerID)
	}

	policy("untagged")
	policy("tagged")

	execSQL(t, ctx, db, `
		INSERT INTO access_policy_tags (access_policy_id, tag_id, created_at)
		SELECT id, ?, now() FROM access_policies WHERE name = 'tagged'
	`, tagID)

	require.NoError(t, provider.ApplyNext(ctx), "047 must apply cleanly")

	allDevices := func(name string) bool {
		var all bool
		require.NoError(t, db.NewRaw("SELECT all_devices FROM access_policies WHERE name = ?", name).Scan(ctx, &all))

		return all
	}

	assert.True(t, allDevices("untagged"), "a policy with no tags meant every device and now says so")
	assert.False(t, allDevices("tagged"), "a tagged policy keeps selecting only its tags")
}
