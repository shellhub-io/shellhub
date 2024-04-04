package migrations

import (
	"context"
	"sort"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration46(t *testing.T) {
	logrus.Info("Testing Migration 46")

	db := dbtest.DB{}
	defer db.Stop()

	keyUsernameEmpty := &models.PublicKey{
		Fingerprint: "fingerprint",
		TenantID:    "tenant",
		PublicKeyFields: models.PublicKeyFields{
			Name:     "key",
			Username: "",
			Filter: models.PublicKeyFilter{
				Tags: []string{"tag1", "tag2", "tag3"},
			},
		},
	}

	keyUsernameRegexp := &models.PublicKey{
		Fingerprint: "fingerprint",
		TenantID:    "tenant",
		PublicKeyFields: models.PublicKeyFields{
			Name:     "key",
			Username: ".*",
			Filter: models.PublicKeyFilter{
				Tags: []string{"tag1", "tag2", "tag3"},
			},
		},
	}

	_, err := db.Client().Database("test").Collection("public_keys").InsertOne(context.Background(), keyUsernameEmpty)
	assert.NoError(t, err)

	cases := []struct {
		description string
		Test        func(t *testing.T)
	}{
		{
			"Success to apply up on migration 46",
			func(t *testing.T) {
				t.Helper()

				migrations := GenerateMigrations()[45:46]
				migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
				err := migrates.Up(context.Background(), migrate.AllAvailable)
				assert.NoError(t, err)

				key := new(models.PublicKey)
				result := db.Client().Database("test").Collection("public_keys").FindOne(context.Background(), bson.M{"tenant_id": keyUsernameEmpty.TenantID})
				assert.NoError(t, result.Err())

				err = result.Decode(key)
				assert.NoError(t, err)

				sort.Strings(key.Filter.Tags)

				assert.Equal(t, keyUsernameRegexp, key)
			},
		},
		{
			"Success to apply down on migration 46",
			func(t *testing.T) {
				t.Helper()

				migrations := GenerateMigrations()[45:46]
				migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
				err := migrates.Down(context.Background(), migrate.AllAvailable)
				assert.NoError(t, err)

				key := new(models.PublicKey)
				result := db.Client().Database("test").Collection("public_keys").FindOne(context.Background(), bson.M{"tenant_id": keyUsernameRegexp.TenantID})
				assert.NoError(t, result.Err())

				err = result.Decode(key)
				assert.NoError(t, err)

				assert.Equal(t, keyUsernameEmpty, key)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, tc.Test)
	}
}
