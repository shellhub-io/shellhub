package pg_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/pg"
	"github.com/shellhub-io/shellhub/api/store/pg/dbtest"
	"github.com/shellhub-io/shellhub/api/store/pg/options"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

var (
	srv    = (*dbtest.Server)(nil)
	s      = (store.Store)(nil)
	driver = (*bun.DB)(nil)
)

func TestMain(m *testing.M) {
	log.Info("Starting store tests")

	ctx := context.Background()

	srv = &dbtest.Server{}

	if err := srv.Up(ctx); err != nil {
		log.WithError(err).Error("Failed to UP the postgres container")
		os.Exit(1)
	}

	c, err := srv.ConnectionString(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to parse postgres connection string")
	}

	log.Info("Connecting to ", c)

	s, err = pg.New(ctx, c, options.Migrate())
	if err != nil {
		log.WithError(err).Error("Failed to create the postgres store")
		os.Exit(1)
	}

	driver, err = connectBun(ctx, c)
	if err != nil {
		log.WithError(err).Error("Failed to create a test driver")
		os.Exit(1)
	}

	code := m.Run()

	log.Info("Stopping store tests")
	if err := srv.Down(ctx); err != nil {
		log.WithError(err).Error("Failed to DOWN the postgres container")
		os.Exit(1)
	}

	os.Exit(code)
}

func connectBun(ctx context.Context, uri string) (*bun.DB, error) {
	config, err := pgxpool.ParseConfig(uri)
	if err != nil {
		return nil, err
	}

	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	return bun.NewDB(stdlib.OpenDBFromPool(pool), pgdialect.New()), nil
}
