package storetest

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
)

// StoreProvider is the interface that each database backend must implement
// to provide a store instance and test utilities for the generic test suite.
//
// This abstraction allows the same test suite to run against multiple
// database implementations (MongoDB, PostgreSQL, etc.) without duplicating
// test logic.
type StoreProvider interface {
	// Store returns the store.Store instance to be tested
	Store() store.Store

	// LoadFixtures loads test data into the database.
	// The fixtures parameter contains a list of fixture names (e.g., "namespaces", "users").
	// Each provider is responsible for loading fixtures in its appropriate format.
	LoadFixtures(t *testing.T, fixtures ...string) error

	// CleanDatabase removes all data from the database.
	// This should be called before each test to ensure isolation.
	CleanDatabase(t *testing.T) error

	// Close closes the database connection and cleans up any resources.
	// This is typically called in TestMain after all tests complete.
	Close(t *testing.T) error
}

// ProviderSetup contains initialization functions for a provider
type ProviderSetup struct {
	// Setup is called once before all tests in TestMain
	Setup func(ctx context.Context) (StoreProvider, error)

	// Teardown is called once after all tests complete
	Teardown func(ctx context.Context, provider StoreProvider) error
}
