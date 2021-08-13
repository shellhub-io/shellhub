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

func TestMigration28(t *testing.T) {
	logrus.Info("Testing Migration 28 - Test whether the collection of users and devices the field created_at was created")

	db := dbtest.DBServer{}
	defer db.Stop()

	user := models.User{
		Name: "Test",
	}

	device := models.Device{
		UID: "1",
	}

	_, err := db.Client().Database("test").Collection("users").InsertOne(context.TODO(), user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("devices").InsertOne(context.TODO(), device)
	assert.NoError(t, err)

	migrations := GenerateMigrations()[27:28]

	migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
	err = migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	version, _, err := migrates.Version()
	assert.NoError(t, err)
	assert.Equal(t, uint64(28), version)

	var migratedUser *models.User
	err = db.Client().Database("test").Collection("users").FindOne(context.TODO(), bson.M{"name": user.Name}).Decode(&migratedUser)
	assert.NoError(t, err)
	assert.NotNil(t, migratedUser.CreatedAt)

	var migratedDevice *models.Device
	err = db.Client().Database("test").Collection("devices").FindOne(context.TODO(), bson.M{"uid": device.UID}).Decode(&migratedDevice)
	assert.NoError(t, err)
	assert.NotNil(t, migratedDevice.CreatedAt)
}
