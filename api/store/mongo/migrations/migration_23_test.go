package migrations

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/models"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestMigration23(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	migrations := GenerateMigrations()[:22]

	migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
	if err := migrates.Up(migrate.AllAvailable); err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	version, description, err := migrates.Version()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if version != 22 || description != "" {
		t.Errorf("Unexpected version/description %v %v", version, description)
		return
	}

	namespace := models.Namespace{
		Name:     "namespace.test",
		Owner:    "owner",
		TenantID: "tenant",
	}
	if _, err := db.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace); err != nil {
		t.Errorf("failed to insert document")
		return
	}

	namespace = models.Namespace{
		Name:     "namespacetest",
		Owner:    "owner",
		TenantID: "tenant2",
	}
	if _, err := db.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace); err != nil {
		t.Errorf("failed to insert document")
		return
	}

	device := models.Device{
		Name:     "device.test",
		UID:      "uid",
		Identity: &models.DeviceIdentity{MAC: "mac"},
		TenantID: "tenant",
		LastSeen: time.Now(),
	}
	if _, err := db.Client().Database("test").Collection("devices").InsertOne(context.TODO(), device); err != nil {
		t.Errorf("failed to insert document")
		return
	}

	device = models.Device{
		Name:     "devicetest",
		UID:      "uid2",
		Identity: &models.DeviceIdentity{MAC: "mac"},
		TenantID: "tenant2",
		LastSeen: time.Now(),
	}
	if _, err := db.Client().Database("test").Collection("devices").InsertOne(context.TODO(), device); err != nil {
		t.Errorf("failed to insert document")
		return
	}

	migration := GenerateMigrations()[22]

	migrates = migrate.NewMigrate(db.Client().Database("test"), migration)
	if err := migrates.Up(migrate.AllAvailable); err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	version, description, err = migrates.Version()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if version != 23 || description != "change dot in namespace name and hostname to -" {
		t.Errorf("Unexpected version/description %v %v", version, description)
		return
	}

	var migratedNamespace *models.Namespace
	if err := db.Client().Database("test").Collection("namespaces").FindOne(context.TODO(), bson.M{"tenant_id": "tenant"}).Decode(&migratedNamespace); err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if migratedNamespace.Name != "namespace-test" {
		t.Errorf("Unexpected data on namespace collection %v", migratedNamespace.Name)
		return
	}

	if err := db.Client().Database("test").Collection("namespaces").FindOne(context.TODO(), bson.M{"tenant_id": "tenant2"}).Decode(&migratedNamespace); err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if migratedNamespace.Name != "namespacetest" {
		t.Errorf("Unexpected data on namespace collection %v", migratedNamespace.Name)
		return
	}

	var migratedDevice *models.Device
	if err := db.Client().Database("test").Collection("devices").FindOne(context.TODO(), bson.M{"tenant_id": "tenant"}).Decode(&migratedDevice); err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if migratedDevice.Name != "device-test" {
		t.Errorf("Unexpected data on namespace collection %v", migratedDevice.Name)
		return
	}

	if err := db.Client().Database("test").Collection("devices").FindOne(context.TODO(), bson.M{"tenant_id": "tenant2"}).Decode(&migratedDevice); err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if migratedDevice.Name != "devicetest" {
		t.Errorf("Unexpected data on namespace collection %v", migratedDevice.Name)
		return
	}
	if err := db.Client().Database("test").Collection("namespaces").FindOne(context.TODO(), bson.M{"name": "name.test"}).Decode(&models.Namespace{}); err != mongo.ErrNoDocuments {
		t.Errorf("Unexpectedly found data from non-applied migration")
		return
	}
}
