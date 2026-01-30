package pgprovider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lib/pq"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/pg"
	"github.com/shellhub-io/shellhub/api/store/pg/dbtest"
	"github.com/shellhub-io/shellhub/api/store/pg/options"
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
		srv.Down(ctx)

		return nil, err
	}

	st, err := pg.New(ctx, connString, options.Migrate())
	if err != nil {
		srv.Down(ctx)

		return nil, err
	}

	// Get direct access to Bun driver for fixture loading
	pgStore := st.(*pg.Pg)
	driver := pgStore.Driver()

	// Get fixtures path relative to this file
	_, file, _, _ := runtime.Caller(0)
	// Navigate from storetest/pgprovider/provider.go to storetest/fixtures
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

// LoadFixtures loads test data from YAML fixture files
func (p *Provider) LoadFixtures(t *testing.T, fixtures ...string) error {
	t.Helper()
	ctx := context.Background()

	for _, fixtureName := range fixtures {
		filePath := filepath.Join(p.fixtureRoot, fixtureName+".yml")

		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read fixture %s: %w", fixtureName, err)
		}

		// Parse YAML into generic structure
		var records []map[string]interface{}
		if err := yaml.Unmarshal(data, &records); err != nil {
			return fmt.Errorf("failed to parse fixture %s: %w", fixtureName, err)
		}

		t.Logf("Loading %d records from fixture %s", len(records), fixtureName)

		// Insert records based on fixture name
		if err := p.insertFixture(ctx, fixtureName, records); err != nil {
			return fmt.Errorf("failed to insert fixture %s: %w", fixtureName, err)
		}

		t.Logf("Successfully loaded fixture %s", fixtureName)

		// Debug: verify records were actually inserted
		count, err := p.driver.NewSelect().Table(fixtureName).Count(ctx)
		if err != nil {
			t.Logf("Warning: could not count records in %s: %v", fixtureName, err)
		} else {
			t.Logf("Verified: %d records now in table %s", count, fixtureName)
		}
	}

	return nil
}

// insertFixture inserts records into the appropriate table
func (p *Provider) insertFixture(ctx context.Context, fixtureName string, records []map[string]interface{}) error {
	if len(records) == 0 {
		return nil
	}

	// Map fixture name to table name
	tableName := fixtureName

	// Insert each record individually to handle different schemas
	for _, record := range records {
		// Convert Go slices to PostgreSQL array format
		processedRecord := p.processRecordForPostgres(record)

		// Use Model with map and Table to specify table name
		_, err := p.driver.NewInsert().
			Model(&processedRecord).
			Table(tableName).
			Exec(ctx)
		if err != nil {
			// Return error instead of silently continuing
			return fmt.Errorf("failed to insert record into %s: %w", tableName, err)
		}
	}

	return nil
}

// processRecordForPostgres converts Go types to PostgreSQL-compatible formats
func (p *Provider) processRecordForPostgres(record map[string]interface{}) map[string]interface{} {
	processed := make(map[string]interface{})

	for key, value := range record {
		switch v := value.(type) {
		case []interface{}:
			// Convert slice to PostgreSQL array format using pq.Array
			if len(v) > 0 {
				// Check if it's a string array
				if _, ok := v[0].(string); ok {
					strArray := make([]string, len(v))
					for i, item := range v {
						strArray[i] = item.(string)
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

	// Query to get all table names in public schema
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

	// If no tables exist yet, nothing to clean
	if tableList == "" {
		return nil
	}

	// TRUNCATE all tables at once with RESTART IDENTITY and CASCADE
	// This is the fastest way - single atomic operation
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
