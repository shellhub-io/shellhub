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

	db := dbtest.DB{}
	err := func() error {
		err := db.Down(context.Background())

		return err
	}()
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(db.MongoClient.Database("test"), GenerateMigrations()[:5]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	device1 := models.Device{
		Status: "accepted",
	}

	device2 := models.Device{
		Status: "accepted",
	}

	_, err = mongoClient.Database("test").Collection("devices").InsertOne(context.TODO(), device1)
	assert.NoError(t, err)

	_, err = mongoClient.Database("test").Collection("devices").InsertOne(context.TODO(), device2)
	assert.NoError(t, err)

	migrates = migrate.NewMigrate(mongoClient.Database("test"), GenerateMigrations()[:6]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)
}
