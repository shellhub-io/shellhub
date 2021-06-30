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

func TestMigration9(t *testing.T) {
	logrus.Info("Testing Migration 9 - Test if the device's name is in lowercase")

	db := dbtest.DBServer{}
	defer db.Stop()

	migrates := migrate.NewMigrate(db.Client().Database("test"), GenerateMigrations()[:8]...)
	err := migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	device := models.Device{
		Name: "Test",
	}

	_, err = db.Client().Database("test").Collection("devices").InsertOne(context.TODO(), device)
	assert.NoError(t, err)

	migrates = migrate.NewMigrate(db.Client().Database("test"), GenerateMigrations()[:9]...)
	err = migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	err = db.Client().Database("test").Collection("devices").FindOne(context.TODO(), bson.M{"name": "test"}).Decode(&device)
	assert.NoError(t, err)
}
