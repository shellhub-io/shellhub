package mongo

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
)

func setup(ctx context.Context, db *mongo.Database) {
	save := func(ctx context.Context, db *mongo.Database, collection string, data interface{}) {
		_, err := db.Collection(collection).InsertOne(ctx, data)
		if err != nil {
			panic(err)
		}
	}

	tenant := "tenant"

	namespace := models.Namespace{
		Name:       "namespace's name",
		TenantID:   tenant,
		Devices:    0,
		MaxDevices: 3,
	}

	device := models.Device{
		Name:     "device's name",
		TenantID: tenant,
		UID:      hex.EncodeToString([]byte(tenant)),
	}

	keyHostname := models.PublicKey{
		Fingerprint:     "fingerprintKeyHostname",
		TenantID:        tenant,
		PublicKeyFields: models.PublicKeyFields{Filter: models.PublicKeyFilter{Hostname: ".*"}},
	}
	keyTags := models.PublicKey{
		TenantID:        tenant,
		Fingerprint:     "fingerprintKeyTags",
		PublicKeyFields: models.PublicKeyFields{Filter: models.PublicKeyFilter{Tags: []string{"tag1", "tag2"}}},
	}

	save(ctx, db, "namespaces", namespace)
	save(ctx, db, "devices", device)
	save(ctx, db, "public_keys", keyHostname)
	save(ctx, db, "public_keys", keyTags)
}

func TestPublicKeyAddTag(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	setup(ctx, db.Client().Database("test"))

	store := NewStore(db.Client().Database("test"), cache.NewNullCache())

	cases := []struct {
		description string
		test        func(t *testing.T)
	}{
		{
			description: "fail to add tag to public key",
			test: func(t *testing.T) {
				err := store.PublicKeyAddTag(ctx, "invalidTenant", "fingerprintKeyTags", "tag")
				assert.Error(t, err)
			},
		},
		{
			description: "success to add tag to public key",
			test: func(t *testing.T) {
				err := store.PublicKeyAddTag(ctx, "tenant", "fingerprintKeyTags", "tag")
				assert.NoError(t, err)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, tc.test)
	}
}
