package migrations

import (
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
)

func TestMigration1(t *testing.T) {
	logrus.Info("Testing Migration 1 - Create the database for the system")

	db := dbtest.DBServer{}
	defer db.Stop()

	migrates := migrate.NewMigrate(db.Client().Database("test"), GenerateMigrations()[:1]...)
	err := migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)
}
