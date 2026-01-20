package pg

import (
	"context"
	"database/sql"
	"errors"
	"io"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/uptrace/bun"
)

func fromSQLError(err error) error {
	switch err {
	case nil:
		return nil
	case sql.ErrNoRows, io.EOF:
		return store.ErrNoDocuments
	default:
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // unique_violation
				return store.ErrDuplicate
			}
		}

		return err
	}
}

func applyOptions(ctx context.Context, query *bun.SelectQuery, opts ...store.QueryOption) error {
	ctxWithQuery := context.WithValue(ctx, "query", query)
	for _, opt := range opts {
		if err := opt(ctxWithQuery); err != nil {
			return fromSQLError(err)
		}
	}

	return nil
}
