package mongo

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestPublicKeyCreate(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	newKey := &models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint", TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste1", Hostname: ".*"},
	}
	err := mongostore.PublicKeyCreate(ctx, newKey)
	assert.NoError(t, err)
}

func TestPublicKeysList(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}
	key := models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint", CreatedAt: clock.Now(), TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste", Hostname: ".*"},
	}
	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("public_keys").InsertOne(ctx, key)
	assert.NoError(t, err)

	_, count, err := mongostore.PublicKeyList(ctx, paginator.Query{Page: -1, PerPage: -1})
	assert.Equal(t, 1, count)
	assert.NoError(t, err)
}

func TestPublicKeyGet(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}
	key := models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint", CreatedAt: clock.Now(), TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste", Hostname: ".*"},
	}
	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("public_keys").InsertOne(ctx, key)
	assert.NoError(t, err)

	k, err := mongostore.PublicKeyGet(ctx, key.Fingerprint, key.TenantID)
	assert.NoError(t, err)
	assert.NotEmpty(t, k)
}

func TestPublicKeyUpdate(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}
	// createdAt := time.Now()
	key := &models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint", TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste", Hostname: ".*"},
	}
	updatedKey := &models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint", TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste2", Hostname: ".*"},
	}
	unexistingKey := &models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint2", TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste", Hostname: ".*"},
	}

	update := &models.PublicKeyUpdate{
		PublicKeyFields: models.PublicKeyFields{Name: "teste2", Hostname: ".*"},
	}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("public_keys").InsertOne(ctx, key)
	assert.NoError(t, err)

	k, err := mongostore.PublicKeyUpdate(ctx, key.Fingerprint, key.TenantID, update)
	assert.NoError(t, err)
	assert.Equal(t, k, updatedKey)
	_, err = mongostore.PublicKeyUpdate(ctx, unexistingKey.Fingerprint, unexistingKey.TenantID, update)
	assert.EqualError(t, err, store.ErrNoDocuments.Error())
}

func TestPublicKeyDelete(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}
	newKey := &models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint", TenantID: "tenant", PublicKeyFields: models.PublicKeyFields{Name: "teste1", Hostname: ".*"},
	}

	_, err := db.Client().Database("test").Collection("public_keys").InsertOne(ctx, newKey)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	err = mongostore.PublicKeyDelete(ctx, newKey.Fingerprint, newKey.TenantID)
	assert.NoError(t, err)
}
