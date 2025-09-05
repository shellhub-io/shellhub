package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
)

func (pg *Pg) WithTransaction(ctx context.Context, cb store.TransactionCb) error {
	return nil
}
