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

func TestMigration31(t *testing.T) {
	logrus.Info("Testing Migration 31 - Test whether the collection of namespaces the field created_at was created")

	db := dbtest.DBServer{}
	defer db.Stop()

	namespace := models.Namespace{
		Name: "Test",
	}

	_, err := db.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace)
	assert.NoError(t, err)

	migrations := GenerateMigrations()[30:31]

	migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
	err = migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	version, _, err := migrates.Version()
	assert.NoError(t, err)
	assert.Equal(t, uint64(31), version)

	var migratedNamespace *models.Namespace
	err = db.Client().Database("test").Collection("namespaces").FindOne(context.TODO(), bson.M{"name": namespace.Name}).Decode(&migratedNamespace)
	assert.NoError(t, err)
	assert.NotNil(t, migratedNamespace.CreatedAt)
}
