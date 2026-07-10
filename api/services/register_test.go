package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	storemock "github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_RegisterUser(t *testing.T) {
	ctx := context.TODO()
	now := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	// A shared hash expectation: the exact plaintext varies per case but the returned digest is
	// irrelevant to the assertions (users are matched with AnythingOfType).
	hashMock.On("Do", "secret123").Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi", nil)

	// runTx executes the transaction callback so the inner store writes are exercised.
	runTx := func(_ context.Context, cb store.TransactionCb) error { return cb(ctx) }

	type Expected struct {
		hasToken  bool
		conflicts []string
		err       error
	}

	cases := []struct {
		description          string
		openSignup           bool
		nonAdminProvisioning bool
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
			expected:      Expected{false, nil, NewErrAuthForbidden()},
		},
		{
			// A non-resolving Sig with an empty email must not fall through to createNewUser:
			// request validation only requires Email when Sig is absent, so this would otherwise
			// persist a blank-email account on open-signup editions.
			description: "refuses an unresolved sig with no email even when open signup is on",
			openSignup:  true,
			req: requests.RegisterUser{
				Name: "Alice", Username: "alice", Password: "secret123", Sig: "STALECODE123",
			},
			requiredMocks: func(storeMock *storemock.MockStore) {
				storeMock.On("MembershipInvitationResolveBySig", ctx, "STALECODE123").
					Return(nil, store.ErrNoDocuments).Once()
			},
			expected: Expected{false, nil, NewErrAuthForbidden()},
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
				storeMock.On("WithTransaction", ctx, mock.AnythingOfType("store.TransactionCb")).
					Return(runTx).Once()
				storeMock.On("UserCreate", ctx, mock.AnythingOfType("*models.User")).
					Return("invitee", nil).Once()
				storeMock.On("UserInvitationUpdate", ctx, mock.AnythingOfType("*models.UserInvitation")).
					Return(nil).Once()
			},
			expected: Expected{false, nil, nil},
		},
		{
			description: "completes an invited account with a sig, joining the namespace and returning a token",
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
				storeMock.On("UserCreate", ctx, mock.AnythingOfType("*models.User")).
					Return("invitee", nil).Once()
				storeMock.On("UserInvitationUpdate", ctx, mock.AnythingOfType("*models.UserInvitation")).
					Return(nil).Once()
				storeMock.On("NamespaceCreateMembership", ctx, "tenant", mock.AnythingOfType("*models.Member")).
					Return(nil).Once()
				storeMock.On("MembershipInvitationDelete", ctx, membership).Return(nil).Once()
				// CreateUserToken (post-commit): resolves the user, then finds no preferred namespace.
				storeMock.On("UserResolve", ctx, store.UserIDResolver, "invitee").
					Return(&models.User{
						ID: "invitee", Status: models.UserStatusConfirmed, UserData: models.UserData{Username: "alice"},
					}, nil).Once()
				storeMock.On("NamespaceGetPreferred", ctx, "invitee").
					Return(nil, store.ErrNoDocuments).Once()
			},
			expected: Expected{true, nil, nil},
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
				// The inviter is a namespace admin but not a system admin, so the account is gated.
				storeMock.On("UserResolve", ctx, store.UserIDResolver, "namespace-admin").
					Return(&models.User{ID: "namespace-admin", Admin: false}, nil).Once()
				storeMock.On("WithTransaction", ctx, mock.AnythingOfType("store.TransactionCb")).
					Return(runTx).Once()
				storeMock.On("UserCreate", ctx, mock.MatchedBy(func(u *models.User) bool { return u.AwaitingApproval })).
					Return("invitee", nil).Once()
				storeMock.On("UserInvitationUpdate", ctx, mock.AnythingOfType("*models.UserInvitation")).
					Return(nil).Once()
				storeMock.On("NamespaceCreateMembership", ctx, "tenant", mock.AnythingOfType("*models.Member")).
					Return(nil).Once()
				storeMock.On("MembershipInvitationDelete", ctx, membership).Return(nil).Once()
			},
			expected: Expected{false, nil, nil},
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
				dupErr := errors.Join(store.ErrDuplicate, store.DuplicateFieldError{Field: "username"})
				storeMock.On("WithTransaction", ctx, mock.AnythingOfType("store.TransactionCb")).
					Return(runTx).Once()
				storeMock.On("UserCreate", ctx, mock.AnythingOfType("*models.User")).
					Return("", dupErr).Once()
			},
			expected: Expected{
				false,
				[]string{"username"},
				NewErrUserDuplicated([]string{"username"}, errors.Join(store.ErrDuplicate, store.DuplicateFieldError{Field: "username"})),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			openSignupEnabled = tc.openSignup
			nonAdminProvisioningEnabled = tc.nonAdminProvisioning
			t.Cleanup(func() {
				openSignupEnabled = false
				nonAdminProvisioningEnabled = false
			})

			mockClockNow(t, now)
			storeMock := storemock.NewMockStore(t)
			tc.requiredMocks(storeMock)

			s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

			res, conflicts, err := s.RegisterUser(ctx, tc.req, "shellhub.test")
			assert.Equal(t, tc.expected.conflicts, conflicts)
			assert.Equal(t, tc.expected.err, err)
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
