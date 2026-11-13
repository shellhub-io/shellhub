package store

import "context"

// TransactionCb defines the function signature expected for transaction operations.
// It typically encompasses a series of store method calls that must be executed within a transaction.
type TransactionCb func(ctx context.Context) error

// TransactionStore runs several store operations as one unit. The transaction travels on the
// context, so an operation joins it without taking it as a parameter.
type TransactionStore interface {
	// WithTransaction executes a callback cb within a transaction, ensuring that a series of store
	// operations are executed as a single unit, committing the changes when the callback returns nil.
	// If any operation fails, the transaction is aborted, rolling back all operations and returning the
	// error from the callback. If the transaction cannot start or commit, it returns ctx's error when
	// it is cancelled or its deadline passes and [ErrInternal] otherwise.
	WithTransaction(ctx context.Context, cb TransactionCb) error

	// WithSavepoint runs cb inside a savepoint of the transaction ctx carries, so a statement that
	// fails in cb rolls back only cb's work and leaves that transaction usable. It releases the
	// savepoint when cb returns nil and returns cb's error otherwise. When ctx carries no transaction
	// it behaves as WithTransaction. If the savepoint cannot be created or released, it returns ctx's
	// error when it is cancelled or its deadline passes and [ErrInternal] otherwise.
	WithSavepoint(ctx context.Context, cb TransactionCb) error
}
