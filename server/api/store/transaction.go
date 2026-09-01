package store

import (
	"context"
	"errors"
)

// ErrStartTransactionFailed is returned when a transaction could not be opened at all, as
// distinct from one that opened and was rolled back.
var ErrStartTransactionFailed = errors.New("start transaction failed")

// TransactionCb defines the function signature expected for transaction operations.
// It typically encompasses a series of store method calls that must be executed within a transaction.
type TransactionCb func(ctx context.Context) error

// TransactionStore runs several store operations as one unit. The transaction travels on the
// context, so an operation joins it without taking it as a parameter.
type TransactionStore interface {
	// WithTransaction executes a callback cb within a transaction, ensuring that a series of store
	// operations are executed as a single unit, committing the changes when the callback returns nil.
	// If any operation fails, the transaction is aborted, rolling back all operations and returning the
	// error from the callback. It returns ErrTransactionFailed if the transaction cannot start.
	WithTransaction(ctx context.Context, cb TransactionCb) error
}
