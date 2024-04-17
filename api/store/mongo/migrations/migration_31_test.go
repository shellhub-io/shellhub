package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration31(t *testing.T) {
	t.Cleanup(func() {
		assert.NoError(t, srv.Reset())
	})

	namespace := models.Namespace{
		Name: "Test",
	}

	_, err := c.Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace)
	assert.NoError(t, err)

	migrations := GenerateMigrations()[30:31]

	migrates := migrate.NewMigrate(c.Database("test"), migrations...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	version, _, err := migrates.Version(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, uint64(31), version)

	var migratedNamespace *models.Namespace
	err = c.Database("test").Collection("namespaces").FindOne(context.TODO(), bson.M{"name": namespace.Name}).Decode(&migratedNamespace)
	assert.NoError(t, err)
	assert.NotNil(t, migratedNamespace.CreatedAt)
}
