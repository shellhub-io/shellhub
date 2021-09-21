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
)

func TestUserGetByUsername(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email", ID: "owner"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	u, err := mongostore.UserGetByUsername(ctx, "username")
	assert.NoError(t, err)
	assert.NotEmpty(t, u)
	assert.Equal(t, u.ID, user.ID)
}

func TestUserGetByEmail(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	u, err := mongostore.UserGetByEmail(ctx, "email")
	assert.NoError(t, err)
	assert.NotEmpty(t, u)
}

func TestUserUpdateData(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	result, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	objID := result.InsertedID.(primitive.ObjectID).Hex()

	userNewData := models.User{
		ID:       objID,
		Name:     "New Name",
		Username: "newusername",
		Password: "password",
		Email:    "new@email.com",
	}

	err = mongostore.UserUpdateData(ctx, &userNewData, objID)
	assert.NoError(t, err)

	us, _, err := mongostore.UserGetByID(ctx, objID, false)
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

	namespaces := append(namespacesOwner, namespacesNotOwner...)
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
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	result, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	objID := result.InsertedID.(primitive.ObjectID).Hex()

	newPassword := "password2"

	err = mongostore.UserUpdatePassword(ctx, newPassword, objID)
	assert.NoError(t, err)

	us, _, err := mongostore.UserGetByID(ctx, objID, false)
	assert.Equal(t, us.Password, newPassword)
	assert.NoError(t, err)
}

func TestUpdateUserFromAdmin(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	result, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	objID := result.InsertedID.(primitive.ObjectID).Hex()
	err = mongostore.UserUpdateFromAdmin(ctx, "newName", "newUsername", "newEmail", "password", objID)
	assert.NoError(t, err)
}

func TestUserCreateToken(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email", Authenticated: false}

	result, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	objID := result.InsertedID.(primitive.ObjectID).Hex()

	UserTokenRecover := models.UserTokenRecover{Token: "token", User: objID}

	err = mongostore.UserCreateToken(ctx, &UserTokenRecover)
	assert.NoError(t, err)

	userToken, err := mongostore.UserGetToken(ctx, objID)
	assert.Equal(t, userToken.Token, UserTokenRecover.Token)
	assert.NoError(t, err)
}

func TestUserUpdateAccountStatus(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email", Authenticated: false}

	result, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	objID := result.InsertedID.(primitive.ObjectID).Hex()

	err = mongostore.UserUpdateAccountStatus(ctx, objID)
	assert.NoError(t, err)

	us, _, err := mongostore.UserGetByID(ctx, objID, false)
	assert.Equal(t, us.Authenticated, true)
	assert.NoError(t, err)
}

func TestUsersList(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	result, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	userID := result.InsertedID.(primitive.ObjectID).Hex()
	namespace := models.Namespace{Name: "name", Owner: userID, TenantID: "tenant"}
	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	users, count, err := mongostore.UserList(ctx, paginator.Query{Page: -1, PerPage: -1}, nil)
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
}
