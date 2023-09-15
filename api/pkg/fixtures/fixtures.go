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

	"github.com/pinzolo/mongotest"
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
			mongotest.SimpleConvertTime("namespaces", "created_at"),
			mongotest.SimpleConvertTime("devices", "created_at"),
			mongotest.SimpleConvertTime("devices", "last_seen"),
			mongotest.SimpleConvertTime("devices", "status_updated_at"),
			mongotest.SimpleConvertTime("connected_devices", "last_seen"),
		},
	})
}

const (
	// DeviceAccepted a device with status `accepted`.
	//
	//  ID: `6500c8f026e1e911042ee820`
	//  UID: `2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c`
	//  Name: `hostname_accepted`
	//  TenantID: `00000000-0000-4000-0000-000000000000`
	DeviceAccepted = "device_accepted"
	// DevicePending represents a device with status `pending`.
	//
	//  ID: `7580b9e026d2d913052fe821`
	//  UID: `3400340f4db3f748747c5d136e2346279025976ec6315c7e226487ddf99019d`
	//  Name: `hostname_pending`
	//  TenantID: `00000000-0000-4000-0000-000000000000`
	DevicePending = "device_pending"
	// DeviceWithTag represents a device with a specific tag.
	//
	//  ID: `8591c9f137d3e914063ff823`
	//  UID: `4500450g5dc4g859858d6e247f3457380136087fd7426d8f337598eef0a120e`
	//  Name: `hostname_with_tag`
	//  TenantID: `00000000-0000-4000-0000-000000000000`
	//  Tags: `["device1"]`
	DeviceWithTag = "device_with_tag"
	// DeviceWithInvalidTenant represents a device with an invalid tenant ID.
	//
	//  ID: `device_with_invalid_id`
	//  UID: `5600560h6ed5h960969e7f358g4568491247198ge8537e9g448609fff1b231f`
	//  Name: `hostname_with_invalid_tenant`
	//  TenantID: `invalid_tenant_id`
	DeviceWithInvalidTenant = "device_with_invalid_tenant_id"
	// DeviceSessions populate 20 different sessions with `device_id` equals to uid1,
	// uid2 and uid3.
	DeviceSessions = "device_sessions"
	// NamespaceUnlimited represents a namespace with unlimited devices.
	//
	//  ID: `device_with_invalid_id`
	//  UID: `5600560h6ed5h960969e7f358g4568491247198ge8537e9g448609fff1b231f`
	//  Name: `hostname_with_invalid_tenant`
	//  TenantID: `invalid_tenant_id`
	NamespaceUnlimited = "namespace_unlimited"
)
