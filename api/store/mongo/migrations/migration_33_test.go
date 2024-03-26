package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration33(t *testing.T) {

	migrations := GenerateMigrations()[:32]

	migrates := migrate.NewMigrate(mongoClient.Database("test"), migrations...)
	err := migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	version, _, err := migrates.Version()
	assert.NoError(t, err)
	assert.Equal(t, uint64(32), version)

	device := models.Device{
		UID:      "1",
		TenantID: "tenant",
	}
	_, err = mongoClient.Database("test").Collection("devices").InsertOne(context.TODO(), &device)
	assert.NoError(t, err)

	migration := GenerateMigrations()[32:33]

	migrates = migrate.NewMigrate(mongoClient.Database("test"), migration...)
	err = migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	version, _, err = migrates.Version()
	assert.NoError(t, err)
	assert.Equal(t, uint64(33), version)

	var migratedDevice *models.Device
	err = mongoClient.Database("test").Collection("devices").FindOne(context.TODO(), bson.M{"uid": device.UID}).Decode(&migratedDevice)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(migratedDevice.Tags))
}
