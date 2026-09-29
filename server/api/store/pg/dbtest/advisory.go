package dbtest

import (
	"context"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

// HoldAdvisoryLock takes the session-level advisory lock key on a connection of its own and
// returns that connection. The test's cleanup releases every lock the session holds and closes it,
// so a test that fails before unlocking never leaves the lock held on a pooled connection.
func HoldAdvisoryLock(t *testing.T, ctx context.Context, db *bun.DB, key int64) bun.Conn {
	t.Helper()

	holder, err := db.Conn(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = holder.ExecContext(context.WithoutCancel(ctx), "SELECT pg_advisory_unlock_all()")
		_ = holder.Close()
	})

	_, err = holder.ExecContext(ctx, "SELECT pg_advisory_lock(?)", key)
	require.NoError(t, err)

	return holder
}

// CloseWithoutUnlocking ends holder's PostgreSQL session without releasing its locks, the way a
// killed process drops them. Closing a database/sql connection would not do it: the pool the store
// opens keeps the session alive.
func CloseWithoutUnlocking(t *testing.T, ctx context.Context, holder bun.Conn) {
	t.Helper()

	err := holder.Raw(func(driverConn any) error {
		c, ok := driverConn.(*stdlib.Conn)
		require.True(t, ok)
		require.NoError(t, c.Conn().Close(ctx))

		return driver.ErrBadConn
	})
	require.ErrorIs(t, err, driver.ErrBadConn)
}

// RequireAdvisoryLockWaiter blocks until exactly one session waits for the advisory lock key, and
// fails the test when none does within 30 seconds.
func RequireAdvisoryLockWaiter(t *testing.T, ctx context.Context, db *bun.DB, key int64) {
	t.Helper()

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		var waiting int
		assert.NoError(tt, db.QueryRowContext(ctx, `
			SELECT count(*) FROM pg_locks
			WHERE locktype = 'advisory' AND NOT granted
			  AND (classid::bigint << 32 | objid::bigint) = ?
		`, key).Scan(&waiting))
		assert.Equal(tt, 1, waiting)
	}, 30*time.Second, 50*time.Millisecond)
}
