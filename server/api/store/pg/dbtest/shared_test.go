package dbtest_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/pg/dbtest"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func TestMain(m *testing.M) {
	os.Exit(dbtest.Run(m))
}

func openDatabase(t *testing.T, ctx context.Context, template string) (*bun.DB, string) {
	t.Helper()

	srv, err := dbtest.Shared(ctx)
	require.NoError(t, err)

	name, err := srv.CreateDatabase(ctx, template)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, srv.DropDatabase(context.WithoutCancel(ctx), name)) })

	dsn, err := srv.DatabaseURL(ctx, name)
	require.NoError(t, err)

	sqldb, err := sql.Open("pgx", dsn)
	require.NoError(t, err)

	db := bun.NewDB(sqldb, pgdialect.New())
	t.Cleanup(func() { _ = db.Close() })

	return db, name
}

func TestSharedDatabasesAreIsolated(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	first, _ := openDatabase(t, ctx, "")
	second, _ := openDatabase(t, ctx, "")

	_, err := first.ExecContext(ctx, "CREATE TABLE marker (id int)")
	require.NoError(t, err)

	var found bool
	require.NoError(t, second.QueryRowContext(ctx, "SELECT to_regclass('public.marker') IS NOT NULL").Scan(&found))
	require.False(t, found)
}

func TestSharedDatabaseClonesItsTemplate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	srv, err := dbtest.Shared(ctx)
	require.NoError(t, err)

	template, err := srv.CreateDatabase(ctx, "")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, srv.DropDatabase(context.WithoutCancel(ctx), template)) })

	dsn, err := srv.DatabaseURL(ctx, template)
	require.NoError(t, err)

	sqldb, err := sql.Open("pgx", dsn)
	require.NoError(t, err)

	_, err = sqldb.ExecContext(ctx, "CREATE TABLE marker (id int); INSERT INTO marker VALUES (7)")
	require.NoError(t, err)
	require.NoError(t, sqldb.Close())

	require.NoError(t, srv.SealTemplate(ctx, template))

	clone, _ := openDatabase(t, ctx, template)
	again, _ := openDatabase(t, ctx, template)

	for _, db := range []*bun.DB{clone, again} {
		var id int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT id FROM marker").Scan(&id))
		require.Equal(t, 7, id)
	}
}

func TestRequireAdvisoryLockWaiterCountsOnlyItsDatabase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	const key = 4242

	for range 2 {
		db, _ := openDatabase(t, ctx, "")
		dbtest.HoldAdvisoryLock(t, ctx, db, key)

		waiter, err := db.Conn(ctx)
		require.NoError(t, err)

		waitCtx, stopWaiting := context.WithCancel(ctx)
		t.Cleanup(func() {
			stopWaiting()
			_ = waiter.Close()
		})

		go func() { _, _ = waiter.ExecContext(waitCtx, "SELECT pg_advisory_lock(?)", key) }()

		dbtest.RequireAdvisoryLockWaiter(t, ctx, db, key)
	}
}
