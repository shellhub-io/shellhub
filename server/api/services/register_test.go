package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/hash"
	hashmock "github.com/shellhub-io/shellhub/pkg/hash/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_RegisterUser(t *testing.T) {
	hashMock := hashmock.NewMockHasher(t)
	hash.Set(t, hashMock)

	ctx := context.TODO()
	now := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	hashMock.On("Do", "secret123").Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi", nil)

	runTx := func(_ context.Context, cb store.TransactionCb) error { return cb(ctx) }
	isAccepted := func(invitation *models.UserInvitation) bool {
		return invitation.Status == models.UserInvitationStatusAccepted
	}

	type Expected struct {
		hasToken          bool
		confirmationsSent int
		err               error
	}

	withStatus := func(status models.UserStatus) any {
		return mock.MatchedBy(func(u *models.User) bool { return u.Status == status })
	}

	cases := []struct {
		description          string
		openSignup           bool
		nonAdminProvisioning bool
		emailConfirmation    bool
		req                  requests.RegisterUser
		requiredMocks        func(storeMock *storemock.MockStore)
		expected             Expected
	}{
		{
			description: "refuses a no-sig registration on invite-only editions",
			openSignup:  false,
			req: requests.RegisterUser{
				Name: "Alice", Username: "alice", Email: "alice@test.com", Password: "secret123",
			},
			requiredMocks: func(_ *storemock.MockStore) {},
			expected:      Expected{false, 0, NewErrAuthForbidden()},
		},
		{
			description: "refuses an unresolved sig with no email even when open signup is on",
			openSignup:  true,
			req: requests.RegisterUser{
				Name: "Alice", Username: "alice", Password: "secret123", Sig: "STALECODE123",
			},
			requiredMocks: func(storeMock *storemock.MockStore) {
				storeMock.On("MembershipInvitationResolveBySig", ctx, "STALECODE123").
					Return(nil, store.ErrNoDocuments).Once()
			},
			expected: Expected{false, 0, NewErrAuthForbidden()},
		},
		{
			description: "completes an invited account by email match when open signup is on",
			openSignup:  true,
			req: requests.RegisterUser{
				Name: "Alice", Username: "alice", Email: "alice@test.com", Password: "secret123",
			},
			requiredMocks: func(storeMock *storemock.MockStore) {
				storeMock.On("UserInvitationGet", ctx, store.UserInvitationEmailResolver, "alice@test.com").
					Return(&models.UserInvitation{
						ID: "invitee", Email: "alice@test.com", Status: models.UserInvitationStatusPending,
					}, nil).Once()
				storeMock.On("SystemGet", ctx).Return(localSignInOn(), nil).Once()
				storeMock.On("WithTransaction", ctx, mock.AnythingOfType("store.TransactionCb")).
					Return(runTx).Once()
				storeMock.On("UserCreate", ctx, mock.AnythingOfType("*models.User")).
					Return("invitee", nil).Once()
				storeMock.On("UserInvitationUpdate", ctx, mock.MatchedBy(isAccepted)).
					Return(nil).Once()
			},
			expected: Expected{false, 0, nil},
		},
		{
			description: "completes an invited account with a sig confirmed when no email confirmation is wired, joining the namespace and returning a token",
			req: requests.RegisterUser{
				Name: "Alice", Username: "alice", Password: "secret123", Sig: "INVITECODE12",
			},
			requiredMocks: func(storeMock *storemock.MockStore) {
				membership := &models.MembershipInvitation{
					UserID: "invitee", TenantID: "tenant", InvitedBy: "owner", Role: authorizer.RoleOperator,
				}
				storeMock.On("MembershipInvitationResolveBySig", ctx, "INVITECODE12").
					Return(membership, nil).Twice()
				storeMock.On("UserInvitationGet", ctx, store.UserInvitationIDResolver, "invitee").
					Return(&models.UserInvitation{
						ID: "invitee", Email: "alice@test.com", Status: models.UserInvitationStatusPending,
					}, nil).Once()
				storeMock.On("WithTransaction", ctx, mock.AnythingOfType("store.TransactionCb")).
					Return(runTx).Once()
				storeMock.On("UserCreate", ctx, withStatus(models.UserStatusConfirmed)).
					Return("invitee", nil).Once()
				storeMock.On("UserInvitationUpdate", ctx, mock.MatchedBy(isAccepted)).
					Return(nil).Once()
				storeMock.On("NamespaceCreateMembership", ctx, scope.MustBounded("tenant"), mock.AnythingOfType("*models.Member")).
					Return(nil).Once()
				storeMock.On("MembershipInvitationDelete", ctx, membership).Return(nil).Once()
				storeMock.On("SystemGet", ctx).Return(localSignInOn(), nil).Twice()
				storeMock.On("UserResolve", ctx, store.UserIDResolver, "invitee").
					Return(&models.User{
						ID: "invitee", Status: models.UserStatusConfirmed, UserData: models.UserData{Username: "alice"},
						Preferences: models.UserPreferences{AuthMethods: []models.UserAuthMethod{models.UserAuthMethodLocal}},
					}, nil).Once()
				storeMock.On("NamespaceGetPreferred", ctx, "invitee").
					Return(nil, store.ErrNoDocuments).Once()
				storeMock.On("UserUpdate", ctx, mock.AnythingOfType("*models.User")).
					Return(nil).Once()
			},
			expected: Expected{true, 0, nil},
		},
		{
			description:       "completes an invited account with a sig not confirmed when an email confirmation is wired, joining the namespace and sending the confirmation",
			openSignup:        true,
			emailConfirmation: true,
			req: requests.RegisterUser{
				Name: "Alice", Username: "alice", Password: "secret123", Sig: "INVITECODE12",
			},
			requiredMocks: func(storeMock *storemock.MockStore) {
				membership := &models.MembershipInvitation{
					UserID: "invitee", TenantID: "tenant", InvitedBy: "owner", Role: authorizer.RoleOperator,
				}
				storeMock.On("MembershipInvitationResolveBySig", ctx, "INVITECODE12").
					Return(membership, nil).Twice()
				storeMock.On("SystemGet", ctx).Return(localSignInOn(), nil).Once()
				storeMock.On("UserInvitationGet", ctx, store.UserInvitationIDResolver, "invitee").
					Return(&models.UserInvitation{
						ID: "invitee", Email: "alice@test.com", Status: models.UserInvitationStatusPending,
					}, nil).Once()
				storeMock.On("WithTransaction", ctx, mock.AnythingOfType("store.TransactionCb")).
					Return(runTx).Once()
				storeMock.On("UserCreate", ctx, withStatus(models.UserStatusNotConfirmed)).
					Return("invitee", nil).Once()
				storeMock.On("UserInvitationUpdate", ctx, mock.MatchedBy(isAccepted)).
					Return(nil).Once()
				storeMock.On("NamespaceCreateMembership", ctx, scope.MustBounded("tenant"), &models.Member{
					ID: "invitee", AddedAt: now, Role: authorizer.RoleOperator,
				}).Return(nil).Once()
				storeMock.On("MembershipInvitationDelete", ctx, membership).Return(nil).Once()
			},
			expected: Expected{false, 1, nil},
		},
		{
			description:          "sets awaiting-approval for a sig registration invited by a non-admin (enterprise)",
			nonAdminProvisioning: true,
			req: requests.RegisterUser{
				Name: "Alice", Username: "alice", Password: "secret123", Sig: "INVITECODE12",
			},
			requiredMocks: func(storeMock *storemock.MockStore) {
				membership := &models.MembershipInvitation{
					UserID: "invitee", TenantID: "tenant", InvitedBy: "namespace-admin", Role: authorizer.RoleOperator,
				}
				storeMock.On("MembershipInvitationResolveBySig", ctx, "INVITECODE12").
					Return(membership, nil).Twice()
				storeMock.On("UserInvitationGet", ctx, store.UserInvitationIDResolver, "invitee").
					Return(&models.UserInvitation{
						ID: "invitee", Email: "alice@test.com", Status: models.UserInvitationStatusPending,
					}, nil).Once()
				storeMock.On("SystemGet", ctx).Return(localSignInOn(), nil).Once()
				storeMock.On("UserResolve", ctx, store.UserIDResolver, "namespace-admin").
					Return(&models.User{ID: "namespace-admin", Admin: false}, nil).Once()
				storeMock.On("WithTransaction", ctx, mock.AnythingOfType("store.TransactionCb")).
					Return(runTx).Once()
				storeMock.On("UserCreate", ctx, mock.MatchedBy(func(u *models.User) bool {
					return u.AwaitingApproval && u.Status == models.UserStatusConfirmed
				})).
					Return("invitee", nil).Once()
				storeMock.On("UserInvitationUpdate", ctx, mock.MatchedBy(isAccepted)).
					Return(nil).Once()
				storeMock.On("NamespaceCreateMembership", ctx, scope.MustBounded("tenant"), mock.AnythingOfType("*models.Member")).
					Return(nil).Once()
				storeMock.On("MembershipInvitationDelete", ctx, membership).Return(nil).Once()
			},
			expected: Expected{false, 0, nil},
		},
		{
			description: "refuses an invited registration before creating it while local sign-in is off",
			req: requests.RegisterUser{
				Name: "Alice", Username: "alice", Password: "secret123", Sig: "INVITECODE12",
			},
			requiredMocks: func(storeMock *storemock.MockStore) {
				membership := &models.MembershipInvitation{
					UserID: "invitee", TenantID: "tenant", InvitedBy: "owner", Role: authorizer.RoleOperator,
				}
				storeMock.On("MembershipInvitationResolveBySig", ctx, "INVITECODE12").
					Return(membership, nil).Once()
				storeMock.On("UserInvitationGet", ctx, store.UserInvitationIDResolver, "invitee").
					Return(&models.UserInvitation{
						ID: "invitee", Email: "alice@test.com", Status: models.UserInvitationStatusPending,
					}, nil).Once()
				storeMock.On("SystemGet", ctx).
					Return(&models.System{Authentication: &models.SystemAuthentication{Local: &models.SystemAuthenticationLocal{Enabled: false}}}, nil).Once()
			},
			expected: Expected{false, 0, NewErrAuthMethodNotAllowed(models.UserAuthMethodLocal.String())},
		},
		{
			description: "returns the conflicting field when the invited user's username is duplicated",
			req: requests.RegisterUser{
				Name: "Alice", Username: "alice", Password: "secret123", Sig: "INVITECODE12",
			},
			requiredMocks: func(storeMock *storemock.MockStore) {
				membership := &models.MembershipInvitation{
					UserID: "invitee", TenantID: "tenant", InvitedBy: "owner", Role: authorizer.RoleOperator,
				}
				storeMock.On("MembershipInvitationResolveBySig", ctx, "INVITECODE12").
					Return(membership, nil).Twice()
				storeMock.On("UserInvitationGet", ctx, store.UserInvitationIDResolver, "invitee").
					Return(&models.UserInvitation{
						ID: "invitee", Email: "alice@test.com", Status: models.UserInvitationStatusPending,
					}, nil).Once()
				storeMock.On("SystemGet", ctx).Return(localSignInOn(), nil).Once()
				dupErr := errors.Join(store.ErrDuplicate, store.DuplicateFieldError{Field: "username"})
				storeMock.On("WithTransaction", ctx, mock.AnythingOfType("store.TransactionCb")).
					Return(runTx).Once()
				storeMock.On("UserCreate", ctx, mock.AnythingOfType("*models.User")).
					Return("", dupErr).Once()
			},
			expected: Expected{
				false,
				0,
				NewErrUserDuplicated([]string{"username"}, errors.Join(store.ErrDuplicate, store.DuplicateFieldError{Field: "username"})),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			override(t, &openSignupEnabled, tc.openSignup)
			override(t, &nonAdminProvisioningEnabled, tc.nonAdminProvisioning)

			confirmationsSent := 0
			override(t, &userRegisteredHooks, nil)
			if tc.emailConfirmation {
				OnUserRegistered(func(context.Context, *models.User, string) error {
					confirmationsSent++

					return nil
				})
			}

			fixUUID(t)
			storeMock := storemock.NewMockStore(t)
			tc.requiredMocks(storeMock)

			s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), WithClock(clock.NewFixed(now)))

			res, err := s.RegisterUser(ctx, tc.req)
			assert.Equal(t, tc.expected.err, err)
			assert.Equal(t, tc.expected.confirmationsSent, confirmationsSent)
			if tc.expected.hasToken {
				assert.NotNil(t, res)
				assert.NotEmpty(t, res.Token)
			} else {
				assert.Nil(t, res)
			}

			storeMock.AssertExpectations(t)
		})
	}
}

func TestService_RegisterUser_passesTheDeviceCodeToTheRegisteredHook(t *testing.T) {
	hashMock := hashmock.NewMockHasher(t)
	hash.Set(t, hashMock)
	hashMock.On("Do", "secret123").Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi", nil)

	ctx := context.TODO()

	var received string
	override(t, &openSignupEnabled, true)
	override(t, &userRegisteredHooks, []UserRegisteredHookFn{
		func(_ context.Context, _ *models.User, deviceCode string) error {
			received = deviceCode

			return nil
		},
	})

	storeMock := storemock.NewMockStore(t)
	storeMock.On("UserInvitationGet", ctx, store.UserInvitationEmailResolver, "alice@test.com").
		Return(nil, store.ErrNoDocuments).Once()
	storeMock.On("UserCreate", ctx, mock.AnythingOfType("*models.User")).
		Return("alice", nil).Once()

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache())

	_, err := s.RegisterUser(ctx, requests.RegisterUser{
		Name: "Alice", Username: "alice", Email: "alice@test.com", Password: "secret123", DeviceCode: "wxyz-2k7q",
	})
	require.NoError(t, err)
	assert.Equal(t, "WXYZ2K7Q", received)
}

func localSignInOn() *models.System {
	return &models.System{
		Authentication: &models.SystemAuthentication{Local: &models.SystemAuthenticationLocal{Enabled: true}},
	}
}

func TestService_RegisterUser_accountDefaults(t *testing.T) {
	ctx := context.TODO()
	req := requests.RegisterUser{
		Name: "Alice", Username: "ALICE", Email: "Request@Example.com", Password: "secret123", EmailMarketing: true,
	}
	runTx := func(txCtx context.Context, cb store.TransactionCb) error { return cb(txCtx) }

	cases := []struct {
		description   string
		requiredMocks func(storeMock *storemock.MockStore)
		expectedID    string
		expectedEmail string
	}{
		{
			description: "builds an open registration from the normalized request email",
			requiredMocks: func(storeMock *storemock.MockStore) {
				storeMock.On("UserInvitationGet", ctx, store.UserInvitationEmailResolver, "request@example.com").
					Return(nil, store.ErrNoDocuments).Once()
			},
			expectedID:    "",
			expectedEmail: "request@example.com",
		},
		{
			description: "builds an invited registration from the normalized invitation email and keeps the invitation ID",
			requiredMocks: func(storeMock *storemock.MockStore) {
				invitation := &models.UserInvitation{
					ID: "invitee", Email: "Invite@Example.com", Status: models.UserInvitationStatusPending,
				}
				storeMock.On("UserInvitationGet", ctx, store.UserInvitationEmailResolver, "request@example.com").
					Return(invitation, nil).Once()
				storeMock.On("SystemGet", ctx).Return(localSignInOn(), nil).Once()
				storeMock.On("WithTransaction", ctx, mock.AnythingOfType("store.TransactionCb")).
					Return(runTx).Once()
				storeMock.On("UserInvitationUpdate", ctx, invitation).Return(nil).Once()
			},
			expectedID:    "invitee",
			expectedEmail: "invite@example.com",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			override(t, &openSignupEnabled, true)
			override(t, &userRegisteredHooks, nil)
			hashMock := hashmock.NewMockHasher(t)
			hash.Set(t, hashMock)
			hashMock.On("Do", "secret123").Return("hashed-password", nil).Once()

			storeMock := storemock.NewMockStore(t)
			tc.requiredMocks(storeMock)
			storeMock.On("UserCreate", ctx, mock.AnythingOfType("*models.User")).
				Run(func(args mock.Arguments) {
					user, ok := args.Get(1).(*models.User)
					require.True(t, ok)
					assert.Equal(t, &models.User{
						ID:             tc.expectedID,
						Origin:         models.UserOriginLocal,
						UserData:       models.UserData{Name: "Alice", Email: tc.expectedEmail, Username: "alice"},
						Password:       models.UserPassword{Plain: "secret123", Hash: "hashed-password"},
						CreatedAt:      now,
						MaxNamespaces:  1,
						EmailMarketing: true,
						Status:         models.UserStatusNotConfirmed,
						Preferences:    models.UserPreferences{AuthMethods: []models.UserAuthMethod{models.UserAuthMethodLocal}},
					}, user)
				}).Return("created-user", nil).Once()

			s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), WithClock(clock.NewFixed(now)))
			_, err := s.RegisterUser(ctx, req)
			require.NoError(t, err)
		})
	}
}
