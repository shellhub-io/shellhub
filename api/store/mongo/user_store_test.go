package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/mongotest"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestUserList(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		users []models.User
		count int
		err   error
	}

	cases := []struct {
		description string
		setup       func() error
		expected    Expected
	}{
		{
			description: "succeeds when users are found",
			setup: func() error {
				return mongotest.UseFixture(fixtures.User)
			},
			expected: Expected{
				users: []models.User{
					{
						ID:             "507f1f77bcf86cd799439011",
						CreatedAt:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						LastLogin:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						EmailMarketing: false,
						Confirmed:      false,
						UserData: models.UserData{
							Name:     "john doe",
							Username: "john_doe",
							Email:    "user@test.com",
						},
						MaxNamespaces: 0,
						UserPassword: models.UserPassword{
							HashedPassword: "fcf730b6d95236ecd3c9fc2d92d7b6b2bb061514961aec041d6c7a7192f592e4",
						},
					},
				},
				count: 1,
				err:   nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			users, count, err := mongostore.UserList(ctx, paginator.Query{Page: -1, PerPage: -1}, nil)
			assert.Equal(t, tc.expected, Expected{users: users, count: count, err: err})

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestUserListWithFilter(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		users []models.User
		count int
		err   error
	}

	cases := []struct {
		description string
		setup       func() error
		expected    Expected
	}{
		{
			description: "succeeds when no users are found",
			setup: func() error {
				return mongotest.UseFixture(fixtures.User)
			},
			expected: Expected{
				users: []models.User{},
				count: 0,
				err:   nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			filters := []models.Filter{
				{
					Type:   "property",
					Params: &models.PropertyParams{Name: "namespaces", Operator: "gt", Value: "1"},
				},
			}
			users, count, err := mongostore.UserList(ctx, paginator.Query{Page: -1, PerPage: -1}, filters)
			assert.Equal(t, tc.expected, Expected{users: users, count: count, err: err})

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestUserCreate(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	cases := []struct {
		description string
		user        *models.User
		setup       func() error
		expected    error
	}{
		{
			description: "succeeds when data is valid",
			user: &models.User{
				ID: "507f1f77bcf86cd799439011",
				UserData: models.UserData{
					Name:     "john doe",
					Username: "john_doe",
					Email:    "user@test.com",
				},
				UserPassword: models.NewUserPassword("secret123"),
			},
			setup: func() error {
				return nil
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.UserCreate(ctx, tc.user)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestUserGetByUsername(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		user *models.User
		err  error
	}

	cases := []struct {
		description string
		username    string
		setup       func() error
		expected    Expected
	}{
		{
			description: "fails when user is not found",
			username:    "nonexistent",
			setup: func() error {
				return mongotest.UseFixture(fixtures.User)
			},
			expected: Expected{
				user: nil,
				err:  store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when user is found",
			username:    "john_doe",
			setup: func() error {
				return mongotest.UseFixture(fixtures.User)
			},
			expected: Expected{
				user: &models.User{
					ID:             "507f1f77bcf86cd799439011",
					CreatedAt:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					LastLogin:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					EmailMarketing: false,
					Confirmed:      false,
					UserData: models.UserData{
						Name:     "john doe",
						Username: "john_doe",
						Email:    "user@test.com",
					},
					MaxNamespaces: 0,
					UserPassword: models.UserPassword{
						HashedPassword: "fcf730b6d95236ecd3c9fc2d92d7b6b2bb061514961aec041d6c7a7192f592e4",
					},
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			user, err := mongostore.UserGetByUsername(ctx, tc.username)
			assert.Equal(t, tc.expected, Expected{user: user, err: err})

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestUserGetByEmail(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		user *models.User
		err  error
	}

	cases := []struct {
		description string
		email       string
		setup       func() error
		expected    Expected
	}{
		{
			description: "fails when email is not found",
			email:       "nonexistent",
			setup: func() error {
				return mongotest.UseFixture(fixtures.User)
			},
			expected: Expected{
				user: nil,
				err:  store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when email is found",
			email:       "user@test.com",
			setup: func() error {
				return mongotest.UseFixture(fixtures.User)
			},
			expected: Expected{
				user: &models.User{
					ID:             "507f1f77bcf86cd799439011",
					CreatedAt:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					LastLogin:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					EmailMarketing: false,
					Confirmed:      false,
					UserData: models.UserData{
						Name:     "john doe",
						Username: "john_doe",
						Email:    "user@test.com",
					},
					MaxNamespaces: 0,
					UserPassword: models.UserPassword{
						HashedPassword: "fcf730b6d95236ecd3c9fc2d92d7b6b2bb061514961aec041d6c7a7192f592e4",
					},
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			user, err := mongostore.UserGetByEmail(ctx, tc.email)
			assert.Equal(t, tc.expected, Expected{user: user, err: err})

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestUserGetByID(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		user *models.User
		ns   int
		err  error
	}

	cases := []struct {
		description string
		id          string
		ns          bool
		setup       func() error
		expected    Expected
	}{
		{
			description: "fails when user is not found",
			id:          "507f1f77bcf86cd7994390bb",
			setup: func() error {
				return mongotest.UseFixture(fixtures.User)
			},
			expected: Expected{
				user: nil,
				ns:   0,
				err:  store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when user is found with ns equal false",
			id:          "507f1f77bcf86cd799439011",
			ns:          false,
			setup: func() error {
				return mongotest.UseFixture(fixtures.Namespace, fixtures.User)
			},
			expected: Expected{
				user: &models.User{
					ID:             "507f1f77bcf86cd799439011",
					CreatedAt:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					LastLogin:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					EmailMarketing: false,
					Confirmed:      false,
					UserData: models.UserData{
						Name:     "john doe",
						Username: "john_doe",
						Email:    "user@test.com",
					},
					MaxNamespaces: 0,
					UserPassword: models.UserPassword{
						HashedPassword: "fcf730b6d95236ecd3c9fc2d92d7b6b2bb061514961aec041d6c7a7192f592e4",
					},
				},
				ns:  0,
				err: nil,
			},
		},
		{
			description: "succeeds when user is found with ns equal true",
			id:          "507f1f77bcf86cd799439011",
			ns:          true,
			setup: func() error {
				return mongotest.UseFixture(fixtures.Namespace, fixtures.User)
			},
			expected: Expected{
				user: &models.User{
					ID:             "507f1f77bcf86cd799439011",
					CreatedAt:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					LastLogin:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					EmailMarketing: false,
					Confirmed:      false,
					UserData: models.UserData{
						Name:     "john doe",
						Username: "john_doe",
						Email:    "user@test.com",
					},
					MaxNamespaces: 0,
					UserPassword: models.UserPassword{
						HashedPassword: "fcf730b6d95236ecd3c9fc2d92d7b6b2bb061514961aec041d6c7a7192f592e4",
					},
				},
				ns:  1,
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, tc.setup())

			user, ns, err := mongostore.UserGetByID(ctx, tc.id, tc.ns)
			assert.Equal(t, tc.expected, Expected{user: user, ns: ns, err: err})

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestUserUpdateData(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		id          string
		data        models.User
		setup       func() error
		expected    error
	}{
		{
			description: "fails when user is not found",
			id:          "000000000000000000000000",
			data: models.User{
				LastLogin: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				UserData: models.UserData{
					Name:     "edited name",
					Username: "edited_name",
					Email:    "edited@test.com",
				},
			},
			setup: func() error {
				return mongotest.UseFixture(fixtures.User)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when user is found",
			id:          "507f1f77bcf86cd799439011",
			setup: func() error {
				return mongotest.UseFixture(fixtures.User)
			},
			data: models.User{
				LastLogin: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				UserData: models.UserData{
					Name:     "edited name",
					Username: "edited_name",
					Email:    "edited@test.com",
				},
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.UserUpdateData(ctx, tc.id, tc.data)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestUserUpdatePassword(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		id          string
		password    string
		setup       func() error
		expected    error
	}{
		{
			description: "fails when user is not found",
			id:          "000000000000000000000000",
			password:    "other_password",
			setup: func() error {
				return mongotest.UseFixture(fixtures.User)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when user is found",
			id:          "507f1f77bcf86cd799439011",
			password:    "other_password",
			setup: func() error {
				return mongotest.UseFixture(fixtures.User)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.UserUpdatePassword(ctx, tc.password, tc.id)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestUserUpdateAccountStatus(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		id          string
		setup       func() error
		expected    error
	}{
		{
			description: "fails when user is not found",
			id:          "000000000000000000000000",
			setup: func() error {
				return mongotest.UseFixture(fixtures.User)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when user is found",
			id:          "507f1f77bcf86cd799439011",
			setup: func() error {
				return mongotest.UseFixture(fixtures.User)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.UserUpdateAccountStatus(ctx, tc.id)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestUserUpdateFromAdmin(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		id          string
		name        string
		username    string
		email       string
		password    string
		setup       func() error
		expected    error
	}{
		{
			description: "fails when user is not found",
			id:          "000000000000000000000000",
			name:        "other name",
			username:    "other_name",
			email:       "other.email@test.com",
			password:    "other_password",
			setup: func() error {
				return mongotest.UseFixture(fixtures.User)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when user is found",
			id:          "507f1f77bcf86cd799439011",
			name:        "other name",
			username:    "other_name",
			email:       "other.email@test.com",
			password:    "other_password",
			setup: func() error {
				return mongotest.UseFixture(fixtures.User)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.UserUpdateFromAdmin(ctx, tc.name, tc.username, tc.email, tc.password, tc.id)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestUserCreateToken(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		token       *models.UserTokenRecover
		setup       func() error
		expected    error
	}{
		{
			description: "succeeds when data is valid",
			token: &models.UserTokenRecover{
				Token: "token",
				User:  "507f1f77bcf86cd799439011",
			},
			setup: func() error {
				return nil
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.UserCreateToken(ctx, tc.token)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestUserTokenGet(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		token *models.UserTokenRecover
		err   error
	}

	cases := []struct {
		description string
		id          string
		setup       func() error
		expected    Expected
	}{
		{
			description: "fails when user is not found",
			id:          "000000000000000000000000",
			setup: func() error {
				return mongotest.UseFixture(fixtures.User)
			},
			expected: Expected{
				token: nil,
				err:   store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when user is found",
			id:          "507f1f77bcf86cd799439011",
			setup: func() error {
				return mongotest.UseFixture(fixtures.User)
			},
			expected: Expected{
				token: &models.UserTokenRecover{
					CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Token:     "token",
					User:      "507f1f77bcf86cd799439011",
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			token, err := mongostore.UserGetToken(ctx, tc.id)
			assert.Equal(t, tc.expected, Expected{token: token, err: err})

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestUserDeleteTokens(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		id          string
		setup       func() error
		expected    error
	}{
		{
			description: "fails when user is not found",
			id:          "000000000000000000000000",
			setup: func() error {
				return mongotest.UseFixture(fixtures.User)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when user is found",
			id:          "507f1f77bcf86cd799439011",
			setup: func() error {
				return mongotest.UseFixture(fixtures.User)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.UserDeleteTokens(ctx, tc.id)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestUserDelete(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		id          string
		setup       func() error
		expected    error
	}{
		{
			description: "fails when user is not found",
			id:          "000000000000000000000000",
			setup: func() error {
				return mongotest.UseFixture(fixtures.User)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when user is found",
			id:          "507f1f77bcf86cd799439011",
			setup: func() error {
				return mongotest.UseFixture(fixtures.User)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.UserDelete(ctx, tc.id)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestUserDetachInfo(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	user := models.User{UserData: models.UserData{Name: "name", Username: "username", Email: "user@email.com"}, UserPassword: models.NewUserPassword("password"), ID: "60af83d418d2dc3007cd445c"}

	objID, err := primitive.ObjectIDFromHex(user.ID)

	assert.NoError(t, err)

	_, _ = db.Client().Database("test").Collection("users").InsertOne(ctx, bson.M{
		"_id":      objID,
		"name":     user.Name,
		"username": user.Username,
		"password": user.HashedPassword,
		"email":    user.Email,
	})

	namespacesOwner := []*models.Namespace{
		{
			Owner: user.ID,
			Name:  "ns2",
			Members: []models.Member{
				{
					ID:   user.ID,
					Role: guard.RoleOwner,
				},
			},
		},
		{
			Owner: user.ID,
			Name:  "ns4",
			Members: []models.Member{
				{
					ID:   user.ID,
					Role: guard.RoleOwner,
				},
			},
		},
	}

	namespacesMember := []*models.Namespace{
		{
			Owner: "id2",
			Name:  "ns1",
			Members: []models.Member{
				{
					ID:   user.ID,
					Role: guard.RoleObserver,
				},
			},
		},
		{
			Owner: "id2",
			Name:  "ns3",
			Members: []models.Member{
				{
					ID:   user.ID,
					Role: guard.RoleObserver,
				},
			},
		},
		{
			Owner: "id2",
			Name:  "ns5",
			Members: []models.Member{
				{
					ID:   user.ID,
					Role: guard.RoleObserver,
				},
			},
		},
	}

	for _, n := range namespacesOwner {
		inserted, err := db.Client().Database("test").Collection("namespaces").InsertOne(ctx, n)
		t.Log(inserted.InsertedID)
		assert.NoError(t, err)
	}

	for _, n := range namespacesMember {
		inserted, err := db.Client().Database("test").Collection("namespaces").InsertOne(ctx, n)
		t.Log(inserted.InsertedID)
		assert.NoError(t, err)
	}

	u, err := mongostore.UserGetByUsername(ctx, "username")
	assert.NoError(t, err)
	assert.Equal(t, user.Username, u.Username)

	namespacesMap, err := mongostore.UserDetachInfo(ctx, user.ID)

	assert.NoError(t, err)
	assert.Equal(t, namespacesMap["owner"], namespacesOwner)
	assert.Equal(t, namespacesMap["member"], namespacesMember)
}
