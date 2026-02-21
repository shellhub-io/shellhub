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

// queryWrapper wraps a SelectQuery pointer to allow mutations
type queryWrapper struct {
	query *bun.SelectQuery
}

func applyOptions(ctx context.Context, query *bun.SelectQuery, opts ...store.QueryOption) (*bun.SelectQuery, error) {
	wrapper := &queryWrapper{query: query}
	ctxWithQuery := context.WithValue(ctx, "query", wrapper)

	for _, opt := range opts {
		if err := opt(ctxWithQuery); err != nil {
			return wrapper.query, fromSQLError(err)
		}
	}

	return wrapper.query, nil
}

// ApplyOptions is the exported version of applyOptions, allowing external packages
// (e.g. cloud store) to reuse the same query-option mechanism.
func ApplyOptions(ctx context.Context, query *bun.SelectQuery, opts ...store.QueryOption) (*bun.SelectQuery, error) {
	return applyOptions(ctx, query, opts...)
}
