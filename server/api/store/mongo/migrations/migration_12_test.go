package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
)

func TestMigration12(t *testing.T) {
	t.Cleanup(func() {
		assert.NoError(t, srv.Reset())
	})

	ns1 := models.Namespace{Name: "name", TenantID: "1"}
	ns2 := models.Namespace{Name: "name", TenantID: "1"}

	_, err := c.Database("test").Collection("namespaces").InsertOne(context.TODO(), ns1)
	assert.NoError(t, err)

	_, err = c.Database("test").Collection("namespaces").InsertOne(context.TODO(), ns2)
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[:11]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	migrates = migrate.NewMigrate(c.Database("test"), GenerateMigrations()[:12]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.Error(t, err)
}
