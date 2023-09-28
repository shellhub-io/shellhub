// fixtures provides utilities to set up a MongoDB testing environment
// using the mongotest library. This package contains predefined fixture data
// that can be loaded into the database for testing purposes.
//
// It also contains constants representing various device states and
// configurations. These constants can be used as references in tests to
// identify specific data states without directly using raw values.
package fixtures

import (
	"path/filepath"
	"runtime"

	"github.com/shellhub-io/mongotest"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
)

// Configure initializes the testing environment for MongoDB.
func Configure(db *dbtest.DBServer) {
	_, filename, _, _ := runtime.Caller(0)
	dataDir := filepath.Join(filepath.Dir(filename), "data")

	mongotest.Configure(mongotest.Config{
		URL:            "mongodb://" + db.Host,
		Database:       "test",
		FixtureRootDir: dataDir,
		FixtureFormat:  mongotest.FixtureFormatJSON,
		PreInsertFuncs: []mongotest.PreInsertFunc{
			mongotest.SimpleConvertObjID("users", "_id"),
			mongotest.SimpleConvertTime("users", "created_at"),
			mongotest.SimpleConvertTime("users", "last_login"),
			mongotest.SimpleConvertObjID("recovery_tokens", "_id"),
			mongotest.SimpleConvertTime("recovery_tokens", "created_at"),
			mongotest.SimpleConvertObjID("announcements", "_id"),
			mongotest.SimpleConvertTime("announcements", "date"),
			mongotest.SimpleConvertObjID("public_keys", "_id"),
			mongotest.SimpleConvertBytes("public_keys", "data"),
			mongotest.SimpleConvertTime("public_keys", "created_at"),
			mongotest.SimpleConvertObjID("private_keys", "_id"),
			mongotest.SimpleConvertBytes("private_keys", "data"),
			mongotest.SimpleConvertTime("private_keys", "created_at"),
			mongotest.SimpleConvertObjID("licenses", "_id"),
			mongotest.SimpleConvertBytes("licenses", "rawdata"),
			mongotest.SimpleConvertTime("licenses", "created_at"),
			mongotest.SimpleConvertObjID("namespaces", "_id"),
			mongotest.SimpleConvertTime("namespaces", "created_at"),
			mongotest.SimpleConvertObjID("devices", "_id"),
			mongotest.SimpleConvertTime("devices", "created_at"),
			mongotest.SimpleConvertTime("devices", "last_seen"),
			mongotest.SimpleConvertTime("devices", "status_updated_at"),
			mongotest.SimpleConvertTime("connected_devices", "last_seen"),
			mongotest.SimpleConvertObjID("firewall_rules", "_id"),
			mongotest.SimpleConvertObjID("sessions", "_id"),
			mongotest.SimpleConvertTime("sessions", "started_at"),
			mongotest.SimpleConvertTime("sessions", "last_seen"),
			mongotest.SimpleConvertObjID("active_sessions", "_id"),
			mongotest.SimpleConvertTime("active_sessions", "last_seen"),
			mongotest.SimpleConvertObjID("recorded_sessions", "_id"),
			mongotest.SimpleConvertTime("recorded_sessions", "time"),
		},
	})
}

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
)
