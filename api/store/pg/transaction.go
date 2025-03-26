package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
)

// TODO: this works with mongodb, but will works with bun?

func (pg *pg) WithTransaction(ctx context.Context, cb store.TransactionCb) error {
	return nil
}
