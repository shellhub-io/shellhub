package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration15(t *testing.T) {
	t.Cleanup(func() {
		assert.NoError(t, srv.Reset())
	})

	migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[:14]...)
	err := migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	ns := models.Namespace{
		Name: "Test",
	}

	_, err = c.Database("test").Collection("namespaces").InsertOne(context.TODO(), ns)
	assert.NoError(t, err)

	migrates = migrate.NewMigrate(c.Database("test"), GenerateMigrations()[:15]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	err = c.Database("test").Collection("namespaces").FindOne(context.TODO(), bson.M{"name": "test"}).Decode(&ns)
	assert.NoError(t, err)
}
