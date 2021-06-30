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

func TestMigration3(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	logrus.Info("Testing Migration 3 - Test if the column attributes was renamed to info")

	type Device struct {
		Attributes *models.DeviceInfo `json:"attributes"`
	}

	device := Device{
		Attributes: &models.DeviceInfo{
			ID: "1",
		},
	}

	_, err := db.Client().Database("test").Collection("devices").InsertOne(context.TODO(), device)
	assert.NoError(t, err)

	var afterMigrateDevice *models.Session
	err = db.Client().Database("test").Collection("devices").FindOne(context.TODO(), bson.M{"attributes": &models.DeviceInfo{ID: "1"}}).Decode(&afterMigrateDevice)
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(db.Client().Database("test"), GenerateMigrations()[:3]...)
	err = migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	var migratedDevice *models.Device
	err = db.Client().Database("test").Collection("devices").FindOne(context.TODO(), bson.M{"info": &models.DeviceInfo{ID: "1"}}).Decode(&migratedDevice)
	assert.NoError(t, err)
}
