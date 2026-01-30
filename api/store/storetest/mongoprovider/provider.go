package mongoprovider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"go.mongodb.org/mongo-driver/bson"
	mongodb "go.mongodb.org/mongo-driver/mongo"
	mongooptions "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"gopkg.in/yaml.v3"
)

// Provider implements storetest.StoreProvider for MongoDB
type Provider struct {
	srv         *dbtest.Server
	store       store.Store
	db          *mongodb.Database
	fixtureRoot string
}

// NewProvider creates a new MongoDB test provider
func NewProvider(ctx context.Context) (*Provider, error) {
	srv := &dbtest.Server{}
	srv.Container.Database = "test"

	if err := srv.Up(ctx); err != nil {
		return nil, err
	}

	connectionString := srv.Container.ConnectionString + "/" + srv.Container.Database

	// Configure MongoDB client with test-specific options for consistency
	// WMajority: ensures writes are durable before returning (prevents flaky tests)
	// Primary: ensures read-after-write consistency in tests
	clientOpts := mongooptions.Client().
		SetWriteConcern(&writeconcern.WriteConcern{W: "majority"}).
		SetReadPreference(readpref.Primary())

	// Create store with test-specific client options
	st, err := mongo.NewStoreWithClientOpts(ctx, connectionString, cache.NewNullCache(), clientOpts)
	if err != nil {
		srv.Down(ctx)

		return nil, err
	}

	// Get database for direct access
	mongoStore := st.(*mongo.Store)
	db := mongoStore.GetDB()

	// Get generic fixtures path
	_, file, _, _ := runtime.Caller(0)
	// Navigate from storetest/mongoprovider/provider.go to storetest/fixtures
	fixturesPath := filepath.Join(filepath.Dir(file), "..", "fixtures")

	return &Provider{
		srv:         srv,
		store:       st,
		db:          db,
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
		yamlPath := filepath.Join(p.fixtureRoot, fixtureName+".yml")

		// Read YAML file
		data, err := os.ReadFile(yamlPath)
		if err != nil {
			return fmt.Errorf("failed to read fixture %s: %w", fixtureName, err)
		}

		// Parse YAML into array of records
		var records []map[string]interface{}
		if err := yaml.Unmarshal(data, &records); err != nil {
			return fmt.Errorf("failed to parse fixture %s: %w", fixtureName, err)
		}

		if len(records) == 0 {
			t.Logf("Fixture %s is empty, skipping", fixtureName)

			continue
		}

		t.Logf("Loading %d records from fixture %s", len(records), fixtureName)

		// Determine collection name and primary key based on fixture name
		collectionName, primaryKey := getCollectionInfo(fixtureName)

		// Insert records
		if err := p.insertRecords(ctx, collectionName, primaryKey, records); err != nil {
			return fmt.Errorf("failed to insert fixture %s: %w", fixtureName, err)
		}

		t.Logf("Successfully loaded fixture %s into collection %s", fixtureName, collectionName)
	}

	return nil
}

// CleanDatabase removes all data from all collections
// Uses deleteMany for each collection to preserve indexes
func (p *Provider) CleanDatabase(t *testing.T) error {
	t.Helper()
	ctx := context.Background()

	// List all collection names in the database
	collections, err := p.db.ListCollectionNames(ctx, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	// Delete all documents from each collection
	// This is fast and preserves indexes (unlike Drop)
	for _, collName := range collections {
		result, err := p.db.Collection(collName).DeleteMany(ctx, map[string]interface{}{})
		if err != nil {
			return fmt.Errorf("failed to clean collection %s: %w", collName, err)
		}
		t.Logf("CleanDatabase: deleted %d documents from collection %s", result.DeletedCount, collName)
	}

	return nil
}

// Close closes the MongoDB connection and stops the container
func (p *Provider) Close(t *testing.T) error {
	t.Helper()
	ctx := context.Background()

	return p.srv.Down(ctx)
}

// getCollectionInfo returns collection name and primary key field for a fixture
func getCollectionInfo(fixtureName string) (collectionName string, primaryKey string) {
	// Map fixture names to collection names and primary keys
	// Most fixtures use "id" as primary key, but some use different fields
	fixtureMap := map[string]struct {
		collection string
		primaryKey string
	}{
		"users":                  {"users", "id"},
		"namespaces":             {"namespaces", "id"},
		"devices":                {"devices", "id"},
		"sessions":               {"sessions", "id"},
		"active_sessions":        {"active_sessions", "session_id"},
		"tags":                   {"tags", "id"},
		"api_keys":               {"api_keys", "key_digest"},
		"public_keys":            {"public_keys", "fingerprint"},
		"private_keys":           {"private_keys", "fingerprint"},
		"user_invitations":       {"user_invitations", "email"},
		"membership_invitations": {"membership_invitations", "id"},
		"memberships":            {"memberships", "id"},
		"device_tags":            {"device_tags", "id"},
		"public_key_tags":        {"public_key_tags", "id"},
	}

	if info, ok := fixtureMap[fixtureName]; ok {
		return info.collection, info.primaryKey
	}

	// Default: fixture name is collection name, "id" is primary key
	return fixtureName, "id"
}

// insertRecords inserts records into MongoDB collection
// MongoDB driver handles string to ObjectID conversion automatically
func (p *Provider) insertRecords(ctx context.Context, collectionName, primaryKey string, records []map[string]interface{}) error {
	collection := p.db.Collection(collectionName)

	for _, record := range records {
		// Prepare document for MongoDB
		doc := bson.M{}

		for key, value := range record {
			// Handle primary key field -> _id
			if key == primaryKey {
				doc["_id"] = value // MongoDB driver converts string to ObjectID automatically

				continue
			}

			// Convert timestamps to time.Time
			if str, ok := value.(string); ok {
				if t, err := time.Parse(time.RFC3339, str); err == nil {
					doc[key] = t

					continue
				}
			}

			// Everything else goes as-is
			doc[key] = value
		}

		// Insert document
		_, err := collection.InsertOne(ctx, doc)
		if err != nil {
			return fmt.Errorf("failed to insert record into %s: %w", collectionName, err)
		}
	}

	return nil
}
