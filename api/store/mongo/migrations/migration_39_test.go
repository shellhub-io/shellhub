package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
)

func TestMigration39(t *testing.T) {
	logrus.Info("Testing Migration 39")

	db := dbtest.DB{}
	err := func() error {
		err := db.Down(context.Background())

		return err
	}()
	assert.NoError(t, err)

	migrations := GenerateMigrations()[:39]

	migrates := migrate.NewMigrate(mongoClient.Database("test"), migrations...)
	err = migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	version, _, err := migrates.Version(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, uint64(39), version)
}
