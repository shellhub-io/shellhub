package fixtures

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/shellhub-io/mongotest"
)

const (
	FixtureAnnouncements    = "announcements"     // Check "fixtures.data.announcements" for fixture info
	FixtureConnectedDevices = "connected_devices" // Check "fixtures.data.connected_devices" for fixture info
	FixtureDevices          = "devices"           // Check "fixtures.data.devices" for fixture info
	FixtureSessions         = "sessions"          // Check "fixtures.data.sessions" for fixture info
	FixtureActiveSessions   = "active_sessions"   // Check "fixtures.data.active_sessions" for fixture info
	FixtureRecordedSessions = "recorded_sessions" // Check "fixtures.data.recorded_sessions" for fixture info
	FixtureFirewallRules    = "firewall_rules"    // Check "fixtures.data.firewall_rules" for fixture info
	FixturePublicKeys       = "public_keys"       // Check "fixtures.data.public_keys" for fixture info
	FixturePrivateKeys      = "private_keys"      // Check "fixtures.data.private_keys" for fixture info
	FixtureLicenses         = "licenses"          // Check "fixtures.data.licenses" for fixture info
	FixtureUsers            = "users"             // Check "fixtures.data.users" for fixture iefo
	FixtureNamespaces       = "namespaces"        // Check "fixtures.data.namespaces" for fixture info
	FixtureRecoveryTokens   = "recovery_tokens"   // Check "fixtures.data.recovery_tokens" for fixture info
)

// Init configures the mongotest for the provided host's database. It is necessary
// before using any fixtures and panics if any errors arise.
func Init(host, database string) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to retrieve the fixtures path at runtime")
	}

	mongotest.Configure(mongotest.Config{
		URL:            "mongodb://" + host,
		Database:       database,
		FixtureRootDir: filepath.Join(filepath.Dir(file), "data"),
		FixtureFormat:  mongotest.FixtureFormatJSON,
		PreInsertFuncs: setupPreInsertFuncs(),
	})
}

// Apply applies 'n' fixtures in the database.
func Apply(fixtures ...string) error {
	err := mongotest.UseFixture(fixtures...)
	if err != nil {
		return err
	}

	fmt.Println("fixtures applied success:", fixtures)

	return nil
}

// Teardown resets all applied fixtures.
func Teardown() error {
	return mongotest.DropDatabase()
}
