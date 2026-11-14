package services

import (
	"context"
	stderrors "errors"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/hash"
	hashmock "github.com/shellhub-io/shellhub/pkg/hash/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const emailChangeHash = "$2a$10$V/6N1wsjheBVvWosVVVV2uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi"

type requestedEmailChange struct {
	user  *models.User
	email string
}

func emailChangeUser() *models.User {
	return &models.User{
		ID: "000000000000000000000000",
		UserData: models.UserData{
			Name:          "James Smith",
			Username:      "james_smith",
			Email:         "james@shellhub.test",
			RecoveryEmail: "recovery@shellhub.test",
		},
		Password: models.UserPassword{Hash: emailChangeHash},
	}
}

func recordEmailsChanged(t *testing.T) *[]EmailsChange {
	t.Helper()

	var changes []EmailsChange
	override(t, &userEmailsChangedHooks, []UserEmailsChangedHookFn{
		func(_ context.Context, change EmailsChange) {
			changes = append(changes, change)
		},
	})

	return &changes
}

func TestUpdateUserHoldsANewEmailForConfirmation(t *testing.T) {
	ctx := context.Background()

	hashMock := hashmock.NewMockHasher(t)
	hash.Set(t, hashMock)
	hashMock.On("CompareWith", "secret", emailChangeHash).Return(true).Maybe()

	hookErr := stderrors.New("mail unavailable")
	lookupErr := stderrors.New("store down")

	renamed := func() *models.User {
		u := emailChangeUser()
		u.Name = "James S."

		return u
	}

	requestedBy := func(u *models.User) *models.User {
		u.Email = "james@example.test"

		return u
	}

	cases := []struct {
		description string
		req         *requests.UpdateUser
		hookErr     error
		mocks       func(*storemock.MockStore)
		requested   []requestedEmailChange
		changes     []EmailsChange
		expected    error
		errorIs     []error
	}{
		{
			description: "saves the rest of the change, keeps the old email, reports the request and hands the new one to the hook",
			req: &requests.UpdateUser{
				UserID: "000000000000000000000000", Name: "James S.", Email: "James@Example.test", CurrentPassword: "secret",
			},
			mocks: func(m *storemock.MockStore) {
				m.On("UserResolve", ctx, store.UserEmailResolver, "james@example.test").Return(nil, store.ErrNoDocuments).Once()
				m.On("UserUpdate", ctx, renamed()).Return(nil).Once()
			},
			requested: []requestedEmailChange{{user: renamed(), email: "james@example.test"}},
			changes:   []EmailsChange{{Before: emailChangeUser(), After: requestedBy(renamed()), EmailHeld: true}},
		},
		{
			description: "refuses the account's own recovery email as the new email before holding anything",
			req: &requests.UpdateUser{
				UserID: "000000000000000000000000", Email: "Recovery@Shellhub.test", CurrentPassword: "secret",
			},
			mocks:    func(*storemock.MockStore) {},
			expected: NewErrInvalidFields(ErrBadRequest, map[string]string{"email": "must be different from the recovery email"}),
		},
		{
			description: "refuses an email another account holds before saving anything",
			req: &requests.UpdateUser{
				UserID: "000000000000000000000000", Name: "James S.", Email: "taken@example.test", CurrentPassword: "secret",
			},
			mocks: func(m *storemock.MockStore) {
				m.On("UserResolve", ctx, store.UserEmailResolver, "taken@example.test").
					Return(&models.User{ID: "111111111111111111111111"}, nil).Once()
			},
			expected: NewErrUserDuplicated([]string{"email"}, nil),
		},
		{
			description: "fails, saving nothing, when the holder of the new email cannot be looked up",
			req: &requests.UpdateUser{
				UserID: "000000000000000000000000", Email: "james@example.test", CurrentPassword: "secret",
			},
			mocks: func(m *storemock.MockStore) {
				m.On("UserResolve", ctx, store.UserEmailResolver, "james@example.test").Return(nil, lookupErr).Once()
			},
			expected: NewErrUserUpdate(emailChangeUser(), lookupErr),
		},
		{
			description: "returns the hook's failure with the rest of the change saved",
			req: &requests.UpdateUser{
				UserID: "000000000000000000000000", Name: "James S.", Email: "james@example.test", CurrentPassword: "secret",
			},
			hookErr: hookErr,
			mocks: func(m *storemock.MockStore) {
				m.On("UserResolve", ctx, store.UserEmailResolver, "james@example.test").Return(nil, store.ErrNoDocuments).Once()
				m.On("UserUpdate", ctx, renamed()).Return(nil).Once()
			},
			requested: []requestedEmailChange{{user: renamed(), email: "james@example.test"}},
			changes:   []EmailsChange{{Before: emailChangeUser(), After: requestedBy(renamed()), EmailHeld: true}},
			errorIs:   []error{ErrUserUpdate, hookErr},
		},
		{
			description: "reports the requested change, recovery email included, even when the hook then fails",
			req: &requests.UpdateUser{
				UserID: "000000000000000000000000", Email: "james@example.test", RecoveryEmail: "backup@example.test", CurrentPassword: "secret",
			},
			hookErr: hookErr,
			mocks: func(m *storemock.MockStore) {
				saved := emailChangeUser()
				saved.RecoveryEmail = "backup@example.test"
				m.On("UserResolve", ctx, store.UserEmailResolver, "james@example.test").Return(nil, store.ErrNoDocuments).Once()
				m.On("UserUpdate", ctx, saved).Return(nil).Once()
			},
			requested: func() []requestedEmailChange {
				saved := emailChangeUser()
				saved.RecoveryEmail = "backup@example.test"

				return []requestedEmailChange{{user: saved, email: "james@example.test"}}
			}(),
			changes: func() []EmailsChange {
				saved := emailChangeUser()
				saved.RecoveryEmail = "backup@example.test"

				return []EmailsChange{{Before: emailChangeUser(), After: requestedBy(saved), EmailHeld: true}}
			}(),
			errorIs: []error{ErrUserUpdate, hookErr},
		},
		{
			description: "writes a new recovery email at once and reports it changed",
			req: &requests.UpdateUser{
				UserID: "000000000000000000000000", RecoveryEmail: "backup@example.test", CurrentPassword: "secret",
			},
			mocks: func(m *storemock.MockStore) {
				saved := emailChangeUser()
				saved.RecoveryEmail = "backup@example.test"
				m.On("UserUpdate", ctx, saved).Return(nil).Once()
			},
			changes: []EmailsChange{{
				Before: emailChangeUser(),
				After: func() *models.User {
					u := emailChangeUser()
					u.RecoveryEmail = "backup@example.test"

					return u
				}(),
			}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			var requested []requestedEmailChange
			override(t, &emailChangeRequestedHooks, []EmailChangeRequestedHookFn{
				func(_ context.Context, user *models.User, email string) error {
					requested = append(requested, requestedEmailChange{user: user, email: email})

					return tc.hookErr
				},
			})
			changes := recordEmailsChanged(t)

			storeMock := storemock.NewMockStore(t)
			storeMock.On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").Return(emailChangeUser(), nil).Once()
			tc.mocks(storeMock)

			s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache())

			err := s.UpdateUser(ctx, tc.req)
			if tc.errorIs != nil {
				for _, target := range tc.errorIs {
					require.ErrorIs(t, err, target)
				}
			} else {
				assert.Equal(t, tc.expected, err)
			}
			assert.Equal(t, tc.requested, requested)
			assert.Equal(t, tc.changes, *changes)
		})
	}
}

func TestUpdateUserWritesANewEmailWithoutAConfirmationHook(t *testing.T) {
	ctx := context.Background()

	hashMock := hashmock.NewMockHasher(t)
	hash.Set(t, hashMock)
	hashMock.On("CompareWith", "secret", emailChangeHash).Return(true).Once()

	override(t, &emailChangeRequestedHooks, nil)
	changes := recordEmailsChanged(t)

	saved := emailChangeUser()
	saved.Email = "james@example.test"

	storeMock := storemock.NewMockStore(t)
	storeMock.On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").Return(emailChangeUser(), nil).Once()
	storeMock.On("UserUpdate", ctx, saved).Return(nil).Once()

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache())

	require.NoError(t, s.UpdateUser(ctx, &requests.UpdateUser{
		UserID: "000000000000000000000000", Email: "james@example.test", CurrentPassword: "secret",
	}))
	assert.Equal(t, []EmailsChange{{Before: emailChangeUser(), After: saved}}, *changes)
}

func TestConfirmEmailChange(t *testing.T) {
	ctx := context.Background()

	override(t, &emailCheckHooks, []EmailCheckHookFn{refuseDomain("@listed.test", "disposable")})

	saveErr := stderrors.New("store down")

	cases := []struct {
		description string
		email       string
		mocks       func(*storemock.MockStore)
		expected    error
	}{
		{
			description: "writes the confirmed email alone, lowercased",
			email:       "James@Example.test",
			mocks: func(m *storemock.MockStore) {
				m.On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").Return(emailChangeUser(), nil).Once()
				m.On("UserUpdateEmail", ctx, "000000000000000000000000", "james@example.test").Return(nil).Once()
			},
		},
		{
			description: "fails when the user is gone",
			email:       "james@example.test",
			mocks: func(m *storemock.MockStore) {
				m.On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").Return(nil, store.ErrNoDocuments).Once()
			},
			expected: NewErrUserNotFound("000000000000000000000000", nil),
		},
		{
			description: "refuses the user's recovery email",
			email:       "Recovery@shellhub.test",
			mocks: func(m *storemock.MockStore) {
				m.On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").Return(emailChangeUser(), nil).Once()
			},
			expected: NewErrInvalidFields(ErrBadRequest, map[string]string{"email": "must be different from the recovery email"}),
		},
		{
			description: "refuses an address an email check hook refuses",
			email:       "james@listed.test",
			mocks: func(m *storemock.MockStore) {
				m.On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").Return(emailChangeUser(), nil).Once()
			},
			expected: NewErrInvalidFields(ErrBadRequest, map[string]string{"email": "disposable"}),
		},
		{
			description: "names the email when another account took it since the request",
			email:       "james@example.test",
			mocks: func(m *storemock.MockStore) {
				m.On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").Return(emailChangeUser(), nil).Once()
				m.On("UserUpdateEmail", ctx, "000000000000000000000000", "james@example.test").
					Return(stderrors.Join(store.ErrDuplicate, store.DuplicateFieldError{Field: "email"})).Once()
			},
			expected: NewErrUserDuplicated([]string{"email"}, stderrors.Join(store.ErrDuplicate, store.DuplicateFieldError{Field: "email"})),
		},
		{
			description: "fails when the user cannot be saved",
			email:       "james@example.test",
			mocks: func(m *storemock.MockStore) {
				m.On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").Return(emailChangeUser(), nil).Once()
				m.On("UserUpdateEmail", ctx, "000000000000000000000000", "james@example.test").Return(saveErr).Once()
			},
			expected: NewErrUserUpdate(emailChangeUser(), saveErr),
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			changes := recordEmailsChanged(t)

			storeMock := storemock.NewMockStore(t)
			tc.mocks(storeMock)

			s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache())

			assert.Equal(t, tc.expected, s.ConfirmEmailChange(ctx, "000000000000000000000000", tc.email))
			assert.Empty(t, *changes, "the change was announced when it was requested")
		})
	}
}

func TestUpdateUserRefusesAnAddressChangeOnAnAccountWithoutAPassword(t *testing.T) {
	ctx := context.Background()

	ssoOnly := func() *models.User {
		u := emailChangeUser()
		u.Password = models.UserPassword{}
		u.Preferences.AuthMethods = []models.UserAuthMethod{models.UserAuthMethodSAML}

		return u
	}

	cases := []struct {
		description string
		req         *requests.UpdateUser
	}{
		{
			description: "refuses a new email whatever password is sent",
			req:         &requests.UpdateUser{UserID: "000000000000000000000000", Email: "james@example.test", CurrentPassword: "anything"},
		},
		{
			description: "refuses a new recovery email sent with no password",
			req:         &requests.UpdateUser{UserID: "000000000000000000000000", RecoveryEmail: "backup@example.test"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			storeMock := storemock.NewMockStore(t)
			storeMock.On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").Return(ssoOnly(), nil).Once()

			s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache())

			assert.Equal(t, NewErrUserPasswordNotMatch(nil), s.UpdateUser(ctx, tc.req))
		})
	}
}
