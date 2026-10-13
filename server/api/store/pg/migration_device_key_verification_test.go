package pg_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

func TestDeviceKeyVerificationMigration(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 45)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	const (
		upgraded = "11111111-1111-4111-8111-111111111111"
		fresh    = "22222222-2222-4222-8222-222222222222"
	)

	execSQL(t, ctx, db, `INSERT INTO systems (id, setup) VALUES (?, true)`, upgraded)

	require.NoError(t, provider.ApplyNext(ctx), "046 must apply cleanly")

	execSQL(t, ctx, db, `INSERT INTO systems (id, setup) VALUES (?, true)`, fresh)

	required := func(id string) bool {
		t.Helper()

		var value bool
		require.NoError(t, db.NewRaw(`SELECT authentication_device_require_verified_key FROM systems WHERE id = ?`, id).Scan(ctx, &value))

		return value
	}

	assert.False(t, required(upgraded), "an instance set up before 046 keeps accepting unsigned agents")
	assert.True(t, required(fresh), "an instance set up after 046 requires verified device keys")
}

func TestDeviceKeyVerificationMigrationWithoutSystemsRow(t *testing.T) {
	ctx := context.Background()

	systems := func(t *testing.T, db *bun.DB) []bool {
		t.Helper()

		var required []bool
		require.NoError(t, db.NewRaw(`SELECT authentication_device_require_verified_key FROM systems`).Scan(ctx, &required))

		return required
	}

	t.Run("an instance in use without a systems row gets one that keeps accepting unsigned agents", func(t *testing.T) {
		provider, err := pgprovider.NewProviderAt(ctx, 45)
		require.NoError(t, err)
		t.Cleanup(func() { _ = provider.Close(t) })

		db := provider.DB()

		const (
			owner  = "11111111-1111-4111-8111-111111111111"
			tenant = "22222222-2222-4222-8222-222222222222"
		)

		seedUser(t, ctx, db, owner, "owner")
		seedNamespace(t, ctx, db, tenant, "in-use", owner, fixtureTime)

		require.NoError(t, provider.ApplyNext(ctx), "046 must apply cleanly")

		assert.Equal(t, []bool{false}, systems(t, db))
	})

	t.Run("an empty database stays without a systems row, so a new instance requires verification", func(t *testing.T) {
		provider, err := pgprovider.NewProviderAt(ctx, 45)
		require.NoError(t, err)
		t.Cleanup(func() { _ = provider.Close(t) })

		require.NoError(t, provider.ApplyNext(ctx), "046 must apply cleanly")

		assert.Empty(t, systems(t, provider.DB()))
	})
}
