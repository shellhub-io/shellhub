package mongo

import (
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

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
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(data.Context, &data.User)
	assert.NoError(t, err)

	_, err = mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)
}

func TestNamespaceDelete(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(data.Context, &data.User)
	assert.NoError(t, err)

	_, err = mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.NamespaceDelete(data.Context, "00000000-0000-4000-0000-000000000000")
	assert.NoError(t, err)
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

	_, err = mongostore.NamespaceGet(data.Context, "00000000-0000-4000-0000-000000000000")
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

	member := data.User
	member.ID = "507f1f77bcf86cd799439012"
	member.Username = "memberFromNamespace"

	err = mongostore.UserCreate(data.Context, &member)
	assert.NoError(t, err)

	_, err = mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	u, err := mongostore.UserGetByUsername(data.Context, member.Username)
	assert.NoError(t, err)

	_, err = mongostore.NamespaceAddMember(data.Context, "00000000-0000-4000-0000-000000000000", u.ID, guard.RoleObserver)
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

	err = mongostore.NamespaceUpdate(data.Context, "00000000-0000-4000-0000-000000000000", &models.Namespace{
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

	member := data.User
	member.ID = "507f1f77bcf86cd799439012"
	member.Username = "memberFromNamespace"

	err = mongostore.UserCreate(data.Context, &member)
	assert.NoError(t, err)

	_, err = mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	u, err := mongostore.UserGetByUsername(data.Context, member.Username)
	assert.NoError(t, err)

	_, err = mongostore.NamespaceAddMember(data.Context, "00000000-0000-4000-0000-000000000000", u.ID, guard.RoleObserver)
	assert.NoError(t, err)

	_, err = mongostore.NamespaceRemoveMember(data.Context, "00000000-0000-4000-0000-000000000000", u.ID)
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
