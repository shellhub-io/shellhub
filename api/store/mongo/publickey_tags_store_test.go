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
				t.Helper()

				err := store.PublicKeyAddTag(ctx, "invalidTenant", "fingerprintKeyTags", "tag")
				assert.Error(t, err)
			},
		},
		{
			description: "success to add tag to public key",
			test: func(t *testing.T) {
				t.Helper()

				err := store.PublicKeyAddTag(ctx, "tenant", "fingerprintKeyTags", "tag")
				assert.NoError(t, err)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, tc.test)
	}
}

func TestPublicKeyRemoveTag(t *testing.T) {
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
			description: "fail to remove a tag from a public key when tenant is invalid",
			test: func(t *testing.T) {
				t.Helper()

				err := store.PublicKeyRemoveTag(ctx, "invalidTenant", "fingerprintKeyTags", "tag1")
				assert.Error(t, err)
			},
		},
		{
			description: "fail to remove a tag from a public key when tag does not exist",
			test: func(t *testing.T) {
				t.Helper()

				err := store.PublicKeyRemoveTag(ctx, "tenant", "fingerprintKeyTags", "tag3")
				assert.Error(t, err)
			},
		},
		{
			description: "success to remove a tag from a public key",
			test: func(t *testing.T) {
				t.Helper()

				err := store.PublicKeyRemoveTag(ctx, "tenant", "fingerprintKeyTags", "tag1")
				assert.NoError(t, err)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, tc.test)
	}
}

func TestPublicKeyUpdateTags(t *testing.T) {
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
			description: "fail to update tags to public key when tenant is not valid",
			test: func(t *testing.T) {
				t.Helper()

				err := store.PublicKeyUpdateTags(ctx, "invalidTenant", "fingerprintKeyTags", []string{"tag1", "tag2", "tag3"})
				assert.Error(t, err)
			},
		},
		{
			description: "success to update tags to public key",
			test: func(t *testing.T) {
				t.Helper()

				err := store.PublicKeyUpdateTags(ctx, "tenant", "fingerprintKeyTags", []string{"tag1", "tag2", "tag3"})
				assert.NoError(t, err)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, tc.test)
	}
}

func TestPublicKeyRenameTag(t *testing.T) {
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
			description: "fail to rename a tags from public key when tenant is not valid",
			test: func(t *testing.T) {
				t.Helper()

				err := store.PublicKeyRenameTag(ctx, "invalidTenant", "tag2", "tag4")
				assert.Error(t, err)
			},
		},
		{
			description: "fail to rename a tags from public key when tag does not exist",
			test: func(t *testing.T) {
				t.Helper()

				err := store.PublicKeyRenameTag(ctx, "tenant", "tag4", "tag5")
				assert.Error(t, err)
			},
		},
		{
			description: "success to rename a tags from public key",
			test: func(t *testing.T) {
				t.Helper()

				err := store.PublicKeyRenameTag(ctx, "tenant", "tag2", "tag4")
				assert.NoError(t, err)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, tc.test)
	}
}

func TestPublicKeyDeleteTag(t *testing.T) {
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
			description: "fail to delete a tags from all public keys",
			test: func(t *testing.T) {
				t.Helper()

				err := store.PublicKeyDeleteTag(ctx, "invalidTenant", "tag2")
				assert.Error(t, err)
			},
		},
		{
			description: "fail to delete a tag from public key when tag does not exist",
			test: func(t *testing.T) {
				t.Helper()

				err := store.PublicKeyDeleteTag(ctx, "tenant", "tag4")
				assert.Error(t, err)
			},
		},
		{
			description: "success to delete a tags from all public keys",
			test: func(t *testing.T) {
				t.Helper()

				err := store.PublicKeyDeleteTag(ctx, "tenant", "tag2")
				assert.NoError(t, err)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, tc.test)
	}
}

func TestPublicKeyGetTags(t *testing.T) {
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
			description: "fail to get all tags from all public keys when tenant is not valid",
			test: func(t *testing.T) {
				t.Helper()

				tags, lines, _ := store.PublicKeyGetTags(ctx, "invalidTenant")
				assert.Equal(t, []string{}, tags)
				assert.Equal(t, 0, lines)
			},
		},
		{
			description: "success to get all tags from all public keys",
			test: func(t *testing.T) {
				t.Helper()

				tags, lines, err := store.PublicKeyGetTags(ctx, "tenant")
				assert.NoError(t, err)
				assert.Equal(t, []string{"tag1", "tag2"}, tags)
				assert.Equal(t, 2, lines)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, tc.test)
	}
}
