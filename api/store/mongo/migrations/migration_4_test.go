package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration4(t *testing.T) {
	t.Cleanup(func() {
		assert.NoError(t, srv.Reset())
	})

	deviceInfo := models.DeviceInfo{
		ID:      "1",
		Version: "0.0.0",
	}

	device := models.Device{
		Info: &deviceInfo,
	}

	_, err := c.Database("test").Collection("devices").InsertOne(context.TODO(), device)
	assert.NoError(t, err)

	var afterMigrateDevice *models.Device
	err = c.Database("test").Collection("devices").FindOne(context.TODO(), bson.M{"info": &deviceInfo}).Decode(&afterMigrateDevice)
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[:4]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	_, err = c.Database("test").Collection("devices").InsertOne(context.TODO(), device)
	assert.NoError(t, err)

	type DeviceInfo struct {
		ID      string `json:"id"`
		Version string `json:"info.version"`
	}

	type Device struct {
		Info *DeviceInfo `json:"info"`
	}

	var migratedDevice *Device
	err = c.Database("test").Collection("devices").FindOne(context.TODO(), bson.M{"info": &deviceInfo}).Decode(&migratedDevice)
	assert.NoError(t, err)
}
