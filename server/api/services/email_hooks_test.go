package services

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/hash"
	hashmock "github.com/shellhub-io/shellhub/pkg/hash/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func refuseDomain(domain, reason string) EmailCheckHookFn {
	return func(_ context.Context, email string) string {
		if strings.HasSuffix(email, domain) {
			return reason
		}

		return ""
	}
}

func TestCheckEmail(t *testing.T) {
	override(t, &emailCheckHooks, nil)

	ctx := context.Background()

	t.Run("nil hook panics", func(t *testing.T) {
		emailCheckHooks = nil
		assert.PanicsWithValue(t, "services: OnEmailCheck called with nil hook", func() {
			OnEmailCheck(nil)
		})
	})

	t.Run("accepts every address when no hook is registered", func(t *testing.T) {
		emailCheckHooks = nil
		assert.Empty(t, checkEmail(ctx, "alice@listed.test"))
	})

	t.Run("passes the context and the address to the hook", func(t *testing.T) {
		emailCheckHooks = nil

		type requestKey struct{}
		requestCtx := context.WithValue(ctx, requestKey{}, "request")

		var gotRequest any
		var gotEmail string
		OnEmailCheck(func(ctx context.Context, email string) string {
			gotRequest, gotEmail = ctx.Value(requestKey{}), email

			return ""
		})

		assert.Empty(t, checkEmail(requestCtx, "alice@shellhub.test"))
		assert.Equal(t, "request", gotRequest)
		assert.Equal(t, "alice@shellhub.test", gotEmail)
	})

	t.Run("returns the first refusal and skips the remaining hooks", func(t *testing.T) {
		emailCheckHooks = nil

		var secondCalled bool
		OnEmailCheck(refuseDomain("@listed.test", "disposable"))
		OnEmailCheck(func(context.Context, string) string {
			secondCalled = true

			return "unreachable"
		})

		assert.Equal(t, "disposable", checkEmail(ctx, "alice@listed.test"))
		assert.False(t, secondCalled)
	})
}

func TestRegisterUserEmailCheck(t *testing.T) {
	ctx := context.TODO()
	now := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	override(t, &openSignupEnabled, true)
	override(t, &userRegisteredHooks, nil)
	override(t, &emailCheckHooks, []EmailCheckHookFn{refuseDomain("@listed.test", "disposable")})

	const hashed = "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi"

	cases := []struct {
		description   string
		email         string
		requiredMocks func(storeMock *storemock.MockStore, hashMock *hashmock.MockHasher)
		expected      error
	}{
		{
			description: "refuses an open sign-up whose email the hook refuses, naming the email field",
			email:       "Alice@Listed.Test",
			requiredMocks: func(storeMock *storemock.MockStore, hashMock *hashmock.MockHasher) {
				storeMock.On("UserInvitationGet", ctx, store.UserInvitationEmailResolver, "alice@listed.test").
					Return(nil, store.ErrNoDocuments).Once()
			},
			expected: NewErrInvalidFields(ErrBadRequest, map[string]string{"email": "disposable"}),
		},
		{
			description: "creates an open sign-up whose email the hook accepts",
			email:       "alice@shellhub.test",
			requiredMocks: func(storeMock *storemock.MockStore, hashMock *hashmock.MockHasher) {
				hashMock.On("Do", "secret123").Return(hashed, nil).Once()
				storeMock.On("UserInvitationGet", ctx, store.UserInvitationEmailResolver, "alice@shellhub.test").
					Return(nil, store.ErrNoDocuments).Once()
				storeMock.On("UserCreate", ctx, mock.AnythingOfType("*models.User")).Return("alice", nil).Once()
			},
		},
		{
			description: "completes an invited account without asking the hook, since the invitation fixed the email",
			email:       "alice@listed.test",
			requiredMocks: func(storeMock *storemock.MockStore, hashMock *hashmock.MockHasher) {
				hashMock.On("Do", "secret123").Return(hashed, nil).Once()
				storeMock.On("UserInvitationGet", ctx, store.UserInvitationEmailResolver, "alice@listed.test").
					Return(&models.UserInvitation{
						ID: "invitee", Email: "alice@listed.test", Status: models.UserInvitationStatusPending,
					}, nil).Once()
				storeMock.On("WithTransaction", ctx, mock.AnythingOfType("store.TransactionCb")).
					Return(func(_ context.Context, cb store.TransactionCb) error { return cb(ctx) }).Once()
				storeMock.On("UserCreate", ctx, mock.AnythingOfType("*models.User")).Return("invitee", nil).Once()
				storeMock.On("UserInvitationUpdate", ctx, mock.AnythingOfType("*models.UserInvitation")).Return(nil).Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			storeMock := storemock.NewMockStore(t)
			hashMock := hashmock.NewMockHasher(t)
			hash.Set(t, hashMock)
			tc.requiredMocks(storeMock, hashMock)

			s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), WithClock(clock.NewFixed(now)))

			res, err := s.RegisterUser(ctx, requests.RegisterUser{
				Name: "Alice", Username: "alice", Email: tc.email, Password: "secret123",
			}, "shellhub.test", "")

			assert.Nil(t, res)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestUpdateUserEmailCheck(t *testing.T) {
	override(t, &emailCheckHooks, []EmailCheckHookFn{refuseDomain("@listed.test", "disposable")})

	ctx := context.Background()

	current := func() *models.User {
		return &models.User{
			ID: "000000000000000000000000",
			UserData: models.UserData{
				Name:          "James Smith",
				Username:      "james_smith",
				Email:         "james@shellhub.test",
				RecoveryEmail: "recovery@shellhub.test",
			},
		}
	}

	cases := []struct {
		description string
		user        *models.User
		req         *requests.UpdateUser
		saves       bool
		expected    error
	}{
		{
			description: "refuses a new email the hook refuses, naming the email field",
			user:        current(),
			req:         &requests.UpdateUser{UserID: "000000000000000000000000", Email: "james@listed.test"},
			expected:    NewErrInvalidFields(ErrBadRequest, map[string]string{"email": "disposable"}),
		},
		{
			description: "refuses a new recovery email the hook refuses, naming the recovery_email field",
			user:        current(),
			req:         &requests.UpdateUser{UserID: "000000000000000000000000", RecoveryEmail: "backup@listed.test"},
			expected:    NewErrInvalidFields(ErrBadRequest, map[string]string{"recovery_email": "disposable"}),
		},
		{
			description: "names both fields when both new addresses are refused",
			user:        current(),
			req: &requests.UpdateUser{
				UserID: "000000000000000000000000", Email: "james@listed.test", RecoveryEmail: "backup@listed.test",
			},
			expected: NewErrInvalidFields(ErrBadRequest, map[string]string{
				"email": "disposable", "recovery_email": "disposable",
			}),
		},
		{
			description: "saves a new email the hook accepts",
			user:        current(),
			req:         &requests.UpdateUser{UserID: "000000000000000000000000", Email: "james@example.test"},
			saves:       true,
		},
		{
			description: "does not ask the hook about an email that is not changing, so a name change still saves",
			user: func() *models.User {
				u := current()
				u.Email = "james@listed.test"

				return u
			}(),
			req: &requests.UpdateUser{
				UserID: "000000000000000000000000", Name: "James S.", Email: "James@Listed.Test",
			},
			saves: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			storeMock := storemock.NewMockStore(t)
			storeMock.On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").Return(tc.user, nil).Once()
			if tc.saves {
				storeMock.On("UserUpdate", ctx, mock.AnythingOfType("*models.User")).Return(nil).Once()
			}

			s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache())

			assert.Equal(t, tc.expected, s.UpdateUser(ctx, tc.req))
		})
	}
}
