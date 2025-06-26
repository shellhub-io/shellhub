package mongo_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
						Status:         models.UserStatusConfirmed,
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
						Status:         models.UserStatusConfirmed,
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
						Status:         models.UserStatusConfirmed,
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
						Status:         models.UserStatusNotConfirmed,
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
					{
						ID:             "6509e169ae6144b2f56bf288",
						CreatedAt:      time.Date(2023, 1, 5, 12, 0, 0, 0, time.UTC),
						LastLogin:      time.Date(2023, 1, 5, 12, 0, 0, 0, time.UTC),
						EmailMarketing: true,
						Status:         models.UserStatusConfirmed,
						UserData: models.UserData{
							Name:     "Maria Garcia",
							Email:    "maria.garcia@test.com",
							Username: "maria_garcia",
						},
						MaxNamespaces: 5,
						Password: models.UserPassword{
							Hash: "c2301b2b7e872843b473d2c301e4fb2e6e9f27f2e7a1b6ad44a3b2c97f1670b3",
						},
					},
				},
				count: 5,
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
						Status:         models.UserStatusConfirmed,
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
						ID:             "6509e169ae6144b2f56bf288",
						CreatedAt:      time.Date(2023, 1, 5, 12, 0, 0, 0, time.UTC),
						LastLogin:      time.Date(2023, 1, 5, 12, 0, 0, 0, time.UTC),
						EmailMarketing: true,
						Status:         models.UserStatusConfirmed,
						UserData: models.UserData{
							Name:     "Maria Garcia",
							Email:    "maria.garcia@test.com",
							Username: "maria_garcia",
						},
						MaxNamespaces: 5,
						Password: models.UserPassword{
							Hash: "c2301b2b7e872843b473d2c301e4fb2e6e9f27f2e7a1b6ad44a3b2c97f1670b3",
						},
					},
				},
				count: 2,
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

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			users, count, err := s.UserList(ctx, tc.page, tc.filters)

			sort(tc.expected.users)
			sort(users)

			assert.Equal(t, tc.expected, Expected{users: users, count: count, err: err})
		})
	}
}

func TestUserResolve(t *testing.T) {
	type Expected struct {
		user *models.User
		err  error
	}

	cases := []struct {
		description string
		resolver    store.UserResolver
		value       string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when invalid ObjectID format",
			resolver:    store.UserIDResolver,
			value:       "invalid-id",
			fixtures:    []string{fixtureUsers},
			expected: Expected{
				user: nil,
				err:  primitive.ErrInvalidHex,
			},
		},
		{
			description: "fails when user not found by ID",
			resolver:    store.UserIDResolver,
			value:       "507f1f77bcf86cd799439999",
			fixtures:    []string{fixtureUsers},
			expected: Expected{
				user: nil,
				err:  store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds resolving user by ID",
			resolver:    store.UserIDResolver,
			value:       "507f1f77bcf86cd799439011",
			fixtures:    []string{fixtureNamespaces, fixtureUsers},
			expected: Expected{
				user: &models.User{
					ID:             "507f1f77bcf86cd799439011",
					Status:         "confirmed",
					CreatedAt:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					LastLogin:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					EmailMarketing: true,
					MaxNamespaces:  0,
					Password: models.UserPassword{
						Hash: "fcf730b6d95236ecd3c9fc2d92d7b6b2bb061514961aec041d6c7a7192f592e4",
					},
					UserData: models.UserData{
						Email:    "john.doe@test.com",
						Name:     "john doe",
						Username: "john_doe",
					},
				},
				err: nil,
			},
		},
		{
			description: "fails when user not found by email",
			resolver:    store.UserEmailResolver,
			value:       "nonexistent@test.com",
			fixtures:    []string{fixtureUsers},
			expected: Expected{
				user: nil,
				err:  store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds resolving user by email",
			resolver:    store.UserEmailResolver,
			value:       "jane.smith@test.com",
			fixtures:    []string{fixtureNamespaces, fixtureUsers},
			expected: Expected{
				user: &models.User{
					ID:             "608f32a2c7351f001f6475e0",
					Status:         "confirmed",
					CreatedAt:      time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
					LastLogin:      time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
					EmailMarketing: true,
					MaxNamespaces:  3,
					Password: models.UserPassword{
						Hash: "a0b8c29f4c8d57e542f5e81d35ebe801fd27f569f116fe670e8962d798512a1d",
					},
					UserData: models.UserData{
						Email:    "jane.smith@test.com",
						Name:     "Jane Smith",
						Username: "jane_smith",
					},
				},
				err: nil,
			},
		},
		{
			description: "fails when user not found by username",
			resolver:    store.UserUsernameResolver,
			value:       "nonexistent_user",
			fixtures:    []string{fixtureUsers},
			expected: Expected{
				user: nil,
				err:  store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds resolving user by username",
			resolver:    store.UserUsernameResolver,
			value:       "bob_johnson",
			fixtures:    []string{fixtureNamespaces, fixtureUsers},
			expected: Expected{
				user: &models.User{
					ID:             "709f45b5e812c1002f3a67e7",
					Status:         "confirmed",
					CreatedAt:      time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
					LastLogin:      time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
					EmailMarketing: true,
					MaxNamespaces:  10,
					Password: models.UserPassword{
						Hash: "5f3b3956a1a150b73e6b27e674f27d7aeb01ab1a40c179c3e1aa6026a36655a2",
					},
					UserData: models.UserData{
						Email:    "bob.johnson@test.com",
						Name:     "Bob Johnson",
						Username: "bob_johnson",
					},
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() { assert.NoError(t, srv.Reset()) })

			user, err := s.UserResolve(context.Background(), tc.resolver, tc.value)
			assert.Equal(t, tc.expected, Expected{user: user, err: err})
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

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			insertedID, err := s.UserCreate(ctx, tc.user)
			assert.Equal(t, tc.expected, err)
			assert.NotEmpty(t, insertedID)
		})
	}
}

func TestStore_UserCreateInvited(t *testing.T) {
	now := time.Now()

	cases := []struct {
		description string
		email       string
		fixtures    []string
		mocks       func()
	}{
		{
			description: "succeeds",
			email:       "john.doe@test.com",
			fixtures:    []string{},
			mocks: func() {
				mockClock := new(clockmock.Clock)
				clock.DefaultBackend = mockClock
				mockClock.On("Now").Return(now)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.Background()
			tc.mocks()

			require.NoError(tt, srv.Apply(tc.fixtures...))
			tt.Cleanup(func() {
				require.NoError(tt, srv.Reset())
			})

			insertedID, err := s.UserCreateInvited(ctx, tc.email)
			require.NoError(tt, err)
			require.NotEmpty(tt, insertedID)

			objID, _ := primitive.ObjectIDFromHex(insertedID)

			tmpUser := make(map[string]interface{})
			require.NoError(tt, db.Collection("users").FindOne(ctx, bson.M{"_id": objID}).Decode(&tmpUser))
			require.Equal(
				tt,
				map[string]interface{}{
					"_id":             objID,
					"created_at":      primitive.NewDateTimeFromTime(now),
					"last_login":      primitive.NewDateTimeFromTime(time.Time{}),
					"origin":          nil,
					"external_id":     nil,
					"status":          models.UserStatusInvited.String(),
					"max_namespaces":  nil,
					"name":            nil,
					"username":        nil,
					"email":           "john.doe@test.com",
					"recovery_email":  nil,
					"email_marketing": nil,
					"password":        nil,
					"preferences":     map[string]interface{}{"preferred_namespace": nil, "auth_methods": nil},
					"mfa":             map[string]interface{}{"enabled": nil, "secret": nil, "recovery_codes": nil},
				},
				tmpUser,
			)
		})
	}
}

func TestUserConflicts(t *testing.T) {
	type Expected struct {
		conflicts []string
		ok        bool
		err       error
	}

	cases := []struct {
		description string
		target      *models.UserConflicts
		fixtures    []string
		expected    Expected
	}{
		{
			description: "no conflicts when target is empty",
			target:      &models.UserConflicts{},
			fixtures:    []string{fixtureUsers},
			expected:    Expected{[]string{}, false, nil},
		},
		{
			description: "no conflicts with non existing email",
			target:      &models.UserConflicts{Email: "other@test.com"},
			fixtures:    []string{fixtureUsers},
			expected:    Expected{[]string{}, false, nil},
		},
		{
			description: "no conflicts with non existing username",
			target:      &models.UserConflicts{Username: "other"},
			fixtures:    []string{fixtureUsers},
			expected:    Expected{[]string{}, false, nil},
		},
		{
			description: "no conflicts with non existing username and email",
			target:      &models.UserConflicts{Email: "other@test.com", Username: "other"},
			fixtures:    []string{fixtureUsers},
			expected:    Expected{[]string{}, false, nil},
		},
		{
			description: "conflict detected with existing email",
			target:      &models.UserConflicts{Email: "john.doe@test.com"},
			fixtures:    []string{fixtureUsers},
			expected:    Expected{[]string{"email"}, true, nil},
		},
		{
			description: "conflict detected with existing username",
			target:      &models.UserConflicts{Username: "john_doe"},
			fixtures:    []string{fixtureUsers},
			expected:    Expected{[]string{"username"}, true, nil},
		},
		{
			description: "conflict detected with existing username and email",
			target:      &models.UserConflicts{Email: "john.doe@test.com", Username: "john_doe"},
			fixtures:    []string{fixtureUsers},
			expected:    Expected{[]string{"username", "email"}, true, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			require.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() { require.NoError(t, srv.Reset()) })

			conflicts, ok, err := s.UserConflicts(ctx, tc.target)
			require.Equal(t, tc.expected, Expected{conflicts, ok, err})
		})
	}
}

func TestUserUpdate(t *testing.T) {
	type Expected struct {
		changes *models.UserChanges
		err     error
	}

	cases := []struct {
		description string
		id          string
		changes     *models.UserChanges
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when user is not found",
			id:          "000000000000000000000000",
			changes:     &models.UserChanges{},
			fixtures:    []string{fixtureUsers},
			expected: Expected{
				changes: nil,
				err:     store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when updating string values",
			id:          "507f1f77bcf86cd799439011",
			changes: &models.UserChanges{
				Name:   "New Value",
				Email:  "new.value@test.com",
				Status: models.UserStatusNotConfirmed,
			},
			fixtures: []string{fixtureUsers},
			expected: Expected{
				changes: &models.UserChanges{
					LastLogin: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Name:      "New Value",
					Email:     "new.value@test.com",
					Username:  "john_doe",
					Status:    models.UserStatusNotConfirmed,
					Password:  "fcf730b6d95236ecd3c9fc2d92d7b6b2bb061514961aec041d6c7a7192f592e4",
				},
				err: nil,
			},
		},
		{
			description: "succeeds when updating time values",
			id:          "507f1f77bcf86cd799439011",
			changes: &models.UserChanges{
				LastLogin: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			fixtures: []string{fixtureUsers},
			expected: Expected{
				changes: &models.UserChanges{
					LastLogin: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
					Name:      "john doe",
					Email:     "john.doe@test.com",
					Username:  "john_doe",
					Status:    models.UserStatusConfirmed,
					Password:  "fcf730b6d95236ecd3c9fc2d92d7b6b2bb061514961aec041d6c7a7192f592e4",
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			require.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() { require.NoError(t, srv.Reset()) })

			if err := s.UserUpdate(ctx, tc.id, tc.changes); err != nil {
				require.Equal(t, tc.expected.err, err)

				return
			}

			id, err := primitive.ObjectIDFromHex(tc.id)
			require.NoError(t, err)

			user := new(models.User)
			require.NoError(t, db.Collection("users").FindOne(ctx, bson.M{"_id": id}).Decode(user))

			// Ensures that only the expected attributes have been updated.
			require.Equal(t, tc.expected.changes.LastLogin, user.LastLogin)
			require.Equal(t, tc.expected.changes.Name, user.Name)
			require.Equal(t, tc.expected.changes.Email, user.Email)
			require.Equal(t, tc.expected.changes.Status, user.Status)
			require.Equal(t, tc.expected.changes.Username, user.Username)
			require.Equal(t, tc.expected.changes.Password, user.Password.Hash)
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
			expected:    store.ErrNoDocuments,
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

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			err := s.UserDelete(ctx, tc.id)
			assert.Equal(t, tc.expected, err)
		})
	}
}
