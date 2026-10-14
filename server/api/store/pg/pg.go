package pg

import (
	"context"
	"fmt"
	"net"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/server/api/store/pg/options"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

type queryOptions struct {
	clock clock.Clock
}

// Pg is the PostgreSQL implementation of [store.Store].
type Pg struct {
	driver  *bun.DB
	clock   clock.Clock
	options *queryOptions
}

// URI assembles a PostgreSQL connection string from its parts.
func URI(host, port, user, password, db string) string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s", user, password, net.JoinHostPort(host, port), db)
}

// New opens the store at uri and applies each option. It pings before returning, so a
// server that starts has a database it can actually reach. The store reads the wall clock unless
// an option gives it another one.
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

	driver := bun.NewDB(stdlib.OpenDBFromPool(pool), pgdialect.New(), bun.WithDiscardUnknownColumns())
	if err := driver.PingContext(ctx); err != nil {
		return nil, err
	}

	driver.RegisterModel(entity.Entities()...)

	prepared := &options.Store{DB: driver, Clock: clock.New()}
	for _, opt := range opts {
		if err := opt(ctx, prepared); err != nil {
			return nil, err
		}
	}

	return &Pg{driver: prepared.DB, clock: prepared.Clock, options: &queryOptions{clock: prepared.Clock}}, nil
}

// Clock returns the clock the store reads, so a store that wraps this one reads the same time.
func (pg *Pg) Clock() clock.Clock {
	return pg.clock
}

// Driver exposes the underlying bun DB, for tests and for the enterprise build's own queries.
func (pg *Pg) Driver() *bun.DB {
	return pg.driver
}
