package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
)

func TestMigration16(t *testing.T) {
	logrus.Info("Testing Migration 16 - Test if the fingerprint is set unique")

	db := dbtest.DBServer{}
	defer db.Stop()

	pk1 := models.PublicKey{Fingerprint: "test"}
	pk2 := models.PublicKey{Fingerprint: "test"}

	_, err := db.Client().Database("test").Collection("public_keys").InsertOne(context.TODO(), pk1)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("public_keys").InsertOne(context.TODO(), pk2)
	assert.NoError(t, err)

	migrates := migrate.NewMigrate(db.Client().Database("test"), GenerateMigrations()[:15]...)
	err = migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	migrates = migrate.NewMigrate(db.Client().Database("test"), GenerateMigrations()[:16]...)
	err = migrates.Up(migrate.AllAvailable)
	assert.Error(t, err)
}
