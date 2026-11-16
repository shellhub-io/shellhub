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
	field EmailField
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
				m.On("UserUpdate", ctx, "000000000000000000000000", &models.UserChanges{Name: new("James S.")}).Return(nil).Once()
			},
			requested: []requestedEmailChange{{user: renamed(), field: EmailFieldEmail, email: "james@example.test"}},
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
				m.On("UserUpdate", ctx, "000000000000000000000000", &models.UserChanges{Name: new("James S.")}).Return(nil).Once()
			},
			requested: []requestedEmailChange{{user: renamed(), field: EmailFieldEmail, email: "james@example.test"}},
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
				m.On("UserResolve", ctx, store.UserEmailResolver, "james@example.test").Return(nil, store.ErrNoDocuments).Once()
				m.On("UserUpdate", ctx, "000000000000000000000000", &models.UserChanges{}).Return(nil).Once()
			},
			requested: []requestedEmailChange{{user: emailChangeUser(), field: EmailFieldEmail, email: "james@example.test"}},
			changes: func() []EmailsChange {
				after := requestedBy(emailChangeUser())
				after.RecoveryEmail = "backup@example.test"

				return []EmailsChange{{Before: emailChangeUser(), After: after, EmailHeld: true, RecoveryEmailHeld: true}}
			}(),
			errorIs: []error{ErrUserUpdate, hookErr},
		},
		{
			description: "holds a new recovery email, reports the request and hands it to the hook",
			req: &requests.UpdateUser{
				UserID: "000000000000000000000000", RecoveryEmail: "Backup@Example.test", CurrentPassword: "secret",
			},
			mocks: func(m *storemock.MockStore) {
				m.On("UserUpdate", ctx, "000000000000000000000000", &models.UserChanges{}).Return(nil).Once()
			},
			requested: []requestedEmailChange{{user: emailChangeUser(), field: EmailFieldRecoveryEmail, email: "backup@example.test"}},
			changes: []EmailsChange{{
				Before: emailChangeUser(),
				After: func() *models.User {
					u := emailChangeUser()
					u.RecoveryEmail = "backup@example.test"

					return u
				}(),
				RecoveryEmailHeld: true,
			}},
		},
		{
			description: "holds both addresses and hands each to the hook, the email first",
			req: &requests.UpdateUser{
				UserID: "000000000000000000000000", Email: "james@example.test", RecoveryEmail: "backup@example.test", CurrentPassword: "secret",
			},
			mocks: func(m *storemock.MockStore) {
				m.On("UserResolve", ctx, store.UserEmailResolver, "james@example.test").Return(nil, store.ErrNoDocuments).Once()
				m.On("UserUpdate", ctx, "000000000000000000000000", &models.UserChanges{}).Return(nil).Once()
			},
			requested: []requestedEmailChange{
				{user: emailChangeUser(), field: EmailFieldEmail, email: "james@example.test"},
				{user: emailChangeUser(), field: EmailFieldRecoveryEmail, email: "backup@example.test"},
			},
			changes: func() []EmailsChange {
				after := requestedBy(emailChangeUser())
				after.RecoveryEmail = "backup@example.test"

				return []EmailsChange{{Before: emailChangeUser(), After: after, EmailHeld: true, RecoveryEmailHeld: true}}
			}(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			var requested []requestedEmailChange
			override(t, &emailChangeRequestedHooks, []EmailChangeRequestedHookFn{
				func(_ context.Context, user *models.User, field EmailField, email string) error {
					requested = append(requested, requestedEmailChange{user: user, field: field, email: email})

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

func TestUpdateUserWritesNewAddressesWithoutAConfirmationHook(t *testing.T) {
	ctx := context.Background()

	hashMock := hashmock.NewMockHasher(t)
	hash.Set(t, hashMock)
	hashMock.On("CompareWith", "secret", emailChangeHash).Return(true).Once()

	override(t, &emailChangeRequestedHooks, nil)
	changes := recordEmailsChanged(t)

	saved := emailChangeUser()
	saved.Email = "james@example.test"
	saved.RecoveryEmail = "backup@example.test"

	storeMock := storemock.NewMockStore(t)
	storeMock.On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").Return(emailChangeUser(), nil).Once()
	storeMock.On("UserUpdate", ctx, "000000000000000000000000", &models.UserChanges{Email: new("james@example.test"), RecoveryEmail: new("backup@example.test")}).Return(nil).Once()

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache())

	require.NoError(t, s.UpdateUser(ctx, &requests.UpdateUser{
		UserID: "000000000000000000000000", Email: "james@example.test", RecoveryEmail: "backup@example.test", CurrentPassword: "secret",
	}))
	assert.Equal(t, []EmailsChange{{Before: emailChangeUser(), After: saved}}, *changes)
}

func TestConfirmEmailChange(t *testing.T) {
	ctx := context.Background()

	override(t, &emailCheckHooks, []EmailCheckHookFn{refuseDomain("@listed.test", "disposable")})

	saveErr := stderrors.New("store down")

	cases := []struct {
		description string
		field       EmailField
		email       string
		mocks       func(*storemock.MockStore)
		expected    error
	}{
		{
			description: "writes the confirmed email alone, lowercased",
			email:       "James@Example.test",
			mocks: func(m *storemock.MockStore) {
				m.On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").Return(emailChangeUser(), nil).Once()
				m.On("UserUpdate", ctx, "000000000000000000000000", &models.UserChanges{Email: new("james@example.test")}).Return(nil).Once()
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
				m.On("UserUpdate", ctx, "000000000000000000000000", &models.UserChanges{Email: new("james@example.test")}).
					Return(stderrors.Join(store.ErrDuplicate, store.DuplicateFieldError{Field: "email"})).Once()
			},
			expected: NewErrUserDuplicated([]string{"email"}, stderrors.Join(store.ErrDuplicate, store.DuplicateFieldError{Field: "email"})),
		},
		{
			description: "fails when the user cannot be saved",
			email:       "james@example.test",
			mocks: func(m *storemock.MockStore) {
				m.On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").Return(emailChangeUser(), nil).Once()
				m.On("UserUpdate", ctx, "000000000000000000000000", &models.UserChanges{Email: new("james@example.test")}).Return(saveErr).Once()
			},
			expected: NewErrUserUpdate(emailChangeUser(), saveErr),
		},
		{
			description: "writes the confirmed recovery email alone, lowercased",
			field:       EmailFieldRecoveryEmail,
			email:       "Backup@Example.test",
			mocks: func(m *storemock.MockStore) {
				m.On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").Return(emailChangeUser(), nil).Once()
				m.On("UserUpdate", ctx, "000000000000000000000000", &models.UserChanges{RecoveryEmail: new("backup@example.test")}).Return(nil).Once()
			},
		},
		{
			description: "refuses the user's email as the recovery email",
			field:       EmailFieldRecoveryEmail,
			email:       "James@shellhub.test",
			mocks: func(m *storemock.MockStore) {
				m.On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").Return(emailChangeUser(), nil).Once()
			},
			expected: NewErrInvalidFields(ErrBadRequest, map[string]string{"recovery_email": "must be different from the email"}),
		},
		{
			description: "refuses a recovery email an email check hook refuses",
			field:       EmailFieldRecoveryEmail,
			email:       "backup@listed.test",
			mocks: func(m *storemock.MockStore) {
				m.On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").Return(emailChangeUser(), nil).Once()
			},
			expected: NewErrInvalidFields(ErrBadRequest, map[string]string{"recovery_email": "disposable"}),
		},
		{
			description: "fails when the recovery email cannot be saved",
			field:       EmailFieldRecoveryEmail,
			email:       "backup@example.test",
			mocks: func(m *storemock.MockStore) {
				m.On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").Return(emailChangeUser(), nil).Once()
				m.On("UserUpdate", ctx, "000000000000000000000000", &models.UserChanges{RecoveryEmail: new("backup@example.test")}).Return(saveErr).Once()
			},
			expected: NewErrUserUpdate(emailChangeUser(), saveErr),
		},
		{
			description: "refuses a field it does not know",
			field:       EmailField("username"),
			email:       "backup@example.test",
			mocks: func(m *storemock.MockStore) {
				m.On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").Return(emailChangeUser(), nil).Once()
			},
			expected: NewErrBadRequest(nil),
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			changes := recordEmailsChanged(t)

			storeMock := storemock.NewMockStore(t)
			tc.mocks(storeMock)

			s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache())

			field := tc.field
			if field == "" {
				field = EmailFieldEmail
			}

			assert.Equal(t, tc.expected, s.ConfirmEmailChange(ctx, "000000000000000000000000", field, tc.email))
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

func TestUpdateUserAsksForTheMFACode(t *testing.T) {
	ctx := context.Background()

	hashMock := hashmock.NewMockHasher(t)
	hash.Set(t, hashMock)
	hashMock.On("CompareWith", "secret", emailChangeHash).Return(true).Maybe()
	hashMock.On("CompareWith", "wrong", emailChangeHash).Return(false).Maybe()
	hashMock.On("Do", "new-secret").Return("new-hash", nil).Maybe()

	checkErr := stderrors.New("cache down")

	withMFA := func() *models.User {
		u := emailChangeUser()
		u.MFA.Enabled = true

		return u
	}

	cases := []struct {
		description string
		user        *models.User
		req         *requests.UpdateUser
		reason      string
		checkErr    error
		checked     []string
		mocks       func(*storemock.MockStore)
		expected    error
		errorIs     []error
	}{
		{
			description: "refuses an address change without a code",
			user:        withMFA(),
			req: &requests.UpdateUser{
				UserID: "000000000000000000000000", RecoveryEmail: "backup@example.test", CurrentPassword: "secret",
			},
			mocks:    func(*storemock.MockStore) {},
			expected: NewErrInvalidFields(ErrBadRequest, map[string]string{"mfa_code": "is required while MFA is enabled"}),
		},
		{
			description: "refuses an address change with a code the hook refuses",
			user:        withMFA(),
			req: &requests.UpdateUser{
				UserID: "000000000000000000000000", Email: "james@example.test", CurrentPassword: "secret", MFACode: "000000",
			},
			reason:   "invalid code",
			checked:  []string{"000000"},
			mocks:    func(*storemock.MockStore) {},
			expected: NewErrInvalidFields(ErrBadRequest, map[string]string{"mfa_code": "invalid code"}),
		},
		{
			description: "fails, saving nothing, when the hook cannot check the code",
			user:        withMFA(),
			req: &requests.UpdateUser{
				UserID: "000000000000000000000000", Email: "james@example.test", CurrentPassword: "secret", MFACode: "123456",
			},
			checkErr: checkErr,
			checked:  []string{"123456"},
			mocks:    func(*storemock.MockStore) {},
			errorIs:  []error{ErrUserUpdate, checkErr},
		},
		{
			description: "checks the password before the code",
			user:        withMFA(),
			req: &requests.UpdateUser{
				UserID: "000000000000000000000000", Email: "james@example.test", CurrentPassword: "wrong", MFACode: "123456",
			},
			mocks:    func(*storemock.MockStore) {},
			expected: NewErrUserPasswordNotMatch(nil),
		},
		{
			description: "changes an address with a code the hook accepts",
			user:        withMFA(),
			req: &requests.UpdateUser{
				UserID: "000000000000000000000000", Email: "james@example.test", CurrentPassword: "secret", MFACode: "123456",
			},
			checked: []string{"123456"},
			mocks: func(m *storemock.MockStore) {
				m.On("UserUpdate", ctx, "000000000000000000000000", &models.UserChanges{Email: new("james@example.test")}).Return(nil).Once()
			},
		},
		{
			description: "refuses a password change without a code",
			user:        withMFA(),
			req: &requests.UpdateUser{
				UserID: "000000000000000000000000", Password: "new-secret", CurrentPassword: "secret",
			},
			mocks:    func(*storemock.MockStore) {},
			expected: NewErrInvalidFields(ErrBadRequest, map[string]string{"mfa_code": "is required while MFA is enabled"}),
		},
		{
			description: "changes the password with a code the hook accepts",
			user:        withMFA(),
			req: &requests.UpdateUser{
				UserID: "000000000000000000000000", Password: "new-secret", CurrentPassword: "secret", MFACode: "ABCDEFGHIJKLMNOP",
			},
			checked: []string{"ABCDEFGHIJKLMNOP"},
			mocks: func(m *storemock.MockStore) {
				m.On("UserUpdateRevokingTokens", ctx, "000000000000000000000000", &models.UserChanges{PasswordHash: new("new-hash")}).Return(nil).Once()
			},
		},
		{
			description: "asks no code for a change that keeps both addresses",
			user:        withMFA(),
			req:         &requests.UpdateUser{UserID: "000000000000000000000000", Name: "James S."},
			mocks: func(m *storemock.MockStore) {
				m.On("UserUpdate", ctx, "000000000000000000000000", &models.UserChanges{Name: new("James S.")}).Return(nil).Once()
			},
		},
		{
			description: "asks no code from an account without MFA",
			user:        emailChangeUser(),
			req: &requests.UpdateUser{
				UserID: "000000000000000000000000", Email: "james@example.test", CurrentPassword: "secret",
			},
			mocks: func(m *storemock.MockStore) {
				m.On("UserUpdate", ctx, "000000000000000000000000", &models.UserChanges{Email: new("james@example.test")}).Return(nil).Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			var checked []string
			override(t, &mfaCodeCheckHooks, []MFACodeCheckHookFn{
				func(_ context.Context, user *models.User, code string) (string, error) {
					assert.Equal(t, "000000000000000000000000", user.ID)
					checked = append(checked, code)

					return tc.reason, tc.checkErr
				},
			})
			override(t, &emailChangeRequestedHooks, nil)
			recordEmailsChanged(t)

			storeMock := storemock.NewMockStore(t)
			storeMock.On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").Return(tc.user, nil).Once()
			storeMock.On("UserResolve", ctx, store.UserEmailResolver, "james@example.test").Return(nil, store.ErrNoDocuments).Maybe()
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
			assert.Equal(t, tc.checked, checked)
		})
	}
}

func TestUpdateUserSpendsNoMFACodeOnARefusedChange(t *testing.T) {
	ctx := context.Background()

	hashMock := hashmock.NewMockHasher(t)
	hash.Set(t, hashMock)
	hashMock.On("CompareWith", "secret", emailChangeHash).Return(true).Maybe()
	hashMock.On("Do", "new-secret").Return("new-hash", nil).Maybe()

	withMFA := func() *models.User {
		u := emailChangeUser()
		u.MFA.Enabled = true

		return u
	}

	cases := []struct {
		description string
		email       string
		username    string
		password    string
		confirms    bool
		mocks       func(*storemock.MockStore)
		expected    error
	}{
		{
			description: "refuses an email a check refuses before the code",
			email:       "james@listed.test",
			mocks:       func(*storemock.MockStore) {},
			expected:    NewErrInvalidFields(ErrBadRequest, map[string]string{"email": "disposable"}),
		},
		{
			description: "refuses an email another account holds before the code",
			email:       "taken@example.test",
			confirms:    true,
			mocks: func(m *storemock.MockStore) {
				m.On("UserResolve", ctx, store.UserEmailResolver, "taken@example.test").Return(&models.User{ID: "111111111111111111111111"}, nil).Once()
			},
			expected: NewErrUserDuplicated([]string{"email"}, nil),
		},
		{
			description: "refuses an email another account holds before the code without a confirmation hook",
			email:       "taken@example.test",
			mocks: func(m *storemock.MockStore) {
				m.On("UserResolve", ctx, store.UserEmailResolver, "taken@example.test").Return(&models.User{ID: "111111111111111111111111"}, nil).Once()
			},
			expected: NewErrUserDuplicated([]string{"email"}, nil),
		},
		{
			description: "refuses a username another account holds before the code",
			username:    "taken",
			password:    "new-secret",
			mocks: func(m *storemock.MockStore) {
				m.On("UserResolve", ctx, store.UserUsernameResolver, "taken").Return(&models.User{ID: "111111111111111111111111"}, nil).Once()
			},
			expected: NewErrUserDuplicated([]string{"username"}, nil),
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			checked := false
			override(t, &mfaCodeCheckHooks, []MFACodeCheckHookFn{
				func(context.Context, *models.User, string) (string, error) {
					checked = true

					return "", nil
				},
			})
			override(t, &emailCheckHooks, []EmailCheckHookFn{refuseDomain("@listed.test", "disposable")})
			override(t, &emailChangeRequestedHooks, nil)
			if tc.confirms {
				override(t, &emailChangeRequestedHooks, []EmailChangeRequestedHookFn{
					func(context.Context, *models.User, EmailField, string) error { return nil },
				})
			}
			recordEmailsChanged(t)

			storeMock := storemock.NewMockStore(t)
			storeMock.On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").Return(withMFA(), nil).Once()
			tc.mocks(storeMock)

			s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache())

			err := s.UpdateUser(ctx, &requests.UpdateUser{
				UserID: "000000000000000000000000", Email: tc.email, Username: tc.username, Password: tc.password,
				CurrentPassword: "secret", MFACode: "ABCDEFGHIJKLMNOP",
			})
			assert.Equal(t, tc.expected, err)
			assert.False(t, checked, "the code was checked for a change that was refused")
		})
	}
}

func TestUpdateUserAsksNoMFACodeWithoutAnMFACheckHook(t *testing.T) {
	ctx := context.Background()

	hashMock := hashmock.NewMockHasher(t)
	hash.Set(t, hashMock)
	hashMock.On("CompareWith", "secret", emailChangeHash).Return(true).Once()

	override(t, &mfaCodeCheckHooks, nil)
	override(t, &emailChangeRequestedHooks, nil)
	recordEmailsChanged(t)

	user := emailChangeUser()
	user.MFA.Enabled = true

	storeMock := storemock.NewMockStore(t)
	storeMock.On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").Return(user, nil).Once()
	storeMock.On("UserUpdate", ctx, "000000000000000000000000", &models.UserChanges{Email: new("james@example.test")}).Return(nil).Once()

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache())

	require.NoError(t, s.UpdateUser(ctx, &requests.UpdateUser{
		UserID: "000000000000000000000000", Email: "james@example.test", CurrentPassword: "secret",
	}))
}
