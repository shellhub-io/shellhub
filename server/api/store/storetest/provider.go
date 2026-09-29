package storetest

import (
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store"
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

	// CleanDatabase removes all data from the database.
	// This should be called before each test to ensure isolation.
	CleanDatabase(t *testing.T) error

	// Close closes the database connection and cleans up any resources.
	// This is typically called in TestMain after all tests complete.
	Close(t *testing.T) error
}
