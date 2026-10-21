package pg_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

func TestSystemTelemetryMigration(t *testing.T) {
	ctx := context.Background()

	t.Run("an existing row gets an instance ID and telemetry on", func(t *testing.T) {
		provider, err := pgprovider.NewProviderAt(ctx, 50)
		require.NoError(t, err)
		t.Cleanup(func() { _ = provider.Close(t) })

		db := provider.DB()

		const systemID = "33333333-3333-4333-8333-333333333333"

		execSQL(t, ctx, db, `INSERT INTO systems (id, setup) VALUES (?, true)`, systemID)

		require.NoError(t, provider.ApplyNext(ctx), "051 must apply cleanly")

		var instanceID string
		var enabled bool
		require.NoError(t, db.NewRaw(`SELECT instance_id, telemetry_enabled FROM systems WHERE id = ?`, systemID).Scan(ctx, &instanceID, &enabled))
		assert.NotEmpty(t, instanceID)
		assert.True(t, enabled)
		assert.Equal(t, 1, countSystems(t, ctx, db), "the existing row is the only row")

		require.NoError(t, provider.Rollback(ctx))

		assert.Empty(t, systemsTelemetryColumns(t, ctx, db))
	})

	t.Run("an empty table gets its row", func(t *testing.T) {
		provider, err := pgprovider.NewProviderAt(ctx, 50)
		require.NoError(t, err)
		t.Cleanup(func() { _ = provider.Close(t) })

		db := provider.DB()

		require.NoError(t, provider.ApplyNext(ctx), "051 must apply cleanly")

		var setup bool
		var instanceID string
		require.NoError(t, db.NewRaw(`SELECT setup, instance_id FROM systems`).Scan(ctx, &setup, &instanceID))
		assert.False(t, setup, "the inserted row is an instance that has not run setup")
		assert.NotEmpty(t, instanceID)
		assert.Equal(t, 1, countSystems(t, ctx, db))
	})

	t.Run("duplicate rows collapse to the set-up row, and no second row can be added", func(t *testing.T) {
		provider, err := pgprovider.NewProviderAt(ctx, 50)
		require.NoError(t, err)
		t.Cleanup(func() { _ = provider.Close(t) })

		db := provider.DB()

		const setUpID = "ffffffff-ffff-4fff-8fff-ffffffffffff"

		execSQL(t, ctx, db, `INSERT INTO systems (id, setup) VALUES (?, false)`, "11111111-1111-4111-8111-111111111111")
		execSQL(t, ctx, db, `INSERT INTO systems (id, setup) VALUES (?, true)`, setUpID)
		execSQL(t, ctx, db, `INSERT INTO systems (id, setup) VALUES (?, false)`, "22222222-2222-4222-8222-222222222222")

		require.NoError(t, provider.ApplyNext(ctx), "051 must apply cleanly")

		var kept string
		require.NoError(t, db.NewRaw(`SELECT id FROM systems`).Scan(ctx, &kept))
		assert.Equal(t, setUpID, kept)

		_, err = db.ExecContext(ctx, `INSERT INTO systems (id) VALUES (gen_random_uuid())`)
		require.Error(t, err, "systems holds a single row")
		assert.Equal(t, 1, countSystems(t, ctx, db))

		require.NoError(t, provider.Rollback(ctx))

		_, err = db.ExecContext(ctx, `INSERT INTO systems (id) VALUES (gen_random_uuid())`)
		require.NoError(t, err, "rolling 051 back lifts the single-row rule")
	})
}

func countSystems(t *testing.T, ctx context.Context, db *bun.DB) int {
	t.Helper()

	var count int
	require.NoError(t, db.NewRaw(`SELECT count(*) FROM systems`).Scan(ctx, &count))

	return count
}

func systemsTelemetryColumns(t *testing.T, ctx context.Context, db *bun.DB) []string {
	t.Helper()

	var columns []string
	require.NoError(t, db.NewRaw(`
		SELECT column_name FROM information_schema.columns
		WHERE table_name = 'systems' AND column_name IN ('instance_id', 'telemetry_enabled')`).Scan(ctx, &columns))

	return columns
}
