package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration19(t *testing.T) {
	logrus.Info("Testing Migration 19 - Test if the fingerprint is removed")

	db := dbtest.DBServer{}
	defer db.Stop()

	migrates := migrate.NewMigrate(db.Client().Database("test"), GenerateMigrations()[:19]...)
	err := migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	pk := models.PublicKey{
		Data:            []byte("teste"),
		TenantID:        "tenant",
		PublicKeyFields: models.PublicKeyFields{Name: "teste1", Hostname: ".*"},
	}

	_, err = db.Client().Database("test").Collection("public_keys").InsertOne(context.TODO(), pk)
	assert.NoError(t, err)

	err = db.Client().Database("test").Collection("public_keys").FindOne(context.TODO(), bson.M{"tenant_id": "tenant"}).Decode(&pk)
	assert.NoError(t, err)
	assert.Equal(t, pk.Fingerprint, "")
}
