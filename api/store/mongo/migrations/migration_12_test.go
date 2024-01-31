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

func TestMigration12(t *testing.T) {
	logrus.Info("Testing Migration 12 - Test if the tenant_id is set unique")

	db := dbtest.DB{}
	err := func() error {
		err := db.Down(context.Background())

		return err
	}()
	assert.NoError(t, err)

	ns1 := models.Namespace{Name: "name", TenantID: "1"}
	ns2 := models.Namespace{Name: "name", TenantID: "1"}

	_, err = mongoClient.Database("test").Collection("namespaces").InsertOne(context.TODO(), ns1)
	assert.NoError(t, err)

	_, err = mongoClient.Database("test").Collection("namespaces").InsertOne(context.TODO(), ns2)
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(mongoClient.Database("test"), GenerateMigrations()[:11]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	migrates = migrate.NewMigrate(mongoClient.Database("test"), GenerateMigrations()[:12]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.Error(t, err)
}
