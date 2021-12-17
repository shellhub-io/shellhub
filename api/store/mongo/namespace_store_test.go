package mongo

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

const user2Username = "username2"

func TestNamespaceGetDataUserSecurity(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(data.Context, &data.User)
	assert.NoError(t, err)

	_, err = mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	returnedStatus, err := mongostore.NamespaceGetSessionRecord(data.Context, data.Namespace.TenantID)
	assert.Equal(t, returnedStatus, data.Namespace.Settings.SessionRecord)
	assert.NoError(t, err)
}

func TestNamespaceUpdateDataUserSecurity(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(data.Context, &data.User)
	assert.NoError(t, err)

	_, err = mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.NamespaceSetSessionRecord(data.Context, false, data.Namespace.TenantID)
	assert.NoError(t, err)

	returnedStatus, err := mongostore.NamespaceGetSessionRecord(data.Context, data.Namespace.TenantID)
	assert.Equal(t, returnedStatus, false)
	assert.NoError(t, err)
}

func TestNamespaceCreate(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()

	userValid := &models.User{}
	namespaceValid := &models.Namespace{}

	store := NewStore(db.Client().Database("test"), cache.NewNullCache())
	tests := []struct {
		description string
		namespace   *models.Namespace
		test        func(t *testing.T)
	}{
		{
			description: "Fail when user is nil",
			namespace:   nil,
			test: func(t *testing.T) {
				err := store.UserCreate(ctx, nil)
				assert.Error(t, err)
			},
		},
		{
			description: "Success to create a namespace",
			namespace:   namespaceValid,
			test: func(t *testing.T) {
				err := store.UserCreate(ctx, userValid)
				assert.NoError(t, err)

				_, err = store.NamespaceCreate(ctx, namespaceValid)
				assert.NoError(t, err)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, test.test)
	}
}

func TestNamespaceDelete(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()

	userValid := &models.User{
		ID: "userValidID",
		UserData: models.UserData{
			Name:     "userValidName",
			Email:    "userValidName@email.com",
			Username: "userValidUsername",
		},
		UserPassword: models.UserPassword{},
	}
	userInvalid := &models.User{
		ID: "userInvalidID",
		UserData: models.UserData{
			Name:     "userInvalidName",
			Email:    "userInvalidName@email.com",
			Username: "userInvalidUsername",
		},
		UserPassword: models.UserPassword{},
	}
	namespaceValid := &models.Namespace{
		Name:     "namespaceValidName",
		Owner:    userValid.ID,
		TenantID: "namespaceValidTenant",
	}
	namespaceInvalid := &models.Namespace{
		Name:     "namespaceInvalidName",
		Owner:    userInvalid.ID,
		TenantID: "namespaceInvalidTenant",
	}

	store := NewStore(db.Client().Database("test"), cache.NewNullCache())
	tests := []struct {
		description string
		test        func(t *testing.T)
	}{
		{
			description: "Fail when cannot find the tenant ID",
			test: func(t *testing.T) {
				err := store.NamespaceDelete(ctx, namespaceInvalid.TenantID)
				assert.Error(t, err)
			},
		},
		{
			description: "Success to delete a namespace",
			test: func(t *testing.T) {
				err := store.UserCreate(ctx, userValid)
				assert.NoError(t, err)

				_, err = store.NamespaceCreate(ctx, namespaceValid)
				assert.NoError(t, err)

				err = store.NamespaceDelete(ctx, namespaceValid.TenantID)
				assert.NoError(t, err)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, test.test)
	}
}

func TestNamespaceGet(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(data.Context, &data.User)
	assert.NoError(t, err)

	_, err = mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	_, err = mongostore.NamespaceGet(data.Context, "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx")
	assert.NoError(t, err)
}

func TestNamespacesList(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(data.Context, &data.User)
	assert.NoError(t, err)

	_, err = mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	var nsList []models.Namespace
	nsList = append(nsList, data.Namespace)

	returnedNsList, count, err := mongostore.NamespaceList(data.Context, paginator.Query{Page: -1, PerPage: -1}, nil, false)
	assert.Equal(t, nsList, returnedNsList)
	assert.Equal(t, 1, count)
	assert.NoError(t, err)
}

func TestNamespaceAddMember(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(data.Context, &data.User)
	assert.NoError(t, err)

	user2 := data.User
	user2.Username = user2Username
	user2.ID = "user2_id"

	err = mongostore.UserCreate(data.Context, &user2)
	assert.NoError(t, err)

	_, err = mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	u, err := mongostore.UserGetByUsername(data.Context, "username")
	assert.NoError(t, err)

	_, err = mongostore.NamespaceAddMember(data.Context, "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", u.ID, authorizer.MemberRoleObserver)
	assert.NoError(t, err)
}

func TestNamespaceUpdate(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(data.Context, &data.User)
	assert.NoError(t, err)

	_, err = mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.NamespaceUpdate(data.Context, "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", &models.Namespace{
		Name:       "name",
		Settings:   &models.NamespaceSettings{SessionRecord: false},
		MaxDevices: 3,
	})
	assert.NoError(t, err)
}

func TestNamespaceRemoveMember(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(data.Context, &data.User)
	assert.NoError(t, err)

	user2 := data.User
	user2.Username = user2Username
	user2.ID = "user2_id"

	err = mongostore.UserCreate(data.Context, &user2)
	assert.NoError(t, err)

	_, err = mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	u, err := mongostore.UserGetByUsername(data.Context, "username")
	assert.NoError(t, err)

	_, err = mongostore.NamespaceAddMember(data.Context, "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", u.ID, "")
	assert.NoError(t, err)

	_, err = mongostore.NamespaceRemoveMember(data.Context, "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", u.ID)
	assert.NoError(t, err)
}

func TestNamespaceGetByName(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(data.Context, &data.User)
	assert.NoError(t, err)

	ns, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	returnedNs, err := mongostore.NamespaceGetByName(data.Context, "namespace")
	assert.NoError(t, err)
	assert.Equal(t, ns, returnedNs)
}
