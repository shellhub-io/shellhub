package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration3(t *testing.T) {
	t.Cleanup(func() {
		assert.NoError(t, fixtures.Teardown())
	})

	type Device struct {
		Attributes *models.DeviceInfo `json:"attributes"`
	}

	device := Device{
		Attributes: &models.DeviceInfo{
			ID: "1",
		},
	}

	_, err := srv.Client().Database("test").Collection("devices").InsertOne(context.TODO(), device)
	assert.NoError(t, err)

	var afterMigrateDevice *models.Session
	err = srv.Client().Database("test").Collection("devices").FindOne(context.TODO(), bson.M{"attributes": &models.DeviceInfo{ID: "1"}}).Decode(&afterMigrateDevice)
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(srv.Client().Database("test"), GenerateMigrations()[:3]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	var migratedDevice *models.Device
	err = srv.Client().Database("test").Collection("devices").FindOne(context.TODO(), bson.M{"info": &models.DeviceInfo{ID: "1"}}).Decode(&migratedDevice)
	assert.NoError(t, err)
}
