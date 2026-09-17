package pg_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAPIKeySurrogateIDMigration covers migration 029. Every existing key must come out of it
// holding an id of its own, and a distinct one: a foreign key pointing at a shared or missing
// value would let one key's rows belong to another. The composite unique is what later pins an
// owned row to the key's own namespace, so it is asserted here rather than where it is relied on.
func TestAPIKeySurrogateIDMigration(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 28)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	const (
		tenant  = "22222222-2222-4222-8222-222222222222"
		ownerID = "11111111-1111-4111-8111-111111111111"
	)

	seedUser(t, ctx, db, ownerID, "owner")
	seedNamespace(t, ctx, db, tenant, "ns", ownerID, fixtureTime)

	apiKey := func(name, digest string) {
		t.Helper()

		execSQL(t, ctx, db, `
			INSERT INTO api_keys (key_digest, namespace_id, name, role, user_id, created_at, updated_at)
			VALUES (?, ?, ?, 'administrator', ?, now(), now())
		`, digest, tenant, name, ownerID)
	}

	apiKey("first", "1111111111111111111111111111111111111111111111111111111111111111")
	apiKey("second", "2222222222222222222222222222222222222222222222222222222222222222")

	require.NoError(t, provider.ApplyNext(ctx), "029 must apply cleanly")

	var ids []string
	require.NoError(t, db.NewRaw("SELECT id::text FROM api_keys ORDER BY name").Scan(ctx, &ids))
	require.Len(t, ids, 2)
	assert.NotEmpty(t, ids[0])
	assert.NotEqual(t, ids[0], ids[1], "a shared default would make two keys one owner")

	_, err = db.ExecContext(ctx, `
		INSERT INTO api_keys (id, key_digest, namespace_id, name, role, user_id, created_at, updated_at)
		VALUES (?, '3333333333333333333333333333333333333333333333333333333333333333', ?, 'third', 'administrator', ?, now(), now())
	`, ids[0], tenant, ownerID)
	assert.Error(t, err, "the surrogate id is unique on its own")

	var count int
	require.NoError(t, db.NewRaw(`
		SELECT count(*) FROM pg_constraint
		WHERE conrelid = 'api_keys'::regclass AND contype = 'u'
		  AND pg_get_constraintdef(oid) = 'UNIQUE (id, namespace_id)'
	`).Scan(ctx, &count))
	assert.Equal(t, 1, count, "a composite foreign key to (id, namespace_id) needs this unique")
}
