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

func TestMigration30(t *testing.T) {
	logrus.Info("Testing Migration 30 - Test whether the collection of devices the field remote_addr was created")

	db := dbtest.DBServer{}
	defer db.Stop()

	device := models.Device{
		UID: "1",
	}

	_, err := db.Client().Database("test").Collection("devices").InsertOne(context.TODO(), device)
	assert.NoError(t, err)

	migrations := GenerateMigrations()[29:30]

	migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
	err = migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	version, _, err := migrates.Version()
	assert.NoError(t, err)
	assert.Equal(t, uint64(30), version)

	var migratedDevice *models.Device
	err = db.Client().Database("test").Collection("devices").FindOne(context.TODO(), bson.M{"uid": device.UID}).Decode(&migratedDevice)
	assert.NoError(t, err)
	assert.Equal(t, migratedDevice.RemoteAddr, "")
}
