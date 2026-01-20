package pg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/api/store/pg/options"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

type queryOptions struct{}

type Pg struct {
	driver  *bun.DB
	options *queryOptions
}

func URI(host, port, user, password, db string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, db)
}

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

	pg := &Pg{driver: bun.NewDB(stdlib.OpenDBFromPool(pool), pgdialect.New()), options: &queryOptions{}}
	if err := pg.driver.Ping(); err != nil {
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

func (pg *Pg) Driver() *bun.DB {
	return pg.driver
}
