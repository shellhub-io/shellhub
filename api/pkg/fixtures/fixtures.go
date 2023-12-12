package fixtures

import (
	"path/filepath"
	"runtime"

	"github.com/shellhub-io/mongotest"
)

const (
	Announcement        = "announcement"          // Check "fixtures.data.announcement" for fixture info
	Device              = "device"                // Device with status "accepted". Check "fixtures.data.device" for fixture info
	DevicePending       = "device_pending"        // Device with status "pending". Check "fixtures.data.device_pending" for fixture info
	DeviceInvalidTenant = "device_invalid_tenant" // Device with tenant "invalid". Check "fixtures.data.device_invalid_tenant" for fixture info
	Session             = "session"               // Check "fixtures.data.session" for fixture info
	FirewallRule        = "firewall_rule"         // Check "fixtures.data.firewall_rule" for fixture info
	PublicKey           = "public_key"            // Check "fixtures.data.public_key" for fixture info
	PrivateKey          = "private_key"           // Check "fixtures.data.private_key" for fixture info
	License             = "license"               // Check "fixtures.data.license" for fixture info
	User                = "user"                  // Check "fixtures.data.user" for fixture info
	Member              = "member"                // Check "fixtures.data.member" for fixture info
	Namespace           = "namespace"             // Check "fixtures.data.namespace" for fixture info
	RecordedSessions    = "recorded_sessions"
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
	return mongotest.UseFixture(fixtures...)
}

// Teardown resets all applied fixtures.
func Teardown() error {
	return mongotest.DropDatabase()
}
