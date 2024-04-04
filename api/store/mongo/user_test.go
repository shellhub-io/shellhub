package mongo_test

import (
	"context"
	"sort"
	"testing"
	"time"

	shstore "github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestUserList(t *testing.T) {
	type Expected struct {
		users []models.User
		count int
		err   error
	}

	cases := []struct {
		description string
		page        query.Paginator
		filters     query.Filters
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when users are found",
			page:        query.Paginator{Page: -1, PerPage: -1},
			filters:     query.Filters{},
			fixtures:    []string{fixtureUsers},
			expected: Expected{
				users: []models.User{
					{
						ID:             "507f1f77bcf86cd799439011",
						CreatedAt:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						LastLogin:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						EmailMarketing: true,
						Confirmed:      true,
						UserData: models.UserData{
							Name:     "john doe",
							Username: "john_doe",
							Email:    "john.doe@test.com",
						},
						MaxNamespaces: 0,
						Password: models.UserPassword{
							Hash: "fcf730b6d95236ecd3c9fc2d92d7b6b2bb061514961aec041d6c7a7192f592e4",
						},
					},
					{
						ID:             "608f32a2c7351f001f6475e0",
						CreatedAt:      time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						LastLogin:      time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						EmailMarketing: true,
						Confirmed:      true,
						UserData: models.UserData{
							Name:     "Jane Smith",
							Username: "jane_smith",
							Email:    "jane.smith@test.com",
						},
						MaxNamespaces: 3,
						Password: models.UserPassword{
							Hash: "a0b8c29f4c8d57e542f5e81d35ebe801fd27f569f116fe670e8962d798512a1d",
						},
					},
					{
						ID:             "709f45b5e812c1002f3a67e7",
						CreatedAt:      time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						LastLogin:      time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						EmailMarketing: true,
						Confirmed:      true,
						UserData: models.UserData{
							Name:     "Bob Johnson",
							Username: "bob_johnson",
							Email:    "bob.johnson@test.com",
						},
						MaxNamespaces: 10,
						Password: models.UserPassword{
							Hash: "5f3b3956a1a150b73e6b27e674f27d7aeb01ab1a40c179c3e1aa6026a36655a2",
						},
					},
					{
						ID:             "80fdcea1d7299c002f3a67e8",
						CreatedAt:      time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						EmailMarketing: false,
						Confirmed:      false,
						UserData: models.UserData{
							Name:     "Alex Rodriguez",
							Username: "alex_rodriguez",
							Email:    "alex.rodriguez@test.com",
						},
						MaxNamespaces: 3,
						Password: models.UserPassword{
							Hash: "c5093eb98678c7a3324825b84c6b67c1127b93786482ddbbd356e67e29b2763f",
						},
					},
				},
				count: 4,
				err:   nil,
			},
		},
		{
			description: "succeeds with filters",
			page:        query.Paginator{Page: -1, PerPage: -1},
			filters: query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "max_namespaces",
							Operator: "gt",
							Value:    "3",
						},
					},
				},
			},
			fixtures: []string{fixtureUsers},
			expected: Expected{
				users: []models.User{
					{
						ID:             "709f45b5e812c1002f3a67e7",
						CreatedAt:      time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						LastLogin:      time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						EmailMarketing: true,
						Confirmed:      true,
						UserData: models.UserData{
							Name:     "Bob Johnson",
							Username: "bob_johnson",
							Email:    "bob.johnson@test.com",
						},
						MaxNamespaces: 10,
						Password: models.UserPassword{
							Hash: "5f3b3956a1a150b73e6b27e674f27d7aeb01ab1a40c179c3e1aa6026a36655a2",
						},
					},
				},
				count: 1,
				err:   nil,
			},
		},
	}

	// Due to the non-deterministic order of applying fixtures when dealing with multiple datasets,
	// we ensure that both the expected and result arrays are correctly sorted.
	sort := func(users []models.User) {
		sort.Slice(users, func(i, j int) bool {
			return users[i].ID < users[j].ID
		})
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, db.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, db.Reset())
			})

			users, count, err := store.UserList(ctx, tc.page, tc.filters)

			sort(tc.expected.users)
			sort(users)

			assert.Equal(t, tc.expected, Expected{users: users, count: count, err: err})
		})
	}
}

func TestUserCreate(t *testing.T) {
	cases := []struct {
		description string
		user        *models.User
		fixtures    []string
		expected    error
	}{
		{
			description: "succeeds when data is valid",
			user: &models.User{
				ID: "507f1f77bcf86cd799439011",
				UserData: models.UserData{
					Name:     "john doe",
					Username: "john_doe",
					Email:    "john.doe@test.com",
				},
				Password: models.UserPassword{
					Hash: "fcf730b6d95236ecd3c9fc2d92d7b6b2bb061514961aec041d6c7a7192f592e4",
				},
			},
			fixtures: []string{},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, db.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, db.Reset())
			})

			err := store.UserCreate(ctx, tc.user)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestUserGetByUsername(t *testing.T) {
	type Expected struct {
		user *models.User
		err  error
	}

	cases := []struct {
		description string
		username    string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when user is not found",
			username:    "nonexistent",
			fixtures:    []string{fixtureUsers},
			expected: Expected{
				user: nil,
				err:  shstore.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when user is found",
			username:    "john_doe",
			fixtures:    []string{fixtureUsers},
			expected: Expected{
				user: &models.User{
					ID:             "507f1f77bcf86cd799439011",
					CreatedAt:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					LastLogin:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					EmailMarketing: true,
					Confirmed:      true,
					UserData: models.UserData{
						Name:     "john doe",
						Username: "john_doe",
						Email:    "john.doe@test.com",
					},
					MaxNamespaces: 0,
					Password: models.UserPassword{
						Hash: "fcf730b6d95236ecd3c9fc2d92d7b6b2bb061514961aec041d6c7a7192f592e4",
					},
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, db.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, db.Reset())
			})

			user, err := store.UserGetByUsername(ctx, tc.username)
			assert.Equal(t, tc.expected, Expected{user: user, err: err})
		})
	}
}

func TestUserGetByEmail(t *testing.T) {
	type Expected struct {
		user *models.User
		err  error
	}

	cases := []struct {
		description string
		email       string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when email is not found",
			email:       "nonexistent",
			fixtures:    []string{fixtureUsers},
			expected: Expected{
				user: nil,
				err:  shstore.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when email is found",
			email:       "john.doe@test.com",
			fixtures:    []string{fixtureUsers},
			expected: Expected{
				user: &models.User{
					ID:             "507f1f77bcf86cd799439011",
					CreatedAt:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					LastLogin:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					EmailMarketing: true,
					Confirmed:      true,
					UserData: models.UserData{
						Name:     "john doe",
						Username: "john_doe",
						Email:    "john.doe@test.com",
					},
					MaxNamespaces: 0,
					Password: models.UserPassword{
						Hash: "fcf730b6d95236ecd3c9fc2d92d7b6b2bb061514961aec041d6c7a7192f592e4",
					},
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, db.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, db.Reset())
			})

			user, err := store.UserGetByEmail(ctx, tc.email)
			assert.Equal(t, tc.expected, Expected{user: user, err: err})
		})
	}
}

func TestUserGetByID(t *testing.T) {
	type Expected struct {
		user *models.User
		ns   int
		err  error
	}

	cases := []struct {
		description string
		id          string
		ns          bool
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when user is not found",
			id:          "507f1f77bcf86cd7994390bb",
			fixtures:    []string{fixtureUsers, fixtureNamespaces},
			expected: Expected{
				user: nil,
				ns:   0,
				err:  shstore.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when user is found with ns equal false",
			id:          "507f1f77bcf86cd799439011",
			ns:          false,
			fixtures:    []string{fixtureUsers, fixtureNamespaces},
			expected: Expected{
				user: &models.User{
					ID:             "507f1f77bcf86cd799439011",
					CreatedAt:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					LastLogin:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					EmailMarketing: true,
					Confirmed:      true,
					UserData: models.UserData{
						Name:     "john doe",
						Username: "john_doe",
						Email:    "john.doe@test.com",
					},
					MaxNamespaces: 0,
					Password: models.UserPassword{
						Hash: "fcf730b6d95236ecd3c9fc2d92d7b6b2bb061514961aec041d6c7a7192f592e4",
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
			fixtures:    []string{fixtureUsers, fixtureNamespaces},
			expected: Expected{
				user: &models.User{
					ID:             "507f1f77bcf86cd799439011",
					CreatedAt:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					LastLogin:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					EmailMarketing: true,
					Confirmed:      true,
					UserData: models.UserData{
						Name:     "john doe",
						Username: "john_doe",
						Email:    "john.doe@test.com",
					},
					MaxNamespaces: 0,
					Password: models.UserPassword{
						Hash: "fcf730b6d95236ecd3c9fc2d92d7b6b2bb061514961aec041d6c7a7192f592e4",
					},
				},
				ns:  1,
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, db.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, db.Reset())
			})

			user, ns, err := store.UserGetByID(ctx, tc.id, tc.ns)
			assert.Equal(t, tc.expected, Expected{user: user, ns: ns, err: err})
		})
	}
}

func TestUserUpdateData(t *testing.T) {
	cases := []struct {
		description string
		id          string
		data        models.User
		fixtures    []string
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
			fixtures: []string{fixtureUsers},
			expected: shstore.ErrNoDocuments,
		},
		{
			description: "succeeds when user is found",
			id:          "507f1f77bcf86cd799439011",
			fixtures:    []string{fixtureUsers},
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
			ctx := context.Background()

			assert.NoError(t, db.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, db.Reset())
			})

			err := store.UserUpdateData(ctx, tc.id, tc.data)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestUserUpdatePassword(t *testing.T) {
	cases := []struct {
		description string
		id          string
		password    string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when user id is not valid",
			id:          "invalid",
			password:    "other_password",
			fixtures:    []string{fixtureUsers},
			expected:    shstore.ErrInvalidHex,
		},
		{
			description: "fails when user is not found",
			id:          "000000000000000000000000",
			password:    "other_password",
			fixtures:    []string{fixtureUsers},
			expected:    shstore.ErrNoDocuments,
		},
		{
			description: "succeeds when user is found",
			id:          "507f1f77bcf86cd799439011",
			password:    "other_password",
			fixtures:    []string{fixtureUsers},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, db.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, db.Reset())
			})

			err := store.UserUpdatePassword(ctx, tc.password, tc.id)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestUserUpdateAccountStatus(t *testing.T) {
	cases := []struct {
		description string
		id          string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when user id is not valid",
			id:          "invalid",
			fixtures:    []string{fixtureUsers},
			expected:    shstore.ErrInvalidHex,
		},
		{
			description: "fails when user is not found",
			id:          "000000000000000000000000",
			fixtures:    []string{fixtureUsers},
			expected:    shstore.ErrNoDocuments,
		},
		{
			description: "succeeds when user is found",
			id:          "80fdcea1d7299c002f3a67e8",
			fixtures:    []string{fixtureUsers},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, db.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, db.Reset())
			})

			err := store.UserUpdateAccountStatus(ctx, tc.id)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestUserUpdateFromAdmin(t *testing.T) {
	cases := []struct {
		description string
		id          string
		name        string
		username    string
		email       string
		password    string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when user is not found",
			id:          "000000000000000000000000",
			name:        "other name",
			username:    "other_name",
			email:       "other.email@test.com",
			password:    "other_password",
			fixtures:    []string{fixtureUsers},
			expected:    shstore.ErrNoDocuments,
		},
		{
			description: "succeeds when user is found",
			id:          "507f1f77bcf86cd799439011",
			name:        "other name",
			username:    "other_name",
			email:       "other.email@test.com",
			password:    "other_password",
			fixtures:    []string{fixtureUsers},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, db.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, db.Reset())
			})

			err := store.UserUpdateFromAdmin(ctx, tc.name, tc.username, tc.email, tc.password, tc.id)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestUserCreateToken(t *testing.T) {
	cases := []struct {
		description string
		token       *models.UserTokenRecover
		fixtures    []string
		expected    error
	}{
		{
			description: "succeeds when data is valid",
			token: &models.UserTokenRecover{
				Token: "token",
				User:  "507f1f77bcf86cd799439011",
			},
			fixtures: []string{},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, db.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, db.Reset())
			})

			err := store.UserCreateToken(ctx, tc.token)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestUserTokenGet(t *testing.T) {
	type Expected struct {
		token *models.UserTokenRecover
		err   error
	}

	cases := []struct {
		description string
		id          string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when user is not found",
			id:          "000000000000000000000000",
			fixtures:    []string{fixtureUsers, fixtureRecoveryTokens},
			expected: Expected{
				token: nil,
				err:   shstore.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when user is found",
			id:          "507f1f77bcf86cd799439011",
			fixtures:    []string{fixtureUsers, fixtureRecoveryTokens},
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
			ctx := context.Background()

			assert.NoError(t, db.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, db.Reset())
			})

			token, err := store.UserGetToken(ctx, tc.id)
			assert.Equal(t, tc.expected, Expected{token: token, err: err})
		})
	}
}

func TestUserDeleteTokens(t *testing.T) {
	cases := []struct {
		description string
		id          string
		fixtures    []string
		expected    error
	}{
		{
			description: "succeeds when user is found",
			id:          "507f1f77bcf86cd799439011",
			fixtures:    []string{fixtureUsers, fixtureRecoveryTokens},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, db.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, db.Reset())
			})

			err := store.UserDeleteTokens(ctx, tc.id)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestUserDelete(t *testing.T) {
	cases := []struct {
		description string
		id          string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when user is not found",
			id:          "000000000000000000000000",
			fixtures:    []string{fixtureUsers},
			expected:    shstore.ErrNoDocuments,
		},
		{
			description: "succeeds when user is found",
			id:          "507f1f77bcf86cd799439011",
			fixtures:    []string{fixtureUsers},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, db.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, db.Reset())
			})

			err := store.UserDelete(ctx, tc.id)
			assert.Equal(t, tc.expected, err)
		})
	}
}
