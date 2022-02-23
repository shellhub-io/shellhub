package migrations

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration42(t *testing.T) {
	logrus.Info("Testing Migration 42")

	db := dbtest.DBServer{}
	defer db.Stop()

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

	keyOld := PublicKey{
		Fingerprint: "fingerprint",
		TenantID:    "tenant",
		PublicKeyFields: PublicKeyFields{
			Name:     "key",
			Username: ".*",
			Hostname: ".*",
		},
	}

	keyNew := models.PublicKey{
		Fingerprint: "fingerprint",
		TenantID:    "tenant",
		PublicKeyFields: models.PublicKeyFields{
			Name:     "key",
			Username: ".*",
			Filter: models.PublicKeyFilter{
				Hostname: ".*",
			},
		},
	}

	_, err := db.Client().Database("test").Collection("public_keys").InsertOne(context.TODO(), keyOld)
	assert.NoError(t, err)

	cases := []struct {
		description string
		Test        func(t *testing.T)
	}{
		{
			"Success to apply up on migration 42",
			func(t *testing.T) {
				t.Helper()

				migrations := GenerateMigrations()[41:42]
				migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
				err := migrates.Up(migrate.AllAvailable)
				assert.NoError(t, err)

				key := new(models.PublicKey)
				result := db.Client().Database("test").Collection("public_keys").FindOne(context.TODO(), bson.M{"tenant_id": keyOld.TenantID})
				assert.NoError(t, result.Err())

				err = result.Decode(key)
				assert.NoError(t, err)

				assert.Equal(t, keyNew, *key)
			},
		},
		{
			"Success to apply down on migration 42",
			func(t *testing.T) {
				t.Helper()

				migrations := GenerateMigrations()[41:42]
				migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
				err := migrates.Down(migrate.AllAvailable)
				assert.NoError(t, err)

				key := new(PublicKey)
				result := db.Client().Database("test").Collection("public_keys").FindOne(context.TODO(), bson.M{"tenant_id": keyNew.TenantID})
				assert.NoError(t, result.Err())

				err = result.Decode(key)
				assert.NoError(t, err)

				assert.Equal(t, keyOld, *key)
			},
		},
	}

	for _, test := range cases {
		tc := test
		t.Run(tc.description, tc.Test)
	}
}
