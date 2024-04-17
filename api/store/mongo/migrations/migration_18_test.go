package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration18(t *testing.T) {
	t.Cleanup(func() {
		assert.NoError(t, fixtures.Teardown())
	})

	namespace := models.Namespace{
		Name:     "name",
		Owner:    "60df59bc65f88d92b974a60f",
		TenantID: "tenant",
	}

	migrations := GenerateMigrations()[:17]

	migrates := migrate.NewMigrate(srv.Client().Database("test"), migrations...)
	err := migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	_, err = srv.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace)
	assert.NoError(t, err)

	migrates = migrate.NewMigrate(srv.Client().Database("test"), GenerateMigrations()[17])
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	err = srv.Client().Database("test").Collection("namespaces").FindOne(context.TODO(), bson.M{"tenant_id": "tenant"}).Decode(&namespace)
	assert.NoError(t, err)
	assert.Equal(t, namespace.MaxDevices, 3)
}
