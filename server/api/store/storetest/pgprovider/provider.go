package pgprovider

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lib/pq"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/pg"
	"github.com/shellhub-io/shellhub/server/api/store/pg/dbtest"
	"github.com/shellhub-io/shellhub/server/api/store/pg/options"
	"github.com/uptrace/bun"
	"gopkg.in/yaml.v3"
)

// Provider implements storetest.StoreProvider for PostgreSQL
type Provider struct {
	srv         *dbtest.Server
	store       store.Store
	driver      *bun.DB
	fixtureRoot string
}

// NewProvider creates a new PostgreSQL test provider
func NewProvider(ctx context.Context) (*Provider, error) {
	srv := &dbtest.Server{}

	if err := srv.Up(ctx); err != nil {
		return nil, err
	}

	connString, err := srv.ConnectionString(ctx)
	if err != nil {
		_ = srv.Down(ctx)

		return nil, err
	}

	st, err := pg.New(ctx, connString, options.Migrate())
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
