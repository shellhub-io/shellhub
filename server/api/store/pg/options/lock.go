package options

import (
	"context"
	"database/sql/driver"
	"fmt"

	"github.com/jackc/pgx/v5/stdlib"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

// MigrationLockKey is the advisory lock key that serializes the core migrations across processes.
const MigrationLockKey int64 = 0x5348_4d49_4752_0001

// WithMigrationLock runs fn while holding the session-level advisory lock key on a connection
// dedicated to it. When another session holds the lock, it logs that it is waiting and blocks until
// the holder releases it or its connection closes; a cancelled ctx aborts the wait and returns the
// context's error without running fn. fn runs on the pool, not on the locked connection, so the lock
// serializes processes and is not a transaction around fn. The lock's session turns off
// idle_session_timeout, since it sits idle while fn runs; a proxy that drops idle TCP connections
// still ends it and releases the lock early. The dedicated connection is closed rather than returned
// to the pool, so neither the lock nor that setting outlives the call.
func WithMigrationLock(ctx context.Context, db *bun.DB, key int64, fn func(ctx context.Context) error) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("open the migration lock connection: %w", err)
	}

	defer destroy(context.WithoutCancel(ctx), conn)

	if _, err := conn.ExecContext(ctx, `DO $$ BEGIN
		IF current_setting('idle_session_timeout', true) IS NOT NULL THEN
			SET idle_session_timeout = 0;
		END IF;
	END $$`); err != nil {
		return fmt.Errorf("turn off idle_session_timeout on the migration lock connection: %w", err)
	}

	var acquired bool
	if err := conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock(?)", key).Scan(&acquired); err != nil {
		return fmt.Errorf("try the migration lock: %w", err)
	}

	if !acquired {
		log.WithField("key", key).Info("waiting for another instance to finish migrating the database")

		if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_lock(?)", key); err != nil {
			return fmt.Errorf("wait for the migration lock: %w", err)
		}
	}

	defer func() {
		if _, err := conn.ExecContext(context.WithoutCancel(ctx), "SELECT pg_advisory_unlock(?)", key); err != nil {
			log.WithError(err).Error("failed to release the migration lock; closing its connection releases it")
		}
	}()

	return fn(ctx)
}

func destroy(ctx context.Context, conn bun.Conn) {
	_ = conn.Raw(func(driverConn any) error {
		c, ok := driverConn.(*stdlib.Conn)
		if !ok {
			log.WithField("driver", fmt.Sprintf("%T", driverConn)).Error("cannot close the migration lock connection; the lock stays held until it closes")

			return driver.ErrBadConn
		}

		if err := c.Conn().Close(ctx); err != nil {
			log.WithError(err).Warn("failed to close the migration lock connection")
		}

		return driver.ErrBadConn
	})
}
