package pgprovider

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"

	"github.com/lib/pq"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/pg"
	"github.com/shellhub-io/shellhub/server/api/store/pg/dbtest"
	"github.com/shellhub-io/shellhub/server/api/store/pg/migrations"
	"github.com/shellhub-io/shellhub/server/api/store/pg/options"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
	"gopkg.in/yaml.v3"
)

// Provider implements storetest.StoreProvider for PostgreSQL
type Provider struct {
	srv         *dbtest.Server
	store       store.Store
	driver      *bun.DB
	fixtureRoot string
	version     int
	booted      int
	failed      int
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

// ApplyNext runs the migration following the one the provider is at, inside its own transaction when
// the file is transactional, and reports the migration's own error. The migrator records a migration
// as applied before running it, so a failed one keeps that record while its changes roll back: the
// provider then refuses every later ApplyNext and Rollback rather than working from a version its
// database no longer matches.
func (p *Provider) ApplyNext(ctx context.Context) error {
	if err := p.describesItsDatabase(); err != nil {
		return err
	}

	next := p.version + 1

	registry, err := migrationsThrough(next)
	if err != nil {
		return err
	}

	if err := applyRegistry(ctx, p.driver, registry, next); err != nil {
		p.failed = next

		return err
	}

	p.version = next

	return nil
}

// Rollback undoes the migration the last ApplyNext ran, and fails when the migrator undoes anything
// else. It refuses when no ApplyNext has run, because the migrator rolls back a whole group and the
// provider booted with every migration up to its version in one. A migration whose down file
// reverses nothing, as 014 and 022 do, is reported rolled back with its rows still in place. A
// rollback that fails leaves the provider refusing later calls, as a failed ApplyNext does: the
// migrator unmarks a migration before running its down, so the database no longer matches either.
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
		return fmt.Errorf("migration %03d failed and stays recorded as applied, so this provider no longer describes its database", p.failed)
	}

	return nil
}

func headVersion() (int, error) {
	sorted := migrations.FetchMigrations().Sorted()
	if len(sorted) == 0 {
		return 0, errors.New("no migrations are registered")
	}

	return strconv.Atoi(sorted[len(sorted)-1].Name)
}

func migrateThrough(version int) options.Option {
	return func(ctx context.Context, db *bun.DB) error {
		registry, err := migrationsThrough(version)
		if err != nil {
			return err
		}

		return applyRegistry(ctx, db, registry, version)
	}
}

func applyRegistry(ctx context.Context, db *bun.DB, registry *migrate.Migrations, version int) error {
	migrator := migrate.NewMigrator(db, registry)
	if err := migrator.Init(ctx); err != nil {
		return err
	}

	group, err := migrator.Migrate(ctx)
	if err != nil {
		return err
	}

	if group.IsZero() {
		return fmt.Errorf("migration %03d is already applied", version)
	}

	return nil
}

func migrationsThrough(version int) (*migrate.Migrations, error) {
	registry := migrate.NewMigrations()

	found := false

	for _, migration := range migrations.FetchMigrations().Sorted() {
		number, err := strconv.Atoi(migration.Name)
		if err != nil {
			return nil, err
		}

		if number > version {
			break
		}

		registry.Add(migration)

		found = number == version
	}

	if !found {
		return nil, fmt.Errorf("no migration numbered %03d", version)
	}

	return registry, nil
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

	_, file, _, _ := runtime.Caller(0)
	fixturesPath := filepath.Join(filepath.Dir(file), "..", "fixtures")

	return &Provider{
		srv:         srv,
		store:       st,
		driver:      driver,
		fixtureRoot: fixturesPath,
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

// LoadFixtures loads test data from YAML fixture files
func (p *Provider) LoadFixtures(t *testing.T, fixtures ...string) error {
	t.Helper()
	ctx := context.Background()

	for _, fixtureName := range fixtures {
		filePath := filepath.Join(p.fixtureRoot, fixtureName+".yml")

		data, err := os.ReadFile(filePath) //nolint:gosec // filePath is constructed from a trusted fixture root, not user input.
		if err != nil {
			return fmt.Errorf("failed to read fixture %s: %w", fixtureName, err)
		}

		var records []map[string]any
		if err := yaml.Unmarshal(data, &records); err != nil {
			return fmt.Errorf("failed to parse fixture %s: %w", fixtureName, err)
		}

		t.Logf("Loading %d records from fixture %s", len(records), fixtureName)

		if err := p.insertFixture(ctx, fixtureName, records); err != nil {
			return fmt.Errorf("failed to insert fixture %s: %w", fixtureName, err)
		}

		t.Logf("Successfully loaded fixture %s", fixtureName)

		count, err := p.driver.NewSelect().Table(fixtureName).Count(ctx)
		if err != nil {
			t.Logf("Warning: could not count records in %s: %v", fixtureName, err)
		} else {
			t.Logf("Verified: %d records now in table %s", count, fixtureName)
		}
	}

	return nil
}

func (p *Provider) insertFixture(ctx context.Context, fixtureName string, records []map[string]any) error {
	if len(records) == 0 {
		return nil
	}

	tableName := fixtureName

	for _, record := range records {
		processedRecord := p.processRecordForPostgres(record)

		_, err := p.driver.NewInsert().
			Model(&processedRecord).
			Table(tableName).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to insert record into %s: %w", tableName, err)
		}
	}

	return nil
}

func (p *Provider) processRecordForPostgres(record map[string]any) map[string]any {
	processed := make(map[string]any)

	for key, value := range record {
		switch v := value.(type) {
		case []any:
			if len(v) > 0 {
				if _, ok := v[0].(string); ok {
					strArray := make([]string, len(v))
					for i, item := range v {
						strArray[i], _ = item.(string)
					}
					processed[key] = pq.Array(strArray)
				} else {
					processed[key] = pq.Array(v)
				}
			} else {
				processed[key] = pq.Array([]string{}) // empty array
			}
		default:
			processed[key] = value
		}
	}

	return processed
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
