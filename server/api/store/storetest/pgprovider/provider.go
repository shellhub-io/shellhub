package pgprovider

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/pg"
	"github.com/shellhub-io/shellhub/server/api/store/pg/dbtest"
	"github.com/shellhub-io/shellhub/server/api/store/pg/migrations"
	"github.com/shellhub-io/shellhub/server/api/store/pg/migrator"
	"github.com/shellhub-io/shellhub/server/api/store/pg/options"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
)

// Provider implements storetest.StoreProvider for PostgreSQL
type Provider struct {
	srv     *dbtest.Server
	store   store.Store
	driver  *bun.DB
	version int
	booted  int
	failed  int
}

// NewProvider creates a provider migrated to head, reporting the version of the last migration in
// the registry so ApplyNext and Rollback name where the database actually is.
func NewProvider(ctx context.Context) (*Provider, error) {
	head, err := headVersion()
	if err != nil {
		return nil, err
	}

	provider, err := newProvider(ctx, options.Migrate())
	if err != nil {
		return nil, err
	}

	provider.version = head
	provider.booted = head

	return provider, nil
}

// NewProviderAt creates a provider whose database is migrated through version and no further, so a
// test can plant rows in the shape the migration after it will find them. It fails when no migration
// carries that version, rather than silently migrating to head.
func NewProviderAt(ctx context.Context, version int) (*Provider, error) {
	provider, err := newProvider(ctx, migrateThrough(version))
	if err != nil {
		return nil, err
	}

	provider.version = version
	provider.booted = version

	return provider, nil
}

// ApplyNext runs the migration following the one the provider is at and reports the migration's own
// error. A failed transactional migration leaves neither its changes nor its record, so the
// provider stays at its version and a later ApplyNext runs it again.
func (p *Provider) ApplyNext(ctx context.Context) error {
	if err := p.describesItsDatabase(); err != nil {
		return err
	}

	next := p.version + 1

	through, err := migrationsThrough(next)
	if err != nil {
		return err
	}

	if err := apply(ctx, p.driver, through, next); err != nil {
		return err
	}

	p.version = next

	return nil
}

// Rollback undoes the migration the last ApplyNext ran, and fails when the migrator undoes anything
// else. It refuses when no ApplyNext has run, because the migrator rolls back a whole group and the
// provider booted with every migration up to its version in one. A migration whose down reverses
// nothing is reported rolled back with its rows still in place: 014, whose down file is a no-op, and
// 022, a Go migration with no down. A rollback that fails leaves the provider refusing later calls: the migrator unmarks a migration
// before running its down, so the database no longer matches its record.
func (p *Provider) Rollback(ctx context.Context) error {
	if err := p.describesItsDatabase(); err != nil {
		return err
	}

	if p.version == p.booted {
		return fmt.Errorf("no migration applied since boot at %03d; rolling back would undo every migration", p.booted)
	}

	group, err := migrate.NewMigrator(p.driver, migrations.FetchMigrations()).Rollback(ctx)
	if err != nil {
		p.failed = p.version

		return err
	}

	want := fmt.Sprintf("%03d", p.version)
	if len(group.Migrations) != 1 || group.Migrations[0].Name != want {
		p.failed = p.version

		return fmt.Errorf("rolled back %s, want migration %s alone", group.Migrations, want)
	}

	p.version--

	return nil
}

func (p *Provider) describesItsDatabase() error {
	if p.failed != 0 {
		return fmt.Errorf("rolling back migration %03d failed, so this provider no longer describes its database", p.failed)
	}

	return nil
}

func headVersion() (int, error) {
	if len(migrations.All) == 0 {
		return 0, errors.New("no migrations are registered")
	}

	return strconv.Atoi(migrations.All[len(migrations.All)-1].Name)
}

func migrateThrough(version int) options.Option {
	return func(ctx context.Context, db *bun.DB) error {
		through, err := migrationsThrough(version)
		if err != nil {
			return err
		}

		return apply(ctx, db, through, version)
	}
}

func apply(ctx context.Context, db *bun.DB, through []migrator.Migration, version int) error {
	applied, err := migrator.Apply(ctx, db, migrations.Tables, through)
	if err != nil {
		return err
	}

	if applied == 0 {
		return fmt.Errorf("migration %03d is already applied", version)
	}

	return nil
}

func migrationsThrough(version int) ([]migrator.Migration, error) {
	for i, migration := range migrations.All {
		number, err := strconv.Atoi(migration.Name)
		if err != nil {
			return nil, err
		}

		if number == version {
			return migrations.All[:i+1], nil
		}

		if number > version {
			break
		}
	}

	return nil, fmt.Errorf("no migration numbered %03d", version)
}

func newProvider(ctx context.Context, migrateOption options.Option) (*Provider, error) {
	srv := &dbtest.Server{}

	if err := srv.Up(ctx); err != nil {
		return nil, err
	}

	connString, err := srv.ConnectionString(ctx)
	if err != nil {
		_ = srv.Down(ctx)

		return nil, err
	}

	st, err := pg.New(ctx, connString, migrateOption)
	if err != nil {
		_ = srv.Down(ctx)

		return nil, err
	}

	pgStore, ok := st.(*pg.Pg)
	if !ok {
		_ = srv.Down(ctx)

		return nil, errors.New("store is not backed by postgres")
	}

	driver := pgStore.Driver()

	return &Provider{
		srv:    srv,
		store:  st,
		driver: driver,
	}, nil
}

// Store returns the store instance
func (p *Provider) Store() store.Store {
	return p.store
}

// DB returns the underlying Bun driver. It is intended for tests that must assert on columns
// not exposed through the store models (e.g. active_sessions.created_at).
func (p *Provider) DB() *bun.DB {
	return p.driver
}

// CleanDatabase removes all data from all tables at once
// Uses a single TRUNCATE for all tables for maximum efficiency
func (p *Provider) CleanDatabase(t *testing.T) error {
	t.Helper()
	ctx := context.Background()

	query := `
		SELECT string_agg(quote_ident(tablename), ', ')
		FROM pg_tables
		WHERE schemaname = 'public'
	`

	var tableList string
	err := p.driver.QueryRowContext(ctx, query).Scan(&tableList)
	if err != nil {
		return fmt.Errorf("failed to list tables: %w", err)
	}

	if tableList == "" {
		return nil
	}

	truncateSQL := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", tableList)
	_, err = p.driver.ExecContext(ctx, truncateSQL)
	if err != nil {
		return fmt.Errorf("failed to truncate tables: %w", err)
	}

	return nil
}

// Close closes the PostgreSQL connection and stops the container
func (p *Provider) Close(t *testing.T) error {
	t.Helper()
	ctx := context.Background()

	return p.srv.Down(ctx)
}
