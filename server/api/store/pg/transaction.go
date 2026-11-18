package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/server/api/store"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

type txKeyType struct{}

var txKey = txKeyType{}

// GetConnection returns the appropriate executor for the given context.
// If the context contains an active transaction, it returns the transaction handle.
// Otherwise, it returns the base database driver.
//
// This allows store methods to be written agnostic of whether they are
// running inside a transaction or not.
func (pg *Pg) GetConnection(ctx context.Context) bun.IDB {
	if tx, ok := ctx.Value(txKey).(bun.Tx); ok {
		return tx
	}

	return pg.driver
}

// WithTransaction implements [store.TransactionStore].
func (pg *Pg) WithTransaction(ctx context.Context, fn store.TransactionCb) error {
	if _, ok := ctx.Value(txKey).(bun.Tx); ok {
		return fn(ctx)
	}

	tx, err := pg.driver.BeginTx(ctx, nil)
	if err != nil {
		return fromSQLError(err)
	}

	return runIn(ctx, tx, fn)
}

// WithSavepoint implements [store.TransactionStore].
func (pg *Pg) WithSavepoint(ctx context.Context, fn store.TransactionCb) error {
	tx, ok := ctx.Value(txKey).(bun.Tx)
	if !ok {
		return pg.WithTransaction(ctx, fn)
	}

	savepoint, err := tx.BeginTx(ctx, nil)
	if err != nil {
		return fromSQLError(err)
	}

	return runIn(ctx, savepoint, fn)
}

func runIn(ctx context.Context, tx bun.Tx, fn store.TransactionCb) error {
	defer func() {
		if p := recover(); p != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.WithError(rollbackErr).Error("rollback failed after panic")
			}

			panic(p)
		}
	}()

	if err := fn(context.WithValue(ctx, txKey, tx)); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.WithError(rollbackErr).Error("rollback failed after error")
		}

		return err
	}

	return fromSQLError(tx.Commit())
}

func (pg *Pg) transactionLock(ctx context.Context, key string) error {
	tx, ok := pg.GetConnection(ctx).(bun.Tx)
	if !ok {
		return store.ErrLockOutsideTransaction
	}

	if _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(hashtextextended(?, 0))", key); err != nil {
		return fromSQLError(err)
	}

	return nil
}
