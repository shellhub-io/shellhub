package pg

import (
	"context"
	"database/sql"
	"io"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/uptrace/bun"
)

func fromSqlError(err error) error {
	switch {
	case err == nil:
		return nil
	case err == sql.ErrNoRows, err == io.EOF:
		return store.ErrNoDocuments
	default:
		return err
	}
}

func applyOptions(ctx context.Context, query *bun.SelectQuery, opts ...store.QueryOption) error {
	ctxWithQuery := context.WithValue(ctx, "query", query)
	for _, opt := range opts {
		if err := opt(ctxWithQuery); err != nil {
			return fromSqlError(err)
		}
	}

	return nil
}
