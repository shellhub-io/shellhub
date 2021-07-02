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

func TestMigration22(t *testing.T) {
	logrus.Info("Testing Migration 22 - Test if the user was added to membres group for the namespace")

	db := dbtest.DBServer{}
	defer db.Stop()

	user := models.User{
		ID: "1",
	}

	ns := models.Namespace{
		Name:       "namespace",
		Owner:      "60df59bc65f88d92b974a60f",
		TenantID:   "tenant",
		Members:    []interface{}{"60df59bc65f88d92b974a60f"},
		MaxDevices: -1,
	}
	_, err := db.Client().Database("test").Collection("devices").InsertOne(context.TODO(), user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), ns)
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(db.Client().Database("test"), GenerateMigrations()[21:22]...)
	err = migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	var migratedNamespace *models.Namespace
	err = db.Client().Database("test").Collection("namespaces").FindOne(context.TODO(), bson.M{"tenant_id": "tenant"}).Decode(&migratedNamespace)
	assert.NoError(t, err)
}
