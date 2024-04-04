package mongo_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/shellhub-io/mongotest"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	shstore "github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	"github.com/shellhub-io/shellhub/pkg/cache"
	log "github.com/sirupsen/logrus"
)

var db = &dbtest.DB{}
var store shstore.Store

const (
	fixtureAnnouncements    = "announcements"     // Check "fixtures.data.announcements" for fixture info
	fixtureConnectedDevices = "connected_devices" // Check "fixtures.data.connected_devices" for fixture info
	fixtureDevices          = "devices"           // Check "fixtures.data.devices" for fixture info
	fixtureSessions         = "sessions"          // Check "fixtures.data.sessions" for fixture info
	fixtureActiveSessions   = "active_sessions"   // Check "fixtures.data.active_sessions" for fixture info
	fixtureRecordedSessions = "recorded_sessions" // Check "fixtures.data.recorded_sessions" for fixture info
	fixtureFirewallRules    = "firewall_rules"    // Check "fixtures.data.firewall_rules" for fixture info
	fixturePublicKeys       = "public_keys"       // Check "fixtures.data.public_keys" for fixture info
	fixturePrivateKeys      = "private_keys"      // Check "fixtures.data.private_keys" for fixture info
	fixtureLicenses         = "licenses"          // Check "fixtures.data.licenses" for fixture info
	fixtureUsers            = "users"             // Check "fixtures.data.users" for fixture iefo
	fixtureNamespaces       = "namespaces"        // Check "fixtures.data.namespaces" for fixture info
	fixtureRecoveryTokens   = "recovery_tokens"   // Check "fixtures.data.recovery_tokens" for fixture info
)

func TestMain(m *testing.M) {
	log.Info("Starting mongo tests")
	defer log.Info("Stopping mongo tests")

	ctx := context.Background()

	db.Database = "mongo"
	_, file, _, _ := runtime.Caller(0)
	db.Fixtures.Root = filepath.Join(filepath.Dir(file), "fixtures")
	db.Fixtures.PreInsertFuncs = []mongotest.PreInsertFunc{
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
	}

	if err := db.Up(ctx); err != nil {
		log.WithError(err).Error("Failed to UP the mongodb container")
		os.Exit(1)
	}

	defer func() {
		if err := db.Down(ctx); err != nil {
			log.WithError(err).Error("Failed to DOWN the mongodb container")
			os.Exit(1)
		}
	}()

	log.Info("Connecting to ", db.URI+"/"+db.Database)

	_, mdb, err := mongo.Connect(ctx, db.URI+"/"+db.Database)
	if err != nil {
		log.WithError(err).Error("Failed to connect to mongodb")
		os.Exit(1)
	}

	store, err = mongo.NewStore(ctx, mdb, cache.NewNullCache())
	if err != nil {
		log.WithError(err).Error("Failed to create the mongodb store")
		os.Exit(1)
	}

	code := m.Run()
	os.Exit(code)
	// os.Exit(0)
}
