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

	db := dbtest.DB{}
	err := func() error {
		err := db.Down(context.Background())

		return err
	}()
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(db.MongoClient.Database("test"), GenerateMigrations()[:8]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	device := models.Device{
		Name: "Test",
	}

	_, err = mongoClient.Database("test").Collection("devices").InsertOne(context.TODO(), device)
	assert.NoError(t, err)

	migrates = migrate.NewMigrate(mongoClient.Database("test"), GenerateMigrations()[:9]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	err = mongoClient.Database("test").Collection("devices").FindOne(context.TODO(), bson.M{"name": "test"}).Decode(&device)
	assert.NoError(t, err)
}
