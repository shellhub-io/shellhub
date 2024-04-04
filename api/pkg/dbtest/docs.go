// Package dbtest provides utilities for setting up MongoDB test environments,
// including container initialization, fixture management, and database reset functionality.
//
// Usage:
//
//	ctx := context.Background()
//
//	// Initialize a new MongoDB test server
//	server := &dbtest.DB{}
//	server.Database = "test"
//	server.Fixtures.Root = "/path/to/fixtures"
//	server.Fixtures.PreInsertFuncs = []mongotest.PreInsertFunc{
//	    mongotest.SimpleConvertObjID("users", "_id"),
//	    // [...]
//	}
//
//	// Start the MongoDB container and configure fixtures
//	if err := server.Up(ctx); err != nil {
//	    log.Fatalf("Error starting MongoDB container: %v", err)
//	}
//
//	defer func() {
//	    if err := server.Down(ctx); err != nil {
//	        log.Fatalf("Error stopping MongoDB container: %v", err)
//	    }
//	}()
//
//	// Apply fixtures
//	if err := server.Apply("fixture1.json", "fixture2.json"); err != nil {
//	    log.Fatalf("Error applying fixtures: %v", err)
//	}
//
//	// Reset the database
//	if err := server.Reset(); err != nil {
//	    log.Fatalf("Error resetting database: %v", err)
//	}
package dbtest
