package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
)

func TestMigration6(t *testing.T) {
	t.Cleanup(func() {
		assert.NoError(t, fixtures.Teardown())
	})

	migrates := migrate.NewMigrate(srv.Client().Database("test"), GenerateMigrations()[:5]...)
	err := migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	device1 := models.Device{
		Status: "accepted",
	}

	device2 := models.Device{
		Status: "accepted",
	}

	_, err = srv.Client().Database("test").Collection("devices").InsertOne(context.TODO(), device1)
	assert.NoError(t, err)

	_, err = srv.Client().Database("test").Collection("devices").InsertOne(context.TODO(), device2)
	assert.NoError(t, err)

	migrates = migrate.NewMigrate(srv.Client().Database("test"), GenerateMigrations()[:6]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)
}
