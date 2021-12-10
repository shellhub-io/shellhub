package mongo

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestListAPIToken(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant", Tokens: []models.Token{}}

	_, err := db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	createdToken1, err := mongostore.TokenCreate(ctx, namespace.TenantID)
	assert.NoError(t, err)

	createdToken2, err := mongostore.TokenCreate(ctx, namespace.TenantID)
	assert.NoError(t, err)

	tokens, err := mongostore.TokenList(ctx, namespace.TenantID)
	assert.NoError(t, err)

	assert.Equal(t, *createdToken1, tokens[0])
	assert.Equal(t, *createdToken2, tokens[1])
}

func TestCreateAPIToken(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant", Tokens: []models.Token{}}

	_, err := db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	_, err = mongostore.TokenCreate(ctx, namespace.TenantID)
	assert.NoError(t, err)
}

func TestGetAPIToken(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant", Tokens: []models.Token{}}

	_, err := db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	createdToken, err := mongostore.TokenCreate(ctx, namespace.TenantID)
	assert.NoError(t, err)

	returnedToken, err := mongostore.TokenGet(ctx, namespace.TenantID, createdToken.ID)
	assert.NoError(t, err)
	assert.Equal(t, createdToken, returnedToken)
}

func TestDeleteAPIToken(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant", Tokens: []models.Token{}}

	_, err := db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	createdToken1, err := mongostore.TokenCreate(ctx, namespace.TenantID)
	assert.NoError(t, err)

	_, err = mongostore.TokenCreate(ctx, namespace.TenantID)
	assert.NoError(t, err)

	err = mongostore.TokenDelete(ctx, namespace.TenantID, createdToken1.ID)
	assert.NoError(t, err)

	tokens, err := mongostore.TokenList(ctx, namespace.TenantID)
	assert.NoError(t, err)
	assert.Equal(t, len(tokens), 1)
}

func TestUpdateAPIToken(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant", Tokens: []models.Token{}}

	_, err := db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	createdToken, err := mongostore.TokenCreate(ctx, namespace.TenantID)
	assert.NoError(t, err)

	err = mongostore.TokenUpdate(ctx, namespace.TenantID, createdToken.ID, false)
	assert.NoError(t, err)

	tokens, err := mongostore.TokenList(ctx, namespace.TenantID)
	assert.NoError(t, err)
	assert.Equal(t, tokens[0].ReadOnly, false)
}
