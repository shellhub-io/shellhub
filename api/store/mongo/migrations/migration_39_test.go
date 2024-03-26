package migrations

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
)

func TestMigration39(t *testing.T) {
	logrus.Info("Testing Migration 39")

	migrations := GenerateMigrations()[:39]

	migrates := migrate.NewMigrate(mongoClient.Database("test"), migrations...)
	err := migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	version, _, err := migrates.Version()
	assert.NoError(t, err)
	assert.Equal(t, uint64(39), version)
}
