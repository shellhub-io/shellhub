package migrations

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration19(t *testing.T) {
	t.Cleanup(func() {
		assert.NoError(t, srv.Reset())
	})

	migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[:19]...)
	err := migrates.Up(context.Background(), migrate.AllAvailable)
	assert.NoError(t, err)

	type PublicKeyFields struct {
		Name     string `json:"name"`
		Username string `json:"username" bson:"username,omitempty" validate:"regexp"`
		Hostname string `json:"hostname" bson:"hostname" validate:"regexp"`
	}

	type PublicKey struct {
		Data            []byte    `json:"data"`
		Fingerprint     string    `json:"fingerprint"`
		CreatedAt       time.Time `json:"created_at" bson:"created_at"`
		TenantID        string    `json:"tenant_id" bson:"tenant_id"`
		PublicKeyFields `bson:",inline"`
	}

	pk := PublicKey{
		Data:            []byte("teste"),
		TenantID:        "tenant",
		PublicKeyFields: PublicKeyFields{Name: "teste1", Hostname: ".*"},
	}

	_, err = c.Database("test").Collection("public_keys").InsertOne(context.TODO(), pk)
	assert.NoError(t, err)

	err = c.Database("test").Collection("public_keys").FindOne(context.TODO(), bson.M{"tenant_id": "tenant"}).Decode(&pk)
	assert.NoError(t, err)
	assert.Equal(t, pk.Fingerprint, "")
}
