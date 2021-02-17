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

func TestMigration39(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	logrus.Info("Testing Migration 39 - Test if the token field was created on namespaces collection")

	migrates := migrate.NewMigrate(db.Client().Database("test"), GenerateMigrations()[:36]...)
	err := migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	version, _, err := migrates.Version()
	assert.NoError(t, err)
	assert.Equal(t, uint64(36), version)

	namespace := models.Namespace{
		Name:     "namespace.test",
		Owner:    "owner",
		TenantID: "tenant",
	}
	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace)
	assert.NoError(t, err)

	migrates = migrate.NewMigrate(db.Client().Database("test"), GenerateMigrations()[:39]...)
	err = migrates.Up(migrate.AllAvailable)
	assert.NoError(t, err)

	version, _, err = migrates.Version()
	assert.NoError(t, err)
	assert.Equal(t, uint64(39), version)

	assert.Equal(t, len(namespace.APITokens), 0)

	APIToken := models.Token{
		ID:       "1",
		TenantID: "tenant",
		ReadOnly: true,
	}

	namespace.APITokens = append(namespace.APITokens, APIToken)

	assert.Equal(t, len(namespace.APITokens), 1)
	assert.Equal(t, namespace.APITokens[0], APIToken)
}
