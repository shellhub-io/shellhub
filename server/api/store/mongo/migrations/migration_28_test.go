package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration28(t *testing.T) {
	t.Cleanup(func() {
		assert.NoError(t, srv.Reset())
	})

	user := models.User{
		UserData: models.UserData{
			Name: "Test",
		},
	}

	device := models.Device{
		UID: "1",
	}

	_, err := c.Database("test").Collection("users").InsertOne(context.TODO(), user)
	assert.NoError(t, err)

	_, err = c.Database("test").Collection("devices").InsertOne(context.TODO(), device)
	assert.NoError(t, err)

	migrations := GenerateMigrations()[27:28]

	migrates := migrate.NewMigrate(c.Database("test"), migrations...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	version, _, err := migrates.Version(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, uint64(28), version)

	var migratedUser *models.User
	err = c.Database("test").Collection("users").FindOne(context.TODO(), bson.M{"name": user.Name}).Decode(&migratedUser)
	assert.NoError(t, err)
	assert.NotNil(t, migratedUser.CreatedAt)

	var migratedDevice *models.Device
	err = c.Database("test").Collection("devices").FindOne(context.TODO(), bson.M{"uid": device.UID}).Decode(&migratedDevice)
	assert.NoError(t, err)
	assert.NotNil(t, migratedDevice.CreatedAt)
}
