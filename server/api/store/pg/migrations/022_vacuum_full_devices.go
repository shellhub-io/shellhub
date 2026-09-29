package migrations

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

const (
	queryCanceled    = "57014"
	lockNotAvailable = "55P03"
)

// VacuumLockTimeout bounds how long 022 waits for the lock on devices before it gives up.
var VacuumLockTimeout = 60 * time.Second

var vacuumStatementTimeout = 10 * time.Minute

func vacuumFullDevices(ctx context.Context, db *bun.DB) (err error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}

	defer func() { err = errors.Join(err, conn.Close()) }()

	if _, err := conn.ExecContext(ctx, "SET lock_timeout = ?", VacuumLockTimeout.Milliseconds()); err != nil {
		return err
	}

	if _, err := conn.ExecContext(ctx, "SET statement_timeout = ?", vacuumStatementTimeout.Milliseconds()); err != nil {
		return errors.Join(err, resetVacuumTimeouts(ctx, conn))
	}

	vacuumErr := vacuumOutOfTimeIsWarning(execVacuumStep(ctx, conn, "VACUUM (FULL, ANALYZE) devices"))

	return errors.Join(vacuumErr, resetVacuumTimeouts(ctx, conn))
}

func vacuumOutOfTimeIsWarning(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || (pgErr.Code != queryCanceled && pgErr.Code != lockNotAvailable) {
		return err
	}

	log.WithError(err).Warn("VACUUM FULL devices ran out of time and the table stays bloated; run VACUUM (FULL, ANALYZE) devices by hand to reclaim the space")

	return nil
}

func resetVacuumTimeouts(ctx context.Context, conn bun.Conn) error {
	return errors.Join(execVacuumStep(ctx, conn, "RESET lock_timeout"), execVacuumStep(ctx, conn, "RESET statement_timeout"))
}

func execVacuumStep(ctx context.Context, conn bun.Conn, query string) error {
	_, err := conn.ExecContext(ctx, query)

	return err
}
