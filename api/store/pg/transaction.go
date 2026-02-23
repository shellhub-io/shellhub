package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
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
		log.Debug("reusing existing SQL transaction from context")

		return tx
	}

	return pg.driver
}

// Example:
//
//	err := store.WithTransaction(ctx, func(ctx context.Context) error {
//	    db := store.GetConnection(ctx)
//	    if _, err := db.NewDelete().Model(&Device{}).Where("id = ?", id).Exec(ctx); err != nil {
//	        return err
//	    }
//
//	    return store.NamespaceIncrementDeviceCount(ctx, tenantID, models.DeviceStatusRemoved, -1)
//	})
//
// TODO: The transaction handle is stored in the context for simplicity.
// This hides the dependency and makes it less explicit.
// Consider refactoring to expose a typed TxStore in the future for better clarity.
func (pg *Pg) WithTransaction(ctx context.Context, fn store.TransactionCb) (err error) {
	tx, err := pg.driver.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.WithError(rollbackErr).Error("transaction rollback failed after panic")
			}

			panic(p)
		}
	}()

	if err := fn(context.WithValue(ctx, txKey, tx)); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.WithError(rollbackErr).Error("transaction rollback failed after error")
		}

		return err
	}

	return tx.Commit()
}
