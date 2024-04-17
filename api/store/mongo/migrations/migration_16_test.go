package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
)

func TestMigration16(t *testing.T) {
	t.Cleanup(func() {
		assert.NoError(t, fixtures.Teardown())
	})

	pk1 := models.PublicKey{Fingerprint: "test"}
	pk2 := models.PublicKey{Fingerprint: "test"}

	_, err := srv.Client().Database("test").Collection("public_keys").InsertOne(context.TODO(), pk1)
	assert.NoError(t, err)

	_, err = srv.Client().Database("test").Collection("public_keys").InsertOne(context.TODO(), pk2)
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(srv.Client().Database("test"), GenerateMigrations()[:15]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	migrates = migrate.NewMigrate(srv.Client().Database("test"), GenerateMigrations()[:16]...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.Error(t, err)
}
