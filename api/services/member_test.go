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
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
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

	storeMock := new(storemock.Store)

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
				err:       NewErrRoleInvalid(),
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
				err:       NewErrRoleInvalid(),
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

func TestService_UpdateNamespaceMember(t *testing.T) {
	envMock := new(envmock.Backend)
	storeMock := new(storemock.Store)

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
			expected: NewErrRoleInvalid(),
		},
		{
			description: "[community|enterprise|cloud] fails when the active member's role cannot act over passive member's role",
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
			expected: NewErrRoleInvalid(),
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
			},
			expected: nil,
		},
	}

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.TODO()
			tc.requiredMocks(ctx)
			err := s.UpdateNamespaceMember(ctx, tc.req)
			assert.Equal(t, tc.expected, err)
		})
	}

	storeMock.AssertExpectations(t)
}

func TestService_RemoveNamespaceMember(t *testing.T) {
	type Expected struct {
		namespace *models.Namespace
		err       error
	}

	envMock := new(envmock.Backend)
	storeMock := new(storemock.Store)

	envs.DefaultBackend = envMock

	cases := []struct {
		description   string
		req           *requests.NamespaceRemoveMember
		requiredMocks func(context.Context)
		expected      Expected
	}{
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
				err:       NewErrRoleInvalid(),
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

	storeMock := new(storemock.Store)
	cacheMock := new(cachemock.Cache)

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
				clockMock := new(clockmock.Clock)
				clock.DefaultBackend = clockMock
				clockMock.On("Now").Return(now)
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
