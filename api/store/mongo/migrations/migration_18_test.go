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

func TestMigration18(t *testing.T) {
	logrus.Info("Testing Migration 18 - Test if the max_devices is 3")

	db := dbtest.DBServer{}
	defer db.Stop()

	namespace := models.Namespace{
		Name:     "name",
		Owner:    "60df59bc65f88d92b974a60f",
		TenantID: "tenant",
	}

	migrations := GenerateMigrations()[:17]

	migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
	err := migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace)
	assert.NoError(t, err)

	migrates = migrate.NewMigrate(db.Client().Database("test"), GenerateMigrations()[17])
	err = migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	err = db.Client().Database("test").Collection("namespaces").FindOne(context.TODO(), bson.M{"tenant_id": "tenant"}).Decode(&namespace)
	assert.NoError(t, err)
	assert.Equal(t, namespace.MaxDevices, 3)
}
