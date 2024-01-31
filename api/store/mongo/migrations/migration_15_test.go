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

func TestMigration15(t *testing.T) {
	logrus.Info("Testing Migration 15 - Test if the name is in lowercase")

	db := dbtest.DB{}
	err := func() error {
		err := db.Down(context.Background())

		return err
	}()
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(db.MongoClient.Database("test"), GenerateMigrations()[:14]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	ns := models.Namespace{
		Name: "Test",
	}

	_, err = mongoClient.Database("test").Collection("namespaces").InsertOne(context.TODO(), ns)
	assert.NoError(t, err)

	migrates = migrate.NewMigrate(mongoClient.Database("test"), GenerateMigrations()[:15]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	err = mongoClient.Database("test").Collection("namespaces").FindOne(context.TODO(), bson.M{"name": "test"}).Decode(&ns)
	assert.NoError(t, err)
}
