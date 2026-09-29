package pg_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/pg/dbtest"
	"github.com/shellhub-io/shellhub/server/api/store/pg/options"
	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/require"
)

func TestMigrateMigrationLock(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProvider(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	t.Run("holder died without unlocking", func(t *testing.T) {
		dbtest.CloseWithoutUnlocking(t, ctx, dbtest.HoldAdvisoryLock(t, ctx, db, options.MigrationLockKey))

		require.NoError(t, options.Migrate()(ctx, db))
	})

	t.Run("holder alive", func(t *testing.T) {
		holder := dbtest.HoldAdvisoryLock(t, ctx, db, options.MigrationLockKey)

		done := make(chan error, 1)
		go func() { done <- options.Migrate()(ctx, db) }()

		dbtest.RequireAdvisoryLockWaiter(t, ctx, db, options.MigrationLockKey)

		select {
		case err := <-done:
			require.FailNow(t, "Migrate finished while another session held the lock", "err: %v", err)
		default:
		}

		_, err := holder.ExecContext(ctx, "SELECT pg_advisory_unlock(?)", options.MigrationLockKey)
		require.NoError(t, err)

		require.NoError(t, <-done)
	})

	t.Run("cancelled while waiting", func(t *testing.T) {
		holder := dbtest.HoldAdvisoryLock(t, ctx, db, options.MigrationLockKey)

		waitCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		done := make(chan error, 1)
		go func() { done <- options.Migrate()(waitCtx, db) }()

		dbtest.RequireAdvisoryLockWaiter(t, ctx, db, options.MigrationLockKey)
		cancel()
		require.ErrorIs(t, <-done, context.Canceled)

		_, err := holder.ExecContext(ctx, "SELECT pg_advisory_unlock(?)", options.MigrationLockKey)
		require.NoError(t, err)

		require.NoError(t, options.Migrate()(ctx, db))
	})

	t.Run("stale row in the old lock table", func(t *testing.T) {
		execSQL(t, ctx, db, "INSERT INTO bun_migration_locks (table_name) VALUES ('bun_migrations')")

		require.NoError(t, options.Migrate()(ctx, db))
	})
}
