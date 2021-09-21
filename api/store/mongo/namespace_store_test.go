package mongo

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestNamespaceGetDataUserSecurity(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email", ID: "hash1"}
	namespace := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Settings: &models.NamespaceSettings{SessionRecord: true}}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	returnedStatus, err := mongostore.NamespaceGetSessionRecord(ctx, namespace.TenantID)
	assert.Equal(t, returnedStatus, namespace.Settings.SessionRecord)
	assert.NoError(t, err)
}

func TestNamespaceUpdateDataUserSecurity(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email", ID: "hash1"}
	namespace := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Settings: &models.NamespaceSettings{SessionRecord: true}}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	err = mongostore.NamespaceSetSessionRecord(ctx, false, namespace.TenantID)
	assert.NoError(t, err)
}

func TestNamespaceCreate(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(ctx, &models.User{
		Name:     "user",
		Email:    "user@shellhub.io",
		Password: "password",
	})
	assert.NoError(t, err)
	_, err = mongostore.NamespaceCreate(ctx, &models.Namespace{
		Name:       "namespace",
		Owner:      "owner",
		TenantID:   "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Members:    []interface{}{"owner"},
		MaxDevices: -1,
	})
	assert.NoError(t, err)
}

func TestNamespaceDelete(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(ctx, &models.User{
		Name:     "user",
		Email:    "user@shellhub.io",
		Password: "password",
	})
	assert.NoError(t, err)
	_, err = mongostore.NamespaceCreate(ctx, &models.Namespace{
		Name:       "namespace",
		Owner:      "owner",
		TenantID:   "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Members:    []interface{}{"owner"},
		MaxDevices: -1,
	})
	assert.NoError(t, err)

	err = mongostore.NamespaceDelete(ctx, "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx")
	assert.NoError(t, err)
}

func TestNamespaceGet(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(ctx, &models.User{
		Name:     "user",
		Email:    "user@shellhub.io",
		Password: "password",
	})
	assert.NoError(t, err)
	_, err = mongostore.NamespaceCreate(ctx, &models.Namespace{
		Name:       "namespace",
		Owner:      "owner",
		TenantID:   "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Members:    []interface{}{"owner"},
		MaxDevices: -1,
	})
	assert.NoError(t, err)

	_, err = mongostore.NamespaceGet(ctx, "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx")
	assert.NoError(t, err)
}

func TestNamespacesList(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(ctx, &models.User{
		Username: "user",
		Email:    "user@shellhub.io",
		Password: "password",
	})
	assert.NoError(t, err)
	_, err = mongostore.NamespaceCreate(ctx, &models.Namespace{
		Name:       "namespace",
		Owner:      "owner",
		TenantID:   "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Members:    []interface{}{"owner"},
		MaxDevices: -1,
	})
	assert.NoError(t, err)

	_, count, err := mongostore.NamespaceList(ctx, paginator.Query{Page: -1, PerPage: -1}, nil, false)
	assert.Equal(t, 1, count)
	assert.NoError(t, err)
}

func TestNamespaceAddMember(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(ctx, &models.User{
		Username: "user",
		Email:    "user@shellhub.io",
		Password: "password",
		ID:       "user_id",
	})
	assert.NoError(t, err)
	err = mongostore.UserCreate(ctx, &models.User{
		Username: "user2",
		Email:    "user@shellhub.io",
		Password: "password",
		ID:       "user2_id",
	})
	assert.NoError(t, err)
	_, err = mongostore.NamespaceCreate(ctx, &models.Namespace{
		Name:       "namespace",
		Owner:      "owner",
		TenantID:   "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Members:    []interface{}{"owner"},
		MaxDevices: -1,
	})
	assert.NoError(t, err)

	u, err := mongostore.UserGetByUsername(ctx, "user")
	assert.NoError(t, err)

	_, err = mongostore.NamespaceAddMember(ctx, "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", u.ID)
	assert.NoError(t, err)
}

func TestNamespaceUpdate(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(ctx, &models.User{
		Name:     "name",
		Username: "user",
		Email:    "user@shellhub.io",
		Password: "password",
	})
	assert.NoError(t, err)

	_, err = mongostore.NamespaceCreate(ctx, &models.Namespace{
		Name:       "namespace",
		Owner:      "owner",
		TenantID:   "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Members:    []interface{}{"owner"},
		Settings:   &models.NamespaceSettings{SessionRecord: true},
		MaxDevices: -1,
	})
	assert.NoError(t, err)

	err = mongostore.NamespaceUpdate(ctx, "tenant", &models.Namespace{
		Name:       "name",
		Settings:   &models.NamespaceSettings{SessionRecord: false},
		MaxDevices: 3,
	})
	assert.NoError(t, err)
}

func TestNamespaceRemoveMember(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(ctx, &models.User{
		Username: "user",
		Email:    "user@shellhub.io",
		Password: "password",
	})
	assert.NoError(t, err)
	err = mongostore.UserCreate(ctx, &models.User{
		Username: "user2",
		Email:    "user@shellhub.io",
		Password: "password",
	})
	assert.NoError(t, err)
	_, err = mongostore.NamespaceCreate(ctx, &models.Namespace{
		Name:       "namespace",
		Owner:      "owner",
		TenantID:   "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Members:    []interface{}{"owner"},
		MaxDevices: -1,
	})
	assert.NoError(t, err)

	u, err := mongostore.UserGetByUsername(ctx, "user")
	assert.NoError(t, err)

	_, err = mongostore.NamespaceAddMember(ctx, "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", u.ID)
	assert.NoError(t, err)

	_, err = mongostore.NamespaceRemoveMember(ctx, "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", u.ID)
	assert.NoError(t, err)
}

func TestNamespaceGetByName(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(ctx, &models.User{
		Name:     "user",
		Email:    "user@shellhub.io",
		Password: "password",
	})
	assert.NoError(t, err)
	ns, err := mongostore.NamespaceCreate(ctx, &models.Namespace{
		Name:       "namespace",
		Owner:      "owner",
		TenantID:   "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Members:    []interface{}{"owner"},
		MaxDevices: -1,
	})
	assert.NoError(t, err)

	returnedNs, err := mongostore.NamespaceGetByName(ctx, "namespace")
	assert.NoError(t, err)
	assert.Equal(t, ns, returnedNs)
}
