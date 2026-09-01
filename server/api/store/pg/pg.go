package pg

import (
	"context"
	"fmt"
	"net"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/server/api/store/pg/options"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

type queryOptions struct{}

// Pg is the PostgreSQL implementation of [store.Store].
type Pg struct {
	driver  *bun.DB
	options *queryOptions
}

// URI assembles a PostgreSQL connection string from its parts.
func URI(host, port, user, password, db string) string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s", user, password, net.JoinHostPort(host, port), db)
}

// New opens the store at uri and applies each option. It pings before returning, so a
// server that starts has a database it can actually reach.
func New(ctx context.Context, uri string, opts ...options.Option) (store.Store, error) {
	config, err := pgxpool.ParseConfig(uri)
	if err != nil {
		return nil, err
	}

	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	pg := &Pg{driver: bun.NewDB(stdlib.OpenDBFromPool(pool), pgdialect.New(), bun.WithDiscardUnknownColumns()), options: &queryOptions{}}
	if err := pg.driver.PingContext(ctx); err != nil {
		return nil, err
	}

	pg.driver.RegisterModel(entity.Entities()...) // We need to register models so we can apply fixtures and relations later
	for _, opt := range opts {
		if err := opt(ctx, pg.driver); err != nil {
			return nil, err
		}
	}

	return pg, nil
}

// Driver exposes the underlying bun DB, for tests and for the enterprise build's own queries.
func (pg *Pg) Driver() *bun.DB {
	return pg.driver
}
