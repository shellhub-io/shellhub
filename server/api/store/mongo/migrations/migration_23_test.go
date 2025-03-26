package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestMigration23(t *testing.T) {
	t.Cleanup(func() {
		assert.NoError(t, srv.Reset())
	})

	migrations := GenerateMigrations()[:22]

	migrates := migrate.NewMigrate(c.Database("test"), migrations...)
	err := migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	version, _, err := migrates.Version(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, uint64(22), version)

	namespace := models.Namespace{
		Name:     "namespace.test",
		Owner:    "owner",
		TenantID: "tenant",
	}
	_, err = c.Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace)
	assert.NoError(t, err)

	namespace = models.Namespace{
		Name:     "namespacetest",
		Owner:    "owner",
		TenantID: "tenant2",
	}
	_, err = c.Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace)
	assert.NoError(t, err)

	device := models.Device{
		Name:     "device.test",
		UID:      "uid",
		Identity: &models.DeviceIdentity{MAC: "mac"},
		TenantID: "tenant",
		LastSeen: clock.Now(),
	}
	_, err = c.Database("test").Collection("devices").InsertOne(context.TODO(), device)
	assert.NoError(t, err)

	device = models.Device{
		Name:     "devicetest",
		UID:      "uid2",
		Identity: &models.DeviceIdentity{MAC: "mac"},
		TenantID: "tenant2",
		LastSeen: clock.Now(),
	}
	_, err = c.Database("test").Collection("devices").InsertOne(context.TODO(), device)
	assert.NoError(t, err)

	migration := GenerateMigrations()[22]

	migrates = migrate.NewMigrate(c.Database("test"), migration)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	version, _, err = migrates.Version(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, uint64(23), version)

	var migratedNamespace *models.Namespace
	err = c.Database("test").Collection("namespaces").FindOne(context.TODO(), bson.M{"tenant_id": "tenant"}).Decode(&migratedNamespace)
	assert.NoError(t, err)
	assert.Equal(t, "namespace-test", migratedNamespace.Name)

	err = c.Database("test").Collection("namespaces").FindOne(context.TODO(), bson.M{"tenant_id": "tenant2"}).Decode(&migratedNamespace)
	assert.NoError(t, err)
	assert.Equal(t, "namespacetest", migratedNamespace.Name)

	var migratedDevice *models.Device
	err = c.Database("test").Collection("devices").FindOne(context.TODO(), bson.M{"tenant_id": "tenant"}).Decode(&migratedDevice)
	assert.NoError(t, err)
	assert.Equal(t, "device-test", migratedDevice.Name)

	err = c.Database("test").Collection("devices").FindOne(context.TODO(), bson.M{"tenant_id": "tenant2"}).Decode(&migratedDevice)
	assert.NoError(t, err)
	assert.Equal(t, "devicetest", migratedDevice.Name)

	err = c.Database("test").Collection("namespaces").FindOne(context.TODO(), bson.M{"name": "name.test"}).Decode(&models.Namespace{})
	assert.EqualError(t, mongo.ErrNoDocuments, err.Error())
}
