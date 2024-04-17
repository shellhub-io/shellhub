package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
)

func TestMigration1(t *testing.T) {
	t.Cleanup(func() {
		assert.NoError(t, fixtures.Teardown())
	})

	migrates := migrate.NewMigrate(srv.Client().Database("test"), GenerateMigrations()[:1]...)
	err := migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)
}
