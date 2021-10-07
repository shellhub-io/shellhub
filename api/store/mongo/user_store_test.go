package mongo

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestUserGetByUsername(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(data.Context, &data.User)
	assert.NoError(t, err)

	u, err := mongostore.UserGetByUsername(data.Context, "username")
	assert.NoError(t, err)
	assert.NotEmpty(t, u)
	assert.Equal(t, u.ID, data.User.ID)
}

func TestUserGetByEmail(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(data.Context, &data.User)
	assert.NoError(t, err)

	u, err := mongostore.UserGetByEmail(data.Context, "user@shellhub.io")
	assert.NoError(t, err)
	assert.NotEmpty(t, u)
}

func TestUserUpdateData(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}

	result, err := db.Client().Database("test").Collection("users").InsertOne(data.Context, user)
	assert.NoError(t, err)

	objID := result.InsertedID.(primitive.ObjectID).Hex()

	userNewData := models.User{
		ID:       objID,
		Name:     "New Name",
		Username: "newusername",
		Password: "password",
		Email:    "new@email.com",
	}

	err = mongostore.UserUpdateData(data.Context, &userNewData, objID)
	assert.NoError(t, err)

	us, _, err := mongostore.UserGetByID(data.Context, objID, false)
	assert.Equal(t, us, &userNewData)
	assert.NoError(t, err)
}

func TestUserGetByID(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	namespacesOwner := []models.Namespace{
		{
			Name:     "namespace1",
			Owner:    "60af83d418d2dc3007cd445c",
			TenantID: "tenant1",
		},
		{
			Name:     "namespace2",
			Owner:    "60af83d418d2dc3007cd445c",
			TenantID: "tenant2",
		},
		{
			Name:     "namespace3",
			Owner:    "60af83d418d2dc3007cd445c",
			TenantID: "tenant3",
		},
		{
			Name:     "namespace4",
			Owner:    "60af83d418d2dc3007cd445c",
			TenantID: "tenant4",
		},
	}

	namespacesNotOwner := []models.Namespace{
		{
			Name:     "namespace18",
			Owner:    "60af83d418d2dc3007cd445d",
			TenantID: "tenant18",
		},
		{
			Name:     "namespace19",
			Owner:    "60af83d418d2dc3007cd445e",
			TenantID: "tenant19",
		},
		{
			Name:     "namespace20",
			Owner:    "60af83d418d2dc3007cd445f",
			TenantID: "tenant20",
		},
		{
			Name:     "namespace21",
			Owner:    "60af83d418d2dc3007cd4451",
			TenantID: "tenant21",
		},
		{
			Name:     "namespace22",
			Owner:    "60af83d418d2dc3007cd4452",
			TenantID: "tenant22",
		},
		{
			Name:     "namespace23",
			Owner:    "60af83d418d2dc3007cd4453",
			TenantID: "tenant23",
		},
		{
			Name:     "namespace24",
			Owner:    "60af83d418d2dc3007cd4454",
			TenantID: "tenant24",
		},
		{
			Name:     "namespace25",
			Owner:    "60af83d418d2dc3007cd4455",
			TenantID: "tenant25",
		},
		{
			Name:     "namespace26",
			Owner:    "060af83d418d2dc3007cd4456",
			TenantID: "tenant26",
		},
	}

	namespaces := namespacesOwner
	namespaces = append(namespaces, namespacesNotOwner...)
	nss := make([]interface{}, len(namespaces))

	for i, v := range namespaces {
		nss[i] = v
	}

	user := models.User{ID: "60af83d418d2dc3007cd445c", Name: "name", Username: "username", Password: "password", Email: "user@email.com"}

	objID, err := primitive.ObjectIDFromHex(user.ID)

	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("users").InsertOne(ctx, bson.M{
		"_id":      objID,
		"name":     user.Name,
		"username": user.Username,
		"password": user.Password,
		"email":    user.Email,
	})

	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertMany(ctx, nss)
	assert.NoError(t, err)

	us, countNs, err := mongostore.UserGetByID(ctx, user.ID, true)
	assert.NoError(t, err)
	assert.Equal(t, countNs, len(namespacesOwner))
	assert.Equal(t, us, &user)

	us, countNs, err = mongostore.UserGetByID(ctx, user.ID, false)
	assert.NoError(t, err)
	assert.Equal(t, countNs, 0)
	assert.Equal(t, us, &user)
}

func TestUserUpdatePassword(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}

	result, err := db.Client().Database("test").Collection("users").InsertOne(data.Context, user)
	assert.NoError(t, err)

	objID := result.InsertedID.(primitive.ObjectID).Hex()

	err = mongostore.UserUpdatePassword(data.Context, "password2", objID)
	assert.NoError(t, err)

	us, _, err := mongostore.UserGetByID(data.Context, objID, false)
	assert.Equal(t, us.Password, "password2")
	assert.NoError(t, err)
}

func TestUpdateUserFromAdmin(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}

	result, err := db.Client().Database("test").Collection("users").InsertOne(data.Context, user)
	assert.NoError(t, err)

	objID := result.InsertedID.(primitive.ObjectID).Hex()

	err = mongostore.UserUpdateFromAdmin(data.Context, "newName", "newUsername", "newEmail", "password", objID)
	assert.NoError(t, err)
}

func TestUserCreateToken(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}

	result, err := db.Client().Database("test").Collection("users").InsertOne(data.Context, user)
	assert.NoError(t, err)

	objID := result.InsertedID.(primitive.ObjectID).Hex()

	UserTokenRecover := models.UserTokenRecover{Token: "token", User: objID}

	err = mongostore.UserCreateToken(data.Context, &UserTokenRecover)
	assert.NoError(t, err)

	userToken, err := mongostore.UserGetToken(data.Context, objID)
	assert.Equal(t, userToken.Token, UserTokenRecover.Token)
	assert.NoError(t, err)
}

func TestUserUpdateAccountStatus(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}

	result, err := db.Client().Database("test").Collection("users").InsertOne(data.Context, user)
	assert.NoError(t, err)

	objID := result.InsertedID.(primitive.ObjectID).Hex()

	err = mongostore.UserUpdateAccountStatus(data.Context, objID)
	assert.NoError(t, err)

	us, _, err := mongostore.UserGetByID(data.Context, objID, false)
	assert.Equal(t, us.Authenticated, true)
	assert.NoError(t, err)
}

func TestUsersList(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(data.Context, &data.User)
	assert.NoError(t, err)

	users, count, err := mongostore.UserList(data.Context, paginator.Query{Page: -1, PerPage: -1}, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, users)
}

func TestUsersListWithFilter(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	result, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	namespace := models.Namespace{Name: "name", Owner: result.InsertedID.(primitive.ObjectID).Hex(), TenantID: "tenant"}
	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	user = models.User{Name: "name", Username: "username-1", Password: "password", Email: "email-1"}
	result, err = db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	namespace = models.Namespace{Name: "name", Owner: result.InsertedID.(primitive.ObjectID).Hex(), TenantID: "tenant"}
	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	user = models.User{Name: "name", Username: "username-2", Password: "password", Email: "email-2"}
	result, err = db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	namespace = models.Namespace{Name: "name", Owner: result.InsertedID.(primitive.ObjectID).Hex(), TenantID: "tenant"}
	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	namespace = models.Namespace{Name: "name", Owner: result.InsertedID.(primitive.ObjectID).Hex(), TenantID: "tenant"}
	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	user = models.User{Name: "name", Username: "username-3", Password: "password", Email: "email-3"}
	result, err = db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	namespace = models.Namespace{Name: "name", Owner: result.InsertedID.(primitive.ObjectID).Hex(), TenantID: "tenant"}
	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	namespace = models.Namespace{Name: "name", Owner: result.InsertedID.(primitive.ObjectID).Hex(), TenantID: "tenant"}
	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	filters := []models.Filter{
		{
			Type:   "property",
			Params: &models.PropertyParams{Name: "namespaces", Operator: "gt", Value: "1"},
		},
	}

	users, count, err := mongostore.UserList(ctx, paginator.Query{Page: -1, PerPage: -1}, filters)
	assert.NoError(t, err)
	assert.Equal(t, len(users), count)
	assert.Equal(t, 2, count)
	assert.NotEmpty(t, users)
}

func TestUserCreate(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(data.Context, &data.User)
	assert.NoError(t, err)
}

func TestUserDetachInfo(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	user := models.User{ID: "60af83d418d2dc3007cd445c", Name: "name", Username: "username", Password: "password", Email: "user@email.com"}

	objID, err := primitive.ObjectIDFromHex(user.ID)

	assert.NoError(t, err)

	_, _ = db.Client().Database("test").Collection("users").InsertOne(ctx, bson.M{
		"_id":      objID,
		"name":     user.Name,
		"username": user.Username,
		"password": user.Password,
		"email":    user.Email,
	})

	namespacesOwner := []*models.Namespace{
		{
			Owner:   user.ID,
			Name:    "ns2",
			Members: []interface{}{user.ID},
		},
		{
			Owner:   user.ID,
			Name:    "ns4",
			Members: []interface{}{user.ID},
		},
	}

	namespacesMember := []*models.Namespace{
		{
			Owner:   "id2",
			Name:    "ns1",
			Members: []interface{}{"id2", user.ID},
		},
		{
			Owner:   "id2",
			Name:    "ns3",
			Members: []interface{}{"id2", user.ID},
		},
		{
			Owner:   "id2",
			Name:    "ns5",
			Members: []interface{}{"id2", user.ID},
		},
	}

	namespaces := namespacesOwner
	namespaces = append(namespaces, namespacesMember...)
	nss := make([]interface{}, len(namespaces))

	for i, v := range namespaces {
		nss[i] = v
	}

	_, _ = db.Client().Database("test").Collection("namespaces").InsertMany(ctx, nss)

	u, err := mongostore.UserGetByUsername(ctx, "username")
	assert.NoError(t, err)
	assert.Equal(t, user.Username, u.Username)

	namespacesMap, err := mongostore.UserDetachInfo(ctx, user.ID)

	assert.NoError(t, err)
	assert.Equal(t, namespacesMap["member"], namespacesMember)
	assert.Equal(t, namespacesMap["owner"], namespacesOwner)
}

func TestUserDelete(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	user := models.User{ID: "60af83d418d2dc3007cd445c", Name: "name", Username: "username", Password: "password", Email: "user@email.com"}

	objID, err := primitive.ObjectIDFromHex(user.ID)

	assert.NoError(t, err)

	_, _ = db.Client().Database("test").Collection("users").InsertOne(ctx, bson.M{
		"_id":      objID,
		"name":     user.Name,
		"username": user.Username,
		"password": user.Password,
		"email":    user.Email,
	})

	err = mongostore.UserDelete(ctx, user.ID)
	assert.NoError(t, err)

	_, err = mongostore.UserGetByUsername(ctx, "username")
	assert.Error(t, err, mongo.ErrNoDocuments)
}
