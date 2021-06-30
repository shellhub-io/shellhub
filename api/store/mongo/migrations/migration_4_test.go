package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration4(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	logrus.Info("Testing Migration 4 - Test if the column version was renamed to info.version")

	deviceInfo := models.DeviceInfo{
		ID:      "1",
		Version: "0.0.0",
	}

	device := models.Device{
		Info: &deviceInfo,
	}

	_, err := db.Client().Database("test").Collection("devices").InsertOne(context.TODO(), device)
	assert.NoError(t, err)

	var afterMigrateDevice *models.Device
	err = db.Client().Database("test").Collection("devices").FindOne(context.TODO(), bson.M{"info": &deviceInfo}).Decode(&afterMigrateDevice)
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(db.Client().Database("test"), GenerateMigrations()[:4]...)
	err = migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("devices").InsertOne(context.TODO(), device)
	assert.NoError(t, err)

	type DeviceInfo struct {
		ID      string `json:"id"`
		Version string `json:"info.version"`
	}

	type Device struct {
		Info *DeviceInfo `json:"info"`
	}

	var migratedDevice *Device
	err = db.Client().Database("test").Collection("devices").FindOne(context.TODO(), bson.M{"info": &deviceInfo}).Decode(&migratedDevice)
	assert.NoError(t, err)
}
