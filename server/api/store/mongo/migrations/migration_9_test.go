package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration9(t *testing.T) {
	t.Cleanup(func() {
		assert.NoError(t, srv.Reset())
	})

	migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[:8]...)
	err := migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	device := models.Device{
		Name: "Test",
	}

	_, err = c.Database("test").Collection("devices").InsertOne(context.TODO(), device)
	assert.NoError(t, err)

	migrates = migrate.NewMigrate(c.Database("test"), GenerateMigrations()[:9]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	err = c.Database("test").Collection("devices").FindOne(context.TODO(), bson.M{"name": "test"}).Decode(&device)
	assert.NoError(t, err)
}
