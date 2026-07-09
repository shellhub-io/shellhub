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
	cachemock "github.com/shellhub-io/shellhub/pkg/cache/mocks"
	"github.com/shellhub-io/shellhub/pkg/envs"
	envmock "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_AddNamespaceMember(t *testing.T) {
	type Expected struct {
		namespace *models.Namespace
		err       error
	}

	storeMock := storemock.NewMockStore(t)

	cases := []struct {
		description   string
		req           *requests.NamespaceAddMember
		requiredMocks func(context.Context)
		expected      Expected
	}{
		{
			description: "fails when the namespace was not found",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(nil, ErrNamespaceNotFound).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceNotFound("00000000-0000-4000-0000-000000000000", ErrNamespaceNotFound),
			},
		},
		{
			description: "fails when the active member was not found",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members:  []models.Member{},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(nil, ErrUserNotFound).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrUserNotFound("000000000000000000000000", ErrUserNotFound),
			},
		},
		{
			description: "fails when the active member is not on the namespace",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members:  []models.Member{},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, nil).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceMemberNotFound("000000000000000000000000", nil),
			},
		},
		{
			description: "fails when the passive role is owner",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleOwner,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOperator,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, nil).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrRoleForbidden(),
			},
		},
		{
			description: "fails when the active member's role cannot act over passive member's role",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleAdministrator,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOperator,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, nil).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrRoleForbidden(),
			},
		},
		{
			description: "fails when the user was not found by email",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOwner,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserEmailResolver, "john.doe@test.com").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrUserNotFound("john.doe@test.com", errors.New("error")),
			},
		},
		{
			description: "admin provisions a new account inline when the email has no account",
			req: &requests.NamespaceAddMember{
				FowardedHost:   "localhost",
				UserID:         "000000000000000000000000",
				TenantID:       "00000000-0000-4000-0000-000000000000",
				MemberEmail:    "john.doe@test.com",
				MemberRole:     authorizer.RoleObserver,
				MemberName:     "John Doe",
				MemberUsername: "john_doe",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{ID: "000000000000000000000000", Role: authorizer.RoleOwner},
						},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{ID: "000000000000000000000000", Admin: true, UserData: models.UserData{Username: "jane_doe"}}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserEmailResolver, "john.doe@test.com").
					Return(nil, errors.New("not found")).
					Once()
				storeMock.
					On("WithTransaction", ctx, mock.Anything).
					Run(func(args mock.Arguments) { _ = args.Get(1).(store.TransactionCb)(ctx) }).
					Return(nil).
					Once()
				// An admin's inline provisioning is auto-approved: the account is created
				// not-confirmed but not awaiting approval, so its link can be minted at once.
				storeMock.
					On("UserCreate", ctx, mock.MatchedBy(func(u *models.User) bool {
						return u.Status == models.UserStatusNotConfirmed &&
							!u.AwaitingApproval &&
							u.Password.Hash == "" &&
							u.Email == "john.doe@test.com" &&
							u.Username == "john_doe"
					})).
					Return("000000000000000000000001", nil).
					Once()
				clockMock.On("Now").Return(now).Once()
				storeMock.
					On("NamespaceCreateMembership", ctx, "00000000-0000-4000-0000-000000000000", mock.MatchedBy(func(m *models.Member) bool {
						return m.ID == "000000000000000000000001" && m.Role == authorizer.RoleObserver
					})).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{ID: "000000000000000000000000", Role: authorizer.RoleOwner},
							{ID: "000000000000000000000001", Role: authorizer.RoleObserver},
						},
					}, nil).
					Once()
			},
			expected: Expected{
				namespace: &models.Namespace{
					TenantID: "00000000-0000-4000-0000-000000000000",
					Name:     "namespace",
					Owner:    "000000000000000000000000",
					Members: []models.Member{
						{ID: "000000000000000000000000", Role: authorizer.RoleOwner},
						{ID: "000000000000000000000001", Role: authorizer.RoleObserver},
					},
				},
				err: nil,
			},
		},
		{
			description: "fails when an admin provisions inline without a name and username",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{ID: "000000000000000000000000", Role: authorizer.RoleOwner},
						},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{ID: "000000000000000000000000", Admin: true, UserData: models.UserData{Username: "jane_doe"}}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserEmailResolver, "john.doe@test.com").
					Return(nil, errors.New("not found")).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceMemberProvisionProfile(nil),
			},
		},
		{
			description: "fails when the member is duplicated",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOwner,
							},
							{
								ID:   "000000000000000000000001",
								Role: authorizer.RoleAdministrator,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserEmailResolver, "john.doe@test.com").
					Return(&models.User{
						ID:       "000000000000000000000001",
						UserData: models.UserData{Username: "john_doe"},
					}, nil).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceMemberDuplicated("000000000000000000000001", nil),
			},
		},
		{
			description: "fails when the transaction fails",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOwner,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserEmailResolver, "john.doe@test.com").
					Return(&models.User{
						ID:       "000000000000000000000001",
						UserData: models.UserData{Username: "john_doe"},
					}, nil).
					Once()
				storeMock.
					On("WithTransaction", ctx, mock.Anything).
					Return(errors.New("error")).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       errors.New("error"),
			},
		},
		{
			description: "succeeds creating the membership directly",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOwner,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserEmailResolver, "john.doe@test.com").
					Return(&models.User{
						ID:       "000000000000000000000001",
						UserData: models.UserData{Username: "john_doe"},
					}, nil).
					Once()
				storeMock.
					On("WithTransaction", ctx, mock.Anything).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOwner,
							},
							{
								ID:   "000000000000000000000001",
								Role: authorizer.RoleObserver,
							},
						},
					}, nil).
					Once()
			},
			expected: Expected{
				namespace: &models.Namespace{
					TenantID: "00000000-0000-4000-0000-000000000000",
					Name:     "namespace",
					Owner:    "000000000000000000000000",
					Members: []models.Member{
						{
							ID:   "000000000000000000000000",
							Role: authorizer.RoleOwner,
						},
						{
							ID:   "000000000000000000000001",
							Role: authorizer.RoleObserver,
						},
					},
				},
				err: nil,
			},
		},
	}

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.TODO()
			tc.requiredMocks(ctx)
			ns, err := s.AddNamespaceMember(ctx, tc.req)
			assert.Equal(t, tc.expected, Expected{ns, err})
		})
	}

	storeMock.AssertExpectations(t)
}

// TestService_AddNamespaceMember_ProvisionRequest covers the enterprise seam where a
// namespace admin who is *not* an instance admin adds an unknown email: with the non-admin
// provisioning capability enabled the account is provisioned awaiting approval instead of
// hitting the community "must already have an account" dead end.
func TestService_AddNamespaceMember_ProvisionRequest(t *testing.T) {
	type Expected struct {
		namespace *models.Namespace
		err       error
	}

	errNotFound := errors.New("not found")

	// The actor is a namespace administrator (Admin: false), so it cannot provision inline.
	nonAdminActor := &models.User{
		ID:       "000000000000000000000000",
		Admin:    false,
		UserData: models.UserData{Username: "jane_doe"},
	}
	namespace := &models.Namespace{
		TenantID: "00000000-0000-4000-0000-000000000000",
		Name:     "namespace",
		Owner:    "000000000000000000000000",
		Members: []models.Member{
			{ID: "000000000000000000000000", Role: authorizer.RoleAdministrator},
		},
	}

	storeMock := storemock.NewMockStore(t)

	cases := []struct {
		description    string
		withCapability bool
		req            *requests.NamespaceAddMember
		requiredMocks  func(context.Context)
		expected       Expected
	}{
		{
			description:    "returns not found when a non-admin adds an unknown email and the capability is off",
			withCapability: false,
			req: &requests.NamespaceAddMember{
				FowardedHost:   "localhost",
				UserID:         "000000000000000000000000",
				TenantID:       "00000000-0000-4000-0000-000000000000",
				MemberEmail:    "john.doe@test.com",
				MemberRole:     authorizer.RoleObserver,
				MemberName:     "John Doe",
				MemberUsername: "john_doe",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(namespace, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(nonAdminActor, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserEmailResolver, "john.doe@test.com").
					Return(nil, errNotFound).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrUserNotFound("john.doe@test.com", errNotFound),
			},
		},
		{
			description:    "requires name and username before provisioning",
			withCapability: true,
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(namespace, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(nonAdminActor, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserEmailResolver, "john.doe@test.com").
					Return(nil, errNotFound).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceMemberProvisionProfile(nil),
			},
		},
		{
			description:    "provisions an awaiting-approval account instead of returning not found",
			withCapability: true,
			req: &requests.NamespaceAddMember{
				FowardedHost:   "localhost",
				UserID:         "000000000000000000000000",
				TenantID:       "00000000-0000-4000-0000-000000000000",
				MemberEmail:    "john.doe@test.com",
				MemberRole:     authorizer.RoleObserver,
				MemberName:     "John Doe",
				MemberUsername: "john_doe",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(namespace, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(nonAdminActor, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserEmailResolver, "john.doe@test.com").
					Return(nil, errNotFound).
					Once()
				storeMock.
					On("WithTransaction", ctx, mock.Anything).
					Run(func(args mock.Arguments) { _ = args.Get(1).(store.TransactionCb)(ctx) }).
					Return(nil).
					Once()
				// A non-admin's provisioned account must be created not-confirmed, without a
				// password, and awaiting approval — this is the enterprise approval gate.
				storeMock.
					On("UserCreate", ctx, mock.MatchedBy(func(u *models.User) bool {
						return u.Status == models.UserStatusNotConfirmed &&
							u.AwaitingApproval &&
							u.Password.Hash == "" &&
							u.Origin == models.UserOriginLocal &&
							u.Email == "john.doe@test.com" &&
							u.Username == "john_doe"
					})).
					Return("000000000000000000000001", nil).
					Once()
				clockMock.On("Now").Return(now).Once()
				storeMock.
					On("NamespaceCreateMembership", ctx, "00000000-0000-4000-0000-000000000000", mock.MatchedBy(func(m *models.Member) bool {
						return m.ID == "000000000000000000000001" && m.Role == authorizer.RoleObserver
					})).
					Return(nil).
					Once()
				// provisionNamespaceMember re-resolves and returns the namespace.
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(namespace, nil).
					Once()
			},
			expected: Expected{
				namespace: namespace,
				err:       nil,
			},
		},
		{
			description:    "propagates the error when provisioning fails",
			withCapability: true,
			req: &requests.NamespaceAddMember{
				FowardedHost:   "localhost",
				UserID:         "000000000000000000000000",
				TenantID:       "00000000-0000-4000-0000-000000000000",
				MemberEmail:    "john.doe@test.com",
				MemberRole:     authorizer.RoleObserver,
				MemberName:     "John Doe",
				MemberUsername: "john_doe",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(namespace, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(nonAdminActor, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserEmailResolver, "john.doe@test.com").
					Return(nil, errNotFound).
					Once()
				storeMock.
					On("WithTransaction", ctx, mock.Anything).
					Return(errors.New("tx failed")).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       errors.New("tx failed"),
			},
		},
	}

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			if tc.withCapability {
				EnableNonAdminProvisioning()
			}
			t.Cleanup(func() { nonAdminProvisioningEnabled = false })

			ctx := context.TODO()
			tc.requiredMocks(ctx)
			ns, err := s.AddNamespaceMember(ctx, tc.req)
			assert.Equal(t, tc.expected, Expected{ns, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestService_UpdateNamespaceMember(t *testing.T) {
	envMock := envmock.NewMockBackend(t)
	storeMock := storemock.NewMockStore(t)
	cacheMock := cachemock.NewMockCache(t)

	prevEnvsBackend := envs.DefaultBackend
	t.Cleanup(func() { envs.DefaultBackend = prevEnvsBackend })
	envs.DefaultBackend = envMock

	cases := []struct {
		description   string
		req           *requests.NamespaceUpdateMember
		requiredMocks func(context.Context)
		expected      error
	}{
		{
			description: "[community|enterprise|cloud] fails when the namespace was not found",
			req: &requests.NamespaceUpdateMember{
				UserID:     "000000000000000000000000",
				TenantID:   "00000000-0000-4000-0000-000000000000",
				MemberID:   "000000000000000000000001",
				MemberRole: authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(nil, ErrNamespaceNotFound).
					Once()
			},
			expected: NewErrNamespaceNotFound("00000000-0000-4000-0000-000000000000", ErrNamespaceNotFound),
		},
		{
			description: "[community|enterprise|cloud] fails when the active member was not found",
			req: &requests.NamespaceUpdateMember{
				UserID:     "000000000000000000000000",
				TenantID:   "00000000-0000-4000-0000-000000000000",
				MemberID:   "000000000000000000000001",
				MemberRole: authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members:  []models.Member{},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(nil, ErrUserNotFound).
					Once()
			},
			expected: NewErrUserNotFound("000000000000000000000000", ErrUserNotFound),
		},
		{
			description: "[community|enterprise|cloud] fails when the active member is not on the namespace",
			req: &requests.NamespaceUpdateMember{
				UserID:     "000000000000000000000000",
				TenantID:   "00000000-0000-4000-0000-000000000000",
				MemberID:   "000000000000000000000001",
				MemberRole: authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members:  []models.Member{},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, nil).
					Once()
			},
			expected: NewErrNamespaceMemberNotFound("000000000000000000000000", nil),
		},
		{
			description: "[community|enterprise] fails when the passive member is not on the namespace",
			req: &requests.NamespaceUpdateMember{
				UserID:     "000000000000000000000000",
				TenantID:   "00000000-0000-4000-0000-000000000000",
				MemberID:   "000000000000000000000001",
				MemberRole: authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOwner,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, nil).
					Once()
			},
			expected: NewErrNamespaceMemberNotFound("000000000000000000000001", nil),
		},
		{
			description: "[community|enterprise|cloud] fails when the passive role's is owner",
			req: &requests.NamespaceUpdateMember{
				UserID:     "000000000000000000000000",
				TenantID:   "00000000-0000-4000-0000-000000000000",
				MemberID:   "000000000000000000000001",
				MemberRole: authorizer.RoleOwner,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOwner,
							},
							{
								ID:   "000000000000000000000001",
								Role: authorizer.RoleOwner,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, nil).
					Once()
			},
			expected: NewErrRoleForbidden(),
		},
		{
			// The current-role guard catches any attempt by a lower-privileged actor to
			// act on a higher-privileged passive member regardless of the requested new
			// role. Here active=operator, passive=admin, MemberRole=observer: the new-role
			// guard (active.Role.HasAuthority(req.MemberRole)) would pass because
			// Operator.HasAuthority(Observer)==true, but the current-role guard
			// (active.Role.HasAuthority(member.Role)) rejects it because
			// Operator.HasAuthority(Administrator)==false. Setting MemberRole=observer
			// (below the active operator) isolates the current-role guard from the
			// new-role guard, so removing the current-role guard would turn this into a
			// pass-through and break this test.
			description: "[community|enterprise|cloud] BFLA: fails when operator tries to act on an admin (current-role guard)",
			req: &requests.NamespaceUpdateMember{
				UserID:     "000000000000000000000000",
				TenantID:   "00000000-0000-4000-0000-000000000000",
				MemberID:   "000000000000000000000001",
				MemberRole: authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOperator,
							},
							{
								ID:   "000000000000000000000001",
								Role: authorizer.RoleAdministrator,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, nil).
					Once()
			},
			expected: NewErrRoleForbidden(),
		},
		{
			description: "[community|enterprise|cloud] fails when cannot update the member",
			req: &requests.NamespaceUpdateMember{
				UserID:     "000000000000000000000000",
				TenantID:   "00000000-0000-4000-0000-000000000000",
				MemberID:   "000000000000000000000001",
				MemberRole: authorizer.RoleAdministrator,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOwner,
							},
							{
								ID:   "000000000000000000000001",
								Role: authorizer.RoleAdministrator,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, nil).
					Once()
				storeMock.
					On("NamespaceUpdateMembership", ctx, "00000000-0000-4000-0000-000000000000", &models.Member{ID: "000000000000000000000001", Role: authorizer.RoleAdministrator}).
					Return(errors.New("error")).
					Once()
			},
			expected: errors.New("error"),
		},
		{
			description: "[community|enterprise|cloud] succeeds",
			req: &requests.NamespaceUpdateMember{
				UserID:     "000000000000000000000000",
				TenantID:   "00000000-0000-4000-0000-000000000000",
				MemberID:   "000000000000000000000001",
				MemberRole: authorizer.RoleAdministrator,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOwner,
							},
							{
								ID:   "000000000000000000000001",
								Role: authorizer.RoleAdministrator,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, nil).
					Once()
				storeMock.
					On("NamespaceUpdateMembership", ctx, "00000000-0000-4000-0000-000000000000", &models.Member{ID: "000000000000000000000001", Role: authorizer.RoleAdministrator}).
					Return(nil).
					Once()
				cacheMock.
					On("Delete", ctx, "token_00000000-0000-4000-0000-000000000000000000000000000000000001").
					Return(nil).
					Once()
			},
			expected: nil,
		},
		// BFLA: the active member must have authority over the passive member's current
		// role, not only over the requested new role. These cases cover that guard.
		{
			// An admin must not be able to modify an owner, even when the new role is
			// below owner level — the passive member's current role must also be checked.
			description: "[community|enterprise|cloud] BFLA: fails when admin tries to demote an owner",
			req: &requests.NamespaceUpdateMember{
				UserID:     "000000000000000000000000",
				TenantID:   "00000000-0000-4000-0000-000000000000",
				MemberID:   "000000000000000000000001",
				MemberRole: authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000001",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleAdministrator,
							},
							{
								ID:   "000000000000000000000001",
								Role: authorizer.RoleOwner,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, nil).
					Once()
			},
			expected: NewErrRoleForbidden(),
		},
		{
			// An admin acting on an equal-rank admin with a lower new role (demotion) is
			// permitted. This case discriminates the current-role guard (line 128) from the
			// new-role guard (line 132): the current-role guard checks
			// Admin.HasAuthority(Admin)=true (passive's *current* role), while the new-role
			// guard checks Admin.HasAuthority(Observer)=true (the requested *new* role). Both
			// pass here. Removing the current-role guard would leave this as a pass-through,
			// but the companion rejection test (operator-demotes-admin) would break, proving
			// the guard's necessity. Together the two cases pin the >= semantics of the
			// current-role check without conflating it with the new-role check.
			description: "[community|enterprise|cloud] succeeds when admin demotes equal-rank admin",
			req: &requests.NamespaceUpdateMember{
				UserID:     "000000000000000000000000",
				TenantID:   "00000000-0000-4000-0000-000000000000",
				MemberID:   "000000000000000000000001",
				MemberRole: authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000002",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleAdministrator,
							},
							{
								ID:   "000000000000000000000001",
								Role: authorizer.RoleAdministrator,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, nil).
					Once()
				storeMock.
					On("NamespaceUpdateMembership", ctx, "00000000-0000-4000-0000-000000000000", &models.Member{ID: "000000000000000000000001", Role: authorizer.RoleObserver}).
					Return(nil).
					Once()
				cacheMock.
					On("Delete", ctx, "token_00000000-0000-4000-0000-000000000000000000000000000000000001").
					Return(nil).
					Once()
			},
			expected: nil,
		},
		{
			// An owner must not be able to self-demote. The self-target guard (active.ID ==
			// member.ID) runs before the BFLA current-role guard, so this now returns
			// NewErrAuthForbidden() instead of NewErrRoleForbidden().
			description: "[community|enterprise|cloud] BFLA: fails when owner tries to self-demote",
			req: &requests.NamespaceUpdateMember{
				UserID:     "000000000000000000000000",
				TenantID:   "00000000-0000-4000-0000-000000000000",
				MemberID:   "000000000000000000000000",
				MemberRole: authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOwner,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, nil).
					Once()
			},
			expected: NewErrAuthForbidden(),
		},
		{
			// A lower-privileged member sending an omitted/empty role against a
			// higher-privileged passive member must be rejected. The current-role guard must
			// run unconditionally (not only when req.MemberRole != RoleInvalid) to prevent a
			// lower-privileged actor from performing a write (NamespaceUpdateMembership) and
			// token-invalidation (AuthUncacheToken) against a member they have no authority
			// over. Without this guard an admin could force-invalidate an owner's cached auth
			// token via a no-op role update.
			description: "[community|enterprise|cloud] BFLA: fails when admin targets owner with empty role",
			req: &requests.NamespaceUpdateMember{
				UserID:     "000000000000000000000000",
				TenantID:   "00000000-0000-4000-0000-000000000000",
				MemberID:   "000000000000000000000001",
				MemberRole: authorizer.RoleInvalid,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000001",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleAdministrator,
							},
							{
								ID:   "000000000000000000000001",
								Role: authorizer.RoleOwner,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, nil).
					Once()
			},
			expected: NewErrRoleForbidden(),
		},
		{
			// A member record with an empty/invalid role (legacy or corrupted data) must
			// be repairable via the normal API path. HasAuthority treats RoleInvalid as the
			// lowest rank (code 0), so any valid active role can act on the corrupted member.
			// This prevents a permanent lock-out where the member could never be fixed or
			// removed without direct DB intervention. Low-privileged actors cannot exploit
			// this because they would still need to pass the new-role guard for the
			// requested role, and they cannot assign RoleInvalid as a new role either.
			description: "[community|enterprise|cloud] succeeds when owner repairs a member with an invalid/empty role",
			req: &requests.NamespaceUpdateMember{
				UserID:     "000000000000000000000000",
				TenantID:   "00000000-0000-4000-0000-000000000000",
				MemberID:   "000000000000000000000001",
				MemberRole: authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000002",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOwner,
							},
							{
								ID:   "000000000000000000000001",
								Role: authorizer.RoleInvalid, // corrupted/legacy record
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, nil).
					Once()
				storeMock.
					On("NamespaceUpdateMembership", ctx, "00000000-0000-4000-0000-000000000000", &models.Member{ID: "000000000000000000000001", Role: authorizer.RoleObserver}).
					Return(nil).
					Once()
				cacheMock.
					On("Delete", ctx, "token_00000000-0000-4000-0000-000000000000000000000000000000000001").
					Return(nil).
					Once()
			},
			expected: nil,
		},
		{
			// A lower-privileged actor must NOT be able to exploit a corrupted passive
			// record to force-write or force-invalidate tokens. The current-role guard
			// (HasAuthority(RoleInvalid)==true) passes for any valid role, but the
			// new-role guard still rejects an admin trying to promote the invalid member
			// to administrator (Admin.HasAuthority(Administrator)==true passes the
			// new-role check). However, a lower actor (observer) cannot request a role
			// above their own authority (Observer.HasAuthority(Observer)==true but
			// Observer.HasAuthority(Administrator)==false). Test that operator cannot
			// assign admin to a corrupted member, verifying new-role guard still applies.
			description: "[community|enterprise|cloud] BFLA: fails when operator tries to promote invalid-role member to administrator",
			req: &requests.NamespaceUpdateMember{
				UserID:     "000000000000000000000000",
				TenantID:   "00000000-0000-4000-0000-000000000000",
				MemberID:   "000000000000000000000001",
				MemberRole: authorizer.RoleAdministrator,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000002",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOperator,
							},
							{
								ID:   "000000000000000000000001",
								Role: authorizer.RoleInvalid, // corrupted/legacy record
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, nil).
					Once()
			},
			expected: NewErrRoleForbidden(),
		},
		{
			// An administrator self-targeting this endpoint must be rejected. A
			// self-demotion would strip NamespaceEditMember authority and lock the
			// caller out permanently. To leave a namespace, use LeaveNamespace instead.
			description: "[community|enterprise|cloud] BFLA: fails when a member targets themselves (self-demote)",
			req: &requests.NamespaceUpdateMember{
				UserID:     "000000000000000000000000",
				TenantID:   "00000000-0000-4000-0000-000000000000",
				MemberID:   "000000000000000000000000",
				MemberRole: authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000002",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleAdministrator,
							},
							{
								ID:   "000000000000000000000002",
								Role: authorizer.RoleOwner,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, nil).
					Once()
				// NO NamespaceUpdateMembership call — guard returns early
				// NO cacheMock.Delete call — guard returns early
			},
			expected: NewErrAuthForbidden(),
		},
	}

	s := NewService(store.Store(storeMock), privateKey, publicKey, cacheMock, clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.TODO()
			tc.requiredMocks(ctx)
			err := s.UpdateNamespaceMember(ctx, tc.req)
			assert.Equal(t, tc.expected, err)
		})
	}

	storeMock.AssertExpectations(t)
	cacheMock.AssertExpectations(t)
}

func TestService_RemoveNamespaceMember(t *testing.T) {
	type Expected struct {
		namespace *models.Namespace
		err       error
	}

	envMock := envmock.NewMockBackend(t)
	storeMock := storemock.NewMockStore(t)

	prevEnvsBackend := envs.DefaultBackend
	t.Cleanup(func() { envs.DefaultBackend = prevEnvsBackend })
	envs.DefaultBackend = envMock

	cases := []struct {
		description   string
		req           *requests.NamespaceRemoveMember
		requiredMocks func(context.Context)
		expected      Expected
	}{
		{
			description: "[community single-namespace] deletes the orphaned account after removing its last membership",
			req: &requests.NamespaceRemoveMember{
				UserID:   "000000000000000000000000",
				TenantID: "00000000-0000-4000-0000-000000000000",
				MemberID: "000000000000000000000001",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{ID: "000000000000000000000000", Role: authorizer.RoleOwner},
							{ID: "000000000000000000000001", Role: authorizer.RoleAdministrator},
						},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{ID: "000000000000000000000000", UserData: models.UserData{Username: "jane_doe"}}, nil).
					Once()
				storeMock.
					On("NamespaceDeleteMembership", ctx, "00000000-0000-4000-0000-000000000000", &models.Member{ID: "000000000000000000000001", Role: authorizer.RoleAdministrator}).
					Return(nil).
					Once()
				// Instance bound to this namespace: single-tenant Community.
				storeMock.
					On("SystemGet", ctx).
					Return(&models.System{InstanceTenantID: "00000000-0000-4000-0000-000000000000"}, nil).
					Once()
				queryOptionsMock := new(storemock.MockQueryOptions)
				storeMock.On("Options").Return(queryOptionsMock).Once()
				queryOptionsMock.On("WithMember", "000000000000000000000001").Return(nil).Once()
				// The removed member has no remaining namespace, so the account is reclaimed.
				storeMock.
					On("NamespaceList", ctx, mock.Anything).
					Return([]models.Namespace{}, 0, nil).
					Once()
				storeMock.
					On("UserDelete", ctx, &models.User{ID: "000000000000000000000001"}).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{ID: "000000000000000000000000", Role: authorizer.RoleOwner},
						},
					}, nil).
					Once()
			},
			expected: Expected{
				namespace: &models.Namespace{
					TenantID: "00000000-0000-4000-0000-000000000000",
					Name:     "namespace",
					Owner:    "000000000000000000000000",
					Members: []models.Member{
						{ID: "000000000000000000000000", Role: authorizer.RoleOwner},
					},
				},
				err: nil,
			},
		},
		{
			description: "[community single-namespace] keeps the account when the removed member still belongs to another namespace",
			req: &requests.NamespaceRemoveMember{
				UserID:   "000000000000000000000000",
				TenantID: "00000000-0000-4000-0000-000000000000",
				MemberID: "000000000000000000000001",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{ID: "000000000000000000000000", Role: authorizer.RoleOwner},
							{ID: "000000000000000000000001", Role: authorizer.RoleAdministrator},
						},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{ID: "000000000000000000000000", UserData: models.UserData{Username: "jane_doe"}}, nil).
					Once()
				storeMock.
					On("NamespaceDeleteMembership", ctx, "00000000-0000-4000-0000-000000000000", &models.Member{ID: "000000000000000000000001", Role: authorizer.RoleAdministrator}).
					Return(nil).
					Once()
				storeMock.
					On("SystemGet", ctx).
					Return(&models.System{InstanceTenantID: "00000000-0000-4000-0000-000000000000"}, nil).
					Once()
				queryOptionsMock := new(storemock.MockQueryOptions)
				storeMock.On("Options").Return(queryOptionsMock).Once()
				queryOptionsMock.On("WithMember", "000000000000000000000001").Return(nil).Once()
				// The removed member is still in another namespace, so the guard preserves
				// the account and UserDelete must never be called.
				storeMock.
					On("NamespaceList", ctx, mock.Anything).
					Return([]models.Namespace{{TenantID: "00000000-0000-4000-0000-000000000001"}}, 1, nil).
					Once()
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{ID: "000000000000000000000000", Role: authorizer.RoleOwner},
						},
					}, nil).
					Once()
			},
			expected: Expected{
				namespace: &models.Namespace{
					TenantID: "00000000-0000-4000-0000-000000000000",
					Name:     "namespace",
					Owner:    "000000000000000000000000",
					Members: []models.Member{
						{ID: "000000000000000000000000", Role: authorizer.RoleOwner},
					},
				},
				err: nil,
			},
		},
		{
			description: "[community|enterprise|cloud] fails when the namespace was not found",
			req: &requests.NamespaceRemoveMember{
				UserID:   "000000000000000000000000",
				TenantID: "00000000-0000-4000-0000-000000000000",
				MemberID: "000000000000000000000001",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(nil, ErrNamespaceNotFound).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceNotFound("00000000-0000-4000-0000-000000000000", ErrNamespaceNotFound),
			},
		},
		{
			description: "[community|enterprise|cloud] fails when the active member was not found",
			req: &requests.NamespaceRemoveMember{
				UserID:   "000000000000000000000000",
				TenantID: "00000000-0000-4000-0000-000000000000",
				MemberID: "000000000000000000000001",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members:  []models.Member{},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(nil, ErrUserNotFound).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrUserNotFound("000000000000000000000000", ErrUserNotFound),
			},
		},
		{
			description: "[community|enterprise|cloud] fails when the active member is not on the namespace",
			req: &requests.NamespaceRemoveMember{
				UserID:   "000000000000000000000000",
				TenantID: "00000000-0000-4000-0000-000000000000",
				MemberID: "000000000000000000000001",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members:  []models.Member{},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, nil).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceMemberNotFound("000000000000000000000000", nil),
			},
		},
		{
			description: "[community|enterprise] fails when the passive member is not on the namespace",
			req: &requests.NamespaceRemoveMember{
				UserID:   "000000000000000000000000",
				TenantID: "00000000-0000-4000-0000-000000000000",
				MemberID: "000000000000000000000001",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOwner,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, nil).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceMemberNotFound("000000000000000000000001", nil),
			},
		},
		{
			description: "[community|enterprise|cloud] fails when the active member's role cannot act over passive member's role",
			req: &requests.NamespaceRemoveMember{
				UserID:   "000000000000000000000000",
				TenantID: "00000000-0000-4000-0000-000000000000",
				MemberID: "000000000000000000000001",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOperator,
							},
							{
								ID:   "000000000000000000000001",
								Role: authorizer.RoleAdministrator,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, nil).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrRoleForbidden(),
			},
		},
		{
			description: "[community|enterprise|cloud] BFLA: fails when a member tries to remove themselves",
			req: &requests.NamespaceRemoveMember{
				UserID:   "000000000000000000000000",
				TenantID: "00000000-0000-4000-0000-000000000000",
				MemberID: "000000000000000000000000",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000002",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleAdministrator,
							},
							{
								ID:   "000000000000000000000002",
								Role: authorizer.RoleOwner,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, nil).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrAuthForbidden(),
			},
		},
		{
			description: "[community|enterprise|cloud] fails when cannot remove the member",
			req: &requests.NamespaceRemoveMember{
				UserID:   "000000000000000000000000",
				TenantID: "00000000-0000-4000-0000-000000000000",
				MemberID: "000000000000000000000001",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOwner,
							},
							{
								ID:   "000000000000000000000001",
								Role: authorizer.RoleAdministrator,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, nil).
					Once()
				storeMock.
					On("NamespaceDeleteMembership", ctx, "00000000-0000-4000-0000-000000000000", &models.Member{ID: "000000000000000000000001", Role: authorizer.RoleAdministrator}).
					Return(errors.New("error")).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       errors.New("error"),
			},
		},
		{
			description: "[community|enterprise|cloud] succeeds",
			req: &requests.NamespaceRemoveMember{
				UserID:   "000000000000000000000000",
				TenantID: "00000000-0000-4000-0000-000000000000",
				MemberID: "000000000000000000000001",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOwner,
							},
							{
								ID:   "000000000000000000000001",
								Role: authorizer.RoleAdministrator,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, nil).
					Once()
				storeMock.
					On("NamespaceDeleteMembership", ctx, "00000000-0000-4000-0000-000000000000", &models.Member{ID: "000000000000000000000001", Role: authorizer.RoleAdministrator}).
					Return(nil).
					Once()
				storeMock.
					On("SystemGet", ctx).
					Return(&models.System{}, nil).
					Once()
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOwner,
							},
						},
					}, nil).
					Once()
			},
			expected: Expected{
				namespace: &models.Namespace{
					TenantID: "00000000-0000-4000-0000-000000000000",
					Name:     "namespace",
					Owner:    "000000000000000000000000",
					Members: []models.Member{
						{
							ID:   "000000000000000000000000",
							Role: authorizer.RoleOwner,
						},
					},
				},
				err: nil,
			},
		},
		// The absent-member check must be caught by FindMember in RemoveNamespaceMember,
		// never by a store-layer sentinel inside removeMember. When NamespaceDeleteMembership
		// unexpectedly returns a store error the error must propagate unchanged so callers
		// can distinguish it from a not-found-namespace response.
		{
			description: "[community|enterprise|cloud] propagates unexpected store error from NamespaceDeleteMembership unchanged",
			req: &requests.NamespaceRemoveMember{
				UserID:   "000000000000000000000000",
				TenantID: "00000000-0000-4000-0000-000000000000",
				MemberID: "000000000000000000000001",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOwner,
							},
							{
								ID:   "000000000000000000000001",
								Role: authorizer.RoleAdministrator,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, nil).
					Once()
				// Simulate an unexpected store error from NamespaceDeleteMembership (e.g.
				// TOCTOU: member disappeared after the FindMember precheck). The service
				// must propagate it unchanged — the default branch must not remap it.
				storeMock.
					On("NamespaceDeleteMembership", ctx, "00000000-0000-4000-0000-000000000000", &models.Member{ID: "000000000000000000000001", Role: authorizer.RoleAdministrator}).
					Return(store.ErrInternal).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       store.ErrInternal,
			},
		},
		{
			// A member record with a corrupted/legacy role (RoleInvalid) must not become
			// permanently un-removable. HasAuthority treats RoleInvalid as the lowest rank
			// so any valid active role can act on it. The owner must be able to remove the
			// corrupted member to restore namespace integrity without DB intervention.
			description: "[community|enterprise|cloud] succeeds when owner removes a member with an invalid/empty role",
			req: &requests.NamespaceRemoveMember{
				UserID:   "000000000000000000000000",
				TenantID: "00000000-0000-4000-0000-000000000000",
				MemberID: "000000000000000000000001",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOwner,
							},
							{
								ID:   "000000000000000000000001",
								Role: authorizer.RoleInvalid, // corrupted/legacy record
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, nil).
					Once()
				storeMock.
					On("NamespaceDeleteMembership", ctx, "00000000-0000-4000-0000-000000000000", &models.Member{ID: "000000000000000000000001", Role: authorizer.RoleInvalid}).
					Return(nil).
					Once()
				storeMock.
					On("SystemGet", ctx).
					Return(&models.System{}, nil).
					Once()
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOwner,
							},
						},
					}, nil).
					Once()
			},
			expected: Expected{
				namespace: &models.Namespace{
					TenantID: "00000000-0000-4000-0000-000000000000",
					Name:     "namespace",
					Owner:    "000000000000000000000000",
					Members: []models.Member{
						{
							ID:   "000000000000000000000000",
							Role: authorizer.RoleOwner,
						},
					},
				},
				err: nil,
			},
		},
	}

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.TODO()
			tc.requiredMocks(ctx)
			ns, err := s.RemoveNamespaceMember(ctx, tc.req)
			assert.Equal(t, tc.expected, Expected{ns, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestService_LeaveNamespace(t *testing.T) {
	type Expected struct {
		res *models.UserAuthResponse
		err error
	}

	storeMock := storemock.NewMockStore(t)
	cacheMock := cachemock.NewMockCache(t)

	cases := []struct {
		description   string
		req           *requests.LeaveNamespace
		requiredMocks func(context.Context)
		expected      Expected
	}{
		{
			description: "fails when the namespace was not found",
			req: &requests.LeaveNamespace{
				UserID:                "000000000000000000000000",
				TenantID:              "00000000-0000-4000-0000-000000000000",
				AuthenticatedTenantID: "00000000-0000-4000-0000-000000000001",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(nil, ErrNamespaceNotFound).
					Once()
			},
			expected: Expected{
				res: nil,
				err: NewErrNamespaceNotFound("00000000-0000-4000-0000-000000000000", ErrNamespaceNotFound),
			},
		},
		{
			description: "fails when the user is not on the namespace",
			req: &requests.LeaveNamespace{
				UserID:                "000000000000000000000000",
				TenantID:              "00000000-0000-4000-0000-000000000000",
				AuthenticatedTenantID: "00000000-0000-4000-0000-000000000001",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members:  []models.Member{},
					}, nil).
					Once()
			},
			expected: Expected{
				res: nil,
				err: NewErrAuthForbidden(),
			},
		},
		{
			description: "fails when the user is owner",
			req: &requests.LeaveNamespace{
				UserID:                "000000000000000000000000",
				TenantID:              "00000000-0000-4000-0000-000000000000",
				AuthenticatedTenantID: "00000000-0000-4000-0000-000000000001",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOwner,
							},
						},
					}, nil).
					Once()
			},
			expected: Expected{
				res: nil,
				err: NewErrAuthForbidden(),
			},
		},
		{
			description: "fails when cannot remove the member",
			req: &requests.LeaveNamespace{
				UserID:                "000000000000000000000000",
				TenantID:              "00000000-0000-4000-0000-000000000000",
				AuthenticatedTenantID: "00000000-0000-4000-0000-000000000001",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleAdministrator,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("NamespaceDeleteMembership", ctx, "00000000-0000-4000-0000-000000000000", &models.Member{ID: "000000000000000000000000", Role: authorizer.RoleAdministrator}).
					Return(errors.New("error")).
					Once()
			},
			expected: Expected{
				res: nil,
				err: errors.New("error"),
			},
		},
		{
			description: "succeeds",
			req: &requests.LeaveNamespace{
				UserID:                "000000000000000000000000",
				TenantID:              "00000000-0000-4000-0000-000000000000",
				AuthenticatedTenantID: "00000000-0000-4000-0000-000000000001",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleAdministrator,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("NamespaceDeleteMembership", ctx, "00000000-0000-4000-0000-000000000000", &models.Member{ID: "000000000000000000000000", Role: authorizer.RoleAdministrator}).
					Return(nil).
					Once()
			},
			expected: Expected{
				res: nil,
				err: nil,
			},
		},
		// The absent-member check is performed by FindMember (before removeMember is ever
		// called). If NamespaceDeleteMembership unexpectedly returns a store error it must
		// NOT be remapped to NewErrNamespaceMemberNotFound — that conversion was a dead
		// branch that leaked store internals into the service layer.
		{
			description: "propagates unexpected store error from NamespaceDeleteMembership unchanged",
			req: &requests.LeaveNamespace{
				UserID:                "000000000000000000000000",
				TenantID:              "00000000-0000-4000-0000-000000000000",
				AuthenticatedTenantID: "00000000-0000-4000-0000-000000000001",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleAdministrator,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("NamespaceDeleteMembership", ctx, "00000000-0000-4000-0000-000000000000", &models.Member{ID: "000000000000000000000000", Role: authorizer.RoleAdministrator}).
					Return(store.ErrInternal).
					Once()
			},
			expected: Expected{
				res: nil,
				err: store.ErrInternal,
			},
		},
		{
			description: "succeeds when TenantID is equal to AuthenticatedTenantID",
			req: &requests.LeaveNamespace{
				UserID:                "000000000000000000000000",
				TenantID:              "00000000-0000-4000-0000-000000000000",
				AuthenticatedTenantID: "00000000-0000-4000-0000-000000000000",
			},
			requiredMocks: func(ctx context.Context) {
				user := &models.User{
					ID:        "000000000000000000000000",
					Status:    models.UserStatusConfirmed,
					Origin:    models.UserOriginLocal,
					LastLogin: now,
					MFA: models.UserMFA{
						Enabled: false,
					},
					UserData: models.UserData{
						Username: "john_doe",
						Email:    "john.doe@test.com",
						Name:     "john doe",
					},
					Password: models.UserPassword{
						Hash: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi",
					},
					Preferences: models.UserPreferences{
						PreferredNamespace: "",
						AuthMethods:        []models.UserAuthMethod{models.UserAuthMethodLocal},
					},
				}
				updatedUser := &models.User{
					ID:        "000000000000000000000000",
					Status:    models.UserStatusConfirmed,
					Origin:    models.UserOriginLocal,
					LastLogin: now,
					MFA: models.UserMFA{
						Enabled: false,
					},
					UserData: models.UserData{
						Username: "john_doe",
						Email:    "john.doe@test.com",
						Name:     "john doe",
					},
					Password: models.UserPassword{
						Hash: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi",
					},
					Preferences: models.UserPreferences{
						PreferredNamespace: "",
						AuthMethods:        []models.UserAuthMethod{models.UserAuthMethodLocal},
					},
				}

				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleAdministrator,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("NamespaceDeleteMembership", ctx, "00000000-0000-4000-0000-000000000000", &models.Member{ID: "000000000000000000000000", Role: authorizer.RoleAdministrator}).
					Return(nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(user, nil).
					Once()
				storeMock.
					On("UserUpdate", ctx, updatedUser).
					Return(nil).
					Once()
				cacheMock.
					On("Delete", ctx, "token_00000000-0000-4000-0000-000000000000000000000000000000000000").
					Return(nil).
					Once()

				// NOTE: This test is a replica of TestService_CreateUserToken because this method
				// internally calls it to create another token. Since this functionality is already tested,
				// we are duplicating the test here to prevent failures. The important tests are all in the lines above.
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(user, nil).
					Once()
				storeMock.
					On("NamespaceGetPreferred", ctx, "000000000000000000000000").
					Return(nil, store.ErrNoDocuments).
					Once()
				cacheMock.
					On("Set", ctx, "token_000000000000000000000000", mock.Anything, time.Hour*72).
					Return(nil).
					Once()
			},
			expected: Expected{
				res: &models.UserAuthResponse{
					ID:          "000000000000000000000000",
					Origin:      models.UserOriginLocal.String(),
					AuthMethods: []models.UserAuthMethod{models.UserAuthMethodLocal},
					Name:        "john doe",
					User:        "john_doe",
					Email:       "john.doe@test.com",
					Tenant:      "",
					Role:        "",
					Token:       "must ignore",
				},
				err: nil,
			},
		},
	}

	s := NewService(storeMock, privateKey, publicKey, cacheMock, clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.TODO()
			tc.requiredMocks(ctx)

			res, err := s.LeaveNamespace(ctx, tc.req)
			// Since the resulting token is not crucial for the assertion and
			// difficult to mock, it is safe to ignore this field.
			if res != nil {
				res.Token = "must ignore"
			}

			assert.Equal(t, tc.expected, Expected{res, err})
		})
	}

	storeMock.AssertExpectations(t)
}
