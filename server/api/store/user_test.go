package store_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserResolveByAuthIdentifier(t *testing.T) {
	ctx := context.Background()
	user := &models.User{ID: "65fdd16b5f62f93184ec8a39"}

	type Expected struct {
		user *models.User
		err  error
	}

	tests := []struct {
		description   string
		identifier    models.UserAuthIdentifier
		requiredMocks func(m *mocks.MockStore)
		expected      Expected
	}{
		{
			description: "resolves a plain identifier as a username",
			identifier:  "john_doe",
			requiredMocks: func(m *mocks.MockStore) {
				m.On("UserResolve", ctx, store.UserUsernameResolver, "john_doe").
					Return(user, nil).
					Once()
			},
			expected: Expected{user, nil},
		},
		{
			description: "lowercases the identifier before resolving",
			identifier:  "John_Doe",
			requiredMocks: func(m *mocks.MockStore) {
				m.On("UserResolve", ctx, store.UserUsernameResolver, "john_doe").
					Return(user, nil).
					Once()
			},
			expected: Expected{user, nil},
		},
		{
			description: "resolves an email-shaped identifier as an email first",
			identifier:  "john.doe@test.com",
			requiredMocks: func(m *mocks.MockStore) {
				m.On("UserResolve", ctx, store.UserEmailResolver, "john.doe@test.com").
					Return(user, nil).
					Once()
			},
			expected: Expected{user, nil},
		},
		{
			description: "falls back to the username when no account owns the email",
			identifier:  "john.doe@personal.test",
			requiredMocks: func(m *mocks.MockStore) {
				m.On("UserResolve", ctx, store.UserEmailResolver, "john.doe@personal.test").
					Return(nil, store.ErrNoDocuments).
					Once()
				m.On("UserResolve", ctx, store.UserUsernameResolver, "john.doe@personal.test").
					Return(user, nil).
					Once()
			},
			expected: Expected{user, nil},
		},
		{
			description: "reports the last error when neither column matches",
			identifier:  "nobody@test.com",
			requiredMocks: func(m *mocks.MockStore) {
				m.On("UserResolve", ctx, store.UserEmailResolver, "nobody@test.com").
					Return(nil, store.ErrNoDocuments).
					Once()
				m.On("UserResolve", ctx, store.UserUsernameResolver, "nobody@test.com").
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expected: Expected{nil, store.ErrNoDocuments},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			m := mocks.NewMockStore(t)
			tc.requiredMocks(m)

			resolved, err := store.UserResolveByAuthIdentifier(ctx, m, tc.identifier)

			assert.Equal(t, tc.expected, Expected{resolved, err})
			require.True(t, m.AssertExpectations(t))
		})
	}
}
