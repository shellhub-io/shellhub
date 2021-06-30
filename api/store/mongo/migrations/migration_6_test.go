package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
)

func TestMigration6(t *testing.T) {
	logrus.Info("Testing Migration 6 - Test if the status is not unique")

	db := dbtest.DBServer{}
	defer db.Stop()

	migrates := migrate.NewMigrate(db.Client().Database("test"), GenerateMigrations()[:5]...)
	err := migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	device1 := models.Device{
		Status: "accepted",
	}

	device2 := models.Device{
		Status: "accepted",
	}

	_, err = db.Client().Database("test").Collection("devices").InsertOne(context.TODO(), device1)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("devices").InsertOne(context.TODO(), device2)
	assert.NoError(t, err)

	migrates = migrate.NewMigrate(db.Client().Database("test"), GenerateMigrations()[:6]...)
	err = migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)
}
