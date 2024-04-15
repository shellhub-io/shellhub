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

func TestMigration44(t *testing.T) {
	logrus.Info("Testing Migration 44")

	db := dbtest.DB{}
	err := func() error {
		err := db.Down(context.Background())

		return err
	}()
	assert.NoError(t, err)

	keyTagDuplicated := &models.PublicKey{
		Fingerprint: "fingerprint",
		TenantID:    "tenant",
		PublicKeyFields: models.PublicKeyFields{
			Name:     "key",
			Username: ".*",
			Filter: models.PublicKeyFilter{
				Tags: []string{"tag1", "tag2", "tag2"},
			},
		},
	}

	keyTagWithoutDuplication := &models.PublicKey{
		Fingerprint: "fingerprint",
		TenantID:    "tenant",
		PublicKeyFields: models.PublicKeyFields{
			Name:     "key",
			Username: ".*",
			Filter: models.PublicKeyFilter{
				Tags: []string{"tag1", "tag2"},
			},
		},
	}

	keyTagNoDuplicated := &models.PublicKey{
		Fingerprint: "fingerprint1",
		TenantID:    "tenant1",
		PublicKeyFields: models.PublicKeyFields{
			Name:     "key1",
			Username: ".*",
			Filter: models.PublicKeyFilter{
				Tags: []string{"tag1", "tag2", "tag3"},
			},
		},
	}

	keyHostname := &models.PublicKey{
		Fingerprint: "fingerprint2",
		TenantID:    "tenant2",
		PublicKeyFields: models.PublicKeyFields{
			Name:     "key2",
			Username: ".*",
			Filter: models.PublicKeyFilter{
				Hostname: ".*",
			},
		},
	}

	_, err = mongoClient.Database("test").Collection("public_keys").InsertOne(context.TODO(), keyTagDuplicated)
	assert.NoError(t, err)
	_, err = mongoClient.Database("test").Collection("public_keys").InsertOne(context.TODO(), keyTagNoDuplicated)
	assert.NoError(t, err)
	_, err = mongoClient.Database("test").Collection("public_keys").InsertOne(context.TODO(), keyHostname)
	assert.NoError(t, err)

	cases := []struct {
		description string
		Test        func(t *testing.T)
	}{
		{
			"Success to apply up on migration 44 when public key tags are duplicated",
			func(t *testing.T) {
				t.Helper()

				migrations := GenerateMigrations()[43:44]
				migrates := migrate.NewMigrate(mongoClient.Database("test"), migrations...)
				err = migrates.Up(context.Background(), migrate.AllAvailable)
				assert.NoError(t, err)

				key := new(models.PublicKey)
				result := mongoClient.Database("test").Collection("public_keys").FindOne(context.TODO(), bson.M{"tenant_id": keyTagDuplicated.TenantID})
				assert.NoError(t, result.Err())

				err = result.Decode(key)
				assert.NoError(t, err)

				sort.Strings(key.Filter.Tags)

				assert.Equal(t, keyTagWithoutDuplication, key)
			},
		},
		{
			"Success to apply up on migration 44 when public key tags are not duplicated",
			func(t *testing.T) {
				t.Helper()

				migrations := GenerateMigrations()[43:44]
				migrates := migrate.NewMigrate(mongoClient.Database("test"), migrations...)
				err = migrates.Up(context.Background(), migrate.AllAvailable)
				assert.NoError(t, err)

				key := new(models.PublicKey)
				result := mongoClient.Database("test").Collection("public_keys").FindOne(context.TODO(), bson.M{"tenant_id": keyTagNoDuplicated.TenantID})
				assert.NoError(t, result.Err())

				err = result.Decode(key)
				assert.NoError(t, err)

				sort.Strings(key.Filter.Tags)

				assert.Equal(t, keyTagNoDuplicated, key)
			},
		},
		{
			"Success to apply up on migration 44 when public key has hostname",
			func(t *testing.T) {
				t.Helper()

				migrations := GenerateMigrations()[43:44]
				migrates := migrate.NewMigrate(mongoClient.Database("test"), migrations...)
				err = migrates.Up(context.Background(), migrate.AllAvailable)
				assert.NoError(t, err)

				key := new(models.PublicKey)
				result := mongoClient.Database("test").Collection("public_keys").FindOne(context.TODO(), bson.M{"tenant_id": keyHostname.TenantID})
				assert.NoError(t, result.Err())

				err = result.Decode(key)
				assert.NoError(t, err)

				assert.Equal(t, keyHostname, key)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, tc.Test)
	}
}
