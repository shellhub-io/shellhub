package pg_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvisioningKeyTagsMigration(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 47)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	const (
		tenant   = "22222222-2222-4222-8222-222222222222"
		neighbor = "44444444-4444-4444-8444-444444444444"
		ownerID  = "11111111-1111-4111-8111-111111111111"
		tagID    = "33333333-3333-4333-8333-333333333333"
	)

	seedUser(t, ctx, db, ownerID, "owner")
	seedNamespace(t, ctx, db, tenant, "ns", ownerID, fixtureTime)
	seedNamespace(t, ctx, db, neighbor, "neighbor", ownerID, fixtureTime)

	execSQL(t, ctx, db, `
		INSERT INTO tags (id, namespace_id, name, created_at, updated_at)
		VALUES (?, ?, 'existing', now(), now())
	`, tagID, tenant)

	key := func(digest, namespace, name, tags string) {
		execSQL(t, ctx, db, `
			INSERT INTO provisioning_keys (key_digest, namespace_id, name, reusable, mode, type, tags, user_id, created_at, updated_at)
			VALUES (?, ?, ?, true, 'automatic', 'user', ?::text[], ?, now(), now())
		`, digest, namespace, name, tags, ownerID)
	}

	key("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", tenant, "tagged", "{existing,missing}")
	key("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", tenant, "sharing", "{missing}")
	key("cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc", tenant, "untagged", "{}")
	key("dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd", neighbor, "neighbor", "{existing}")

	require.NoError(t, provider.ApplyNext(ctx), "048 must apply cleanly")

	type link struct {
		Key       string `bun:"key"`
		Namespace string `bun:"namespace"`
		Tag       string `bun:"tag"`
	}

	links := []link{}
	require.NoError(t, db.NewRaw(`
		SELECT k.name AS key, t.namespace_id AS namespace, t.name AS tag
		FROM provisioning_key_tags kt
		JOIN provisioning_keys k ON k.key_digest = kt.provisioning_key_digest AND k.namespace_id = kt.provisioning_key_namespace_id
		JOIN tags t ON t.id = kt.tag_id
	`).Scan(ctx, &links))

	assert.ElementsMatch(t, []link{
		{Key: "tagged", Namespace: tenant, Tag: "existing"},
		{Key: "tagged", Namespace: tenant, Tag: "missing"},
		{Key: "sharing", Namespace: tenant, Tag: "missing"},
		{Key: "neighbor", Namespace: neighbor, Tag: "existing"},
	}, links, "every name a key held becomes a link to the namespace's tag of that name")

	var existing string
	require.NoError(t, db.NewRaw("SELECT id FROM tags WHERE namespace_id = ? AND name = 'existing'", tenant).Scan(ctx, &existing))
	assert.Equal(t, tagID, existing, "a tag the namespace already held is linked, not recreated")

	var column bool
	require.NoError(t, db.NewRaw(`
		SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'provisioning_keys' AND column_name = 'tags')
	`).Scan(ctx, &column))
	assert.False(t, column, "the name array is gone, so nothing can read a stale name")
}
