package pg_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/pg/migrations"
	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDropPrivateKeysMigration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	before := 0
	for _, migration := range migrations.All {
		number, err := strconv.Atoi(migration.Name)
		require.NoError(t, err)

		if number < 58 {
			before = number
		}
	}

	provider, err := pgprovider.NewProviderAt(ctx, before)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()
	require.True(t, tableExists(t, ctx, db, "private_keys"))

	require.NoError(t, provider.ApplyNext(ctx))
	assert.False(t, tableExists(t, ctx, db, "private_keys"))

	require.NoError(t, provider.Rollback(ctx))
	assert.True(t, tableExists(t, ctx, db, "private_keys"), "the down migration restores the table, empty")
}
