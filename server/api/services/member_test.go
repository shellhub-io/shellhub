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
	cachemock "github.com/shellhub-io/shellhub/pkg/cache/mocks"
	"github.com/shellhub-io/shellhub/pkg/envs"
	envmock "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_AddNamespaceMember(t *testing.T) {
	type Expected struct {
		namespace *models.Namespace
		err       error
	}

	storeMock := storemock.NewMockStore(t)

	now := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		description   string
		req           *requests.NamespaceAddMember
		requiredMocks func(context.Context)
		expected      Expected
	}{
		{
			description: "fails when the namespace was not found",
			req: &requests.NamespaceAddMember{
				ForwardedHost: "localhost",
				UserID:        "000000000000000000000000",
				TenantID:      "00000000-0000-4000-0000-000000000000",
				MemberEmail:   "john.doe@test.com",
				MemberRole:    authorizer.RoleObserver,
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
			description: "fails when the active member is not on the namespace",
			req: &requests.NamespaceAddMember{
				ForwardedHost: "localhost",
				UserID:        "000000000000000000000000",
				TenantID:      "00000000-0000-4000-0000-000000000000",
				MemberEmail:   "john.doe@test.com",
				MemberRole:    authorizer.RoleObserver,
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
				namespace: nil,
				err:       NewErrNamespaceMemberNotFound("000000000000000000000000", nil),
			},
		},
		{
			description: "fails when the passive role is owner",
			req: &requests.NamespaceAddMember{
				ForwardedHost: "localhost",
				UserID:        "000000000000000000000000",
				TenantID:      "00000000-0000-4000-0000-000000000000",
				MemberEmail:   "john.doe@test.com",
				MemberRole:    authorizer.RoleOwner,
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
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrRoleForbidden(),
			},
		},
		{
			description: "fails when the active member's role cannot act over passive member's role",
			req: &requests.NamespaceAddMember{
				ForwardedHost: "localhost",
				UserID:        "000000000000000000000000",
				TenantID:      "00000000-0000-4000-0000-000000000000",
				MemberEmail:   "john.doe@test.com",
				MemberRole:    authorizer.RoleAdministrator,
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
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrRoleForbidden(),
			},
		},
		{
			description: "succeeds inviting a brand-new email via the invitation flow",
			req: &requests.NamespaceAddMember{
				ForwardedHost: "localhost",
				UserID:        "000000000000000000000000",
				TenantID:      "00000000-0000-4000-0000-000000000000",
				MemberEmail:   "john.doe@test.com",
				MemberRole:    authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				ns := &models.Namespace{
					TenantID: "00000000-0000-4000-0000-000000000000",
					Name:     "namespace",
					Owner:    "000000000000000000000000",
					Members:  []models.Member{{ID: "000000000000000000000000", Role: authorizer.RoleOwner}},
				}
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(ns, nil).
					Twice()
				storeMock.
					On("WithTransaction", ctx, mock.AnythingOfType("store.TransactionCb")).
					Return(func(ctx context.Context, cb store.TransactionCb) error { return cb(ctx) }).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserEmailResolver, "john.doe@test.com").
					Return(nil, store.ErrNoDocuments).
					Once()
				storeMock.
					On("UserInvitationsUpsert", ctx, "john.doe@test.com").
					Return("placeholder-id", nil).
					Once()
				storeMock.
					On("MembershipInvitationResolve", ctx, scope.MustBounded("00000000-0000-4000-0000-000000000000"), "placeholder-id").
					Return(nil, store.ErrNoDocuments).
					Once()
				storeMock.
					On("MembershipInvitationCreate", ctx, mock.AnythingOfType("*models.MembershipInvitation")).
					Return(nil).
					Once()
			},
			expected: Expected{
				namespace: &models.Namespace{
					TenantID: "00000000-0000-4000-0000-000000000000",
					Name:     "namespace",
					Owner:    "000000000000000000000000",
					Members:  []models.Member{{ID: "000000000000000000000000", Role: authorizer.RoleOwner}},
				},
				err: nil,
			},
		},
	}

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache())

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.TODO()
			mockClockNow(t, now)
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
			},
			expected: NewErrRoleForbidden(),
		},
		{
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
					On("NamespaceUpdateMembership", ctx, scope.MustBounded("00000000-0000-4000-0000-000000000000"), &models.Member{ID: "000000000000000000000001", Role: authorizer.RoleAdministrator}).
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
					On("NamespaceUpdateMembership", ctx, scope.MustBounded("00000000-0000-4000-0000-000000000000"), &models.Member{ID: "000000000000000000000001", Role: authorizer.RoleAdministrator}).
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
			},
			expected: NewErrRoleForbidden(),
		},
		{
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
					On("NamespaceUpdateMembership", ctx, scope.MustBounded("00000000-0000-4000-0000-000000000000"), &models.Member{ID: "000000000000000000000001", Role: authorizer.RoleObserver}).
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
			},
			expected: NewErrAuthForbidden(),
		},
		{
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
			},
			expected: NewErrRoleForbidden(),
		},
		{
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
					On("NamespaceUpdateMembership", ctx, scope.MustBounded("00000000-0000-4000-0000-000000000000"), &models.Member{ID: "000000000000000000000001", Role: authorizer.RoleObserver}).
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
			},
			expected: NewErrRoleForbidden(),
		},
		{
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
			},
			expected: NewErrAuthForbidden(),
		},
	}

	s := NewService(store.Store(storeMock), privateKey, publicKey, cacheMock)

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
					On("NamespaceDeleteMembership", ctx, scope.MustBounded("00000000-0000-4000-0000-000000000000"), &models.Member{ID: "000000000000000000000001", Role: authorizer.RoleAdministrator}).
					Return(nil).
					Once()
				storeMock.
					On("APIKeyDeleteAllByCreator", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000001").
					Return(nil).
					Once()
				storeMock.
					On("SystemGet", ctx).
					Return(&models.System{InstanceTenantID: "00000000-0000-4000-0000-000000000000"}, nil).
					Once()
				queryOptionsMock := new(storemock.MockQueryOptions)
				storeMock.On("Options").Return(queryOptionsMock).Once()
				queryOptionsMock.On("WithMember", "000000000000000000000001").Return(nil).Once()
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
					On("NamespaceDeleteMembership", ctx, scope.MustBounded("00000000-0000-4000-0000-000000000000"), &models.Member{ID: "000000000000000000000001", Role: authorizer.RoleAdministrator}).
					Return(nil).
					Once()
				storeMock.
					On("APIKeyDeleteAllByCreator", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000001").
					Return(nil).
					Once()
				storeMock.
					On("SystemGet", ctx).
					Return(&models.System{InstanceTenantID: "00000000-0000-4000-0000-000000000000"}, nil).
					Once()
				queryOptionsMock := new(storemock.MockQueryOptions)
				storeMock.On("Options").Return(queryOptionsMock).Once()
				queryOptionsMock.On("WithMember", "000000000000000000000001").Return(nil).Once()
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
					On("NamespaceDeleteMembership", ctx, scope.MustBounded("00000000-0000-4000-0000-000000000000"), &models.Member{ID: "000000000000000000000001", Role: authorizer.RoleAdministrator}).
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
					On("NamespaceDeleteMembership", ctx, scope.MustBounded("00000000-0000-4000-0000-000000000000"), &models.Member{ID: "000000000000000000000001", Role: authorizer.RoleAdministrator}).
					Return(nil).
					Once()
				storeMock.
					On("APIKeyDeleteAllByCreator", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000001").
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
					On("NamespaceDeleteMembership", ctx, scope.MustBounded("00000000-0000-4000-0000-000000000000"), &models.Member{ID: "000000000000000000000001", Role: authorizer.RoleAdministrator}).
					Return(store.ErrInternal).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       store.ErrInternal,
			},
		},
		{
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
					On("NamespaceDeleteMembership", ctx, scope.MustBounded("00000000-0000-4000-0000-000000000000"), &models.Member{ID: "000000000000000000000001", Role: authorizer.RoleInvalid}).
					Return(nil).
					Once()
				storeMock.
					On("APIKeyDeleteAllByCreator", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000001").
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

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache())

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
				err: NewErrNamespaceMemberNotFound("000000000000000000000000", nil),
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
					On("NamespaceDeleteMembership", ctx, scope.MustBounded("00000000-0000-4000-0000-000000000000"), &models.Member{ID: "000000000000000000000000", Role: authorizer.RoleAdministrator}).
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
					On("NamespaceDeleteMembership", ctx, scope.MustBounded("00000000-0000-4000-0000-000000000000"), &models.Member{ID: "000000000000000000000000", Role: authorizer.RoleAdministrator}).
					Return(nil).
					Once()
				storeMock.
					On("APIKeyDeleteAllByCreator", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000000").
					Return(nil).
					Once()
			},
			expected: Expected{
				res: nil,
				err: nil,
			},
		},
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
					On("NamespaceDeleteMembership", ctx, scope.MustBounded("00000000-0000-4000-0000-000000000000"), &models.Member{ID: "000000000000000000000000", Role: authorizer.RoleAdministrator}).
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
					On("NamespaceDeleteMembership", ctx, scope.MustBounded("00000000-0000-4000-0000-000000000000"), &models.Member{ID: "000000000000000000000000", Role: authorizer.RoleAdministrator}).
					Return(nil).
					Once()
				storeMock.
					On("APIKeyDeleteAllByCreator", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000000").
					Return(nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(user, nil).
					Once()
				storeMock.
					On("UserUpdatePreferredNamespace", ctx, "000000000000000000000000", "").
					Return(nil).
					Once()
				cacheMock.
					On("Delete", ctx, "token_00000000-0000-4000-0000-000000000000000000000000000000000000").
					Return(nil).
					Once()

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

	s := NewService(storeMock, privateKey, publicKey, cacheMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.TODO()
			tc.requiredMocks(ctx)

			res, err := s.LeaveNamespace(ctx, tc.req)
			if res != nil {
				res.Token = "must ignore"
			}

			assert.Equal(t, tc.expected, Expected{res, err})
		})
	}

	storeMock.AssertExpectations(t)
}

// TestService_AddNamespaceMember_LowercasesEmail pins that the intake flow lowercases the invited
// email before both the account lookup and the placeholder upsert, so inviting the same person in
// different letter-casing never produces a second placeholder account.
func TestService_AddNamespaceMember_LowercasesEmail(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	ctx := context.TODO()
	now := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	mockClockNow(t, now)

	ns := &models.Namespace{
		TenantID: "00000000-0000-4000-0000-000000000000",
		Owner:    "000000000000000000000000",
		Members:  []models.Member{{ID: "000000000000000000000000", Role: authorizer.RoleOwner}},
	}

	storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, ns.TenantID).Return(ns, nil).Twice()
	storeMock.On("WithTransaction", ctx, mock.AnythingOfType("store.TransactionCb")).
		Return(func(ctx context.Context, cb store.TransactionCb) error { return cb(ctx) }).Once()
	storeMock.On("UserResolve", ctx, store.UserEmailResolver, "jane@test.com").
		Return(nil, store.ErrNoDocuments).Once()
	storeMock.On("UserInvitationsUpsert", ctx, "jane@test.com").Return("placeholder", nil).Once()
	storeMock.On("MembershipInvitationResolve", ctx, scope.MustBounded(ns.TenantID), "placeholder").Return(nil, store.ErrNoDocuments).Once()
	storeMock.On("MembershipInvitationCreate", ctx, mock.AnythingOfType("*models.MembershipInvitation")).Return(nil).Once()

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache())
	_, err := s.AddNamespaceMember(ctx, &requests.NamespaceAddMember{
		ForwardedHost: "localhost",
		UserID:        "000000000000000000000000",
		TenantID:      ns.TenantID,
		MemberEmail:   "Jane@Test.com",
		MemberRole:    authorizer.RoleObserver,
	})

	require.NoError(t, err)
	storeMock.AssertExpectations(t)
}

// TestService_AddNamespaceMember_FiresNotification pins that the post-commit hook receives a
// fully-populated notification assembled from the resolved invitee, the created invitation, and the
// request — the whole reason the worker no longer needs to reload anything.
func TestService_AddNamespaceMember_FiresNotification(t *testing.T) {
	original := membershipInvitedHooks
	t.Cleanup(func() { membershipInvitedHooks = original })
	membershipInvitedHooks = nil

	storeMock := storemock.NewMockStore(t)
	ctx := context.TODO()
	now := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	mockClockNow(t, now)

	ns := &models.Namespace{
		TenantID: "00000000-0000-4000-0000-000000000000",
		Owner:    "000000000000000000000000",
		Members:  []models.Member{{ID: "000000000000000000000000", Role: authorizer.RoleOwner}},
	}

	storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, ns.TenantID).Return(ns, nil).Twice()
	storeMock.On("WithTransaction", ctx, mock.AnythingOfType("store.TransactionCb")).
		Return(func(ctx context.Context, cb store.TransactionCb) error { return cb(ctx) }).Once()
	storeMock.On("UserResolve", ctx, store.UserEmailResolver, "invitee@test.com").
		Return(&models.User{ID: "invitee", UserData: models.UserData{Name: "Invitee Person", Email: "invitee@test.com"}}, nil).Once()
	storeMock.On("MembershipInvitationResolve", ctx, scope.MustBounded(ns.TenantID), "invitee").Return(nil, store.ErrNoDocuments).Once()

	var created *models.MembershipInvitation
	storeMock.On("MembershipInvitationCreate", ctx, mock.AnythingOfType("*models.MembershipInvitation")).
		Run(func(args mock.Arguments) {
			invitation, ok := args.Get(1).(*models.MembershipInvitation)
			require.True(t, ok)
			created = invitation
		}).
		Return(nil).Once()

	var got *models.MembershipInvitationNotification
	OnMembershipInvited(func(_ context.Context, n *models.MembershipInvitationNotification) error {
		got = n

		return nil
	})

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache())
	_, err := s.AddNamespaceMember(ctx, &requests.NamespaceAddMember{
		ForwardedHost:  "localhost",
		ForwardedProto: "https",
		UserID:         "000000000000000000000000",
		TenantID:       ns.TenantID,
		MemberEmail:    "Invitee@Test.com",
		MemberRole:     authorizer.RoleObserver,
	})

	require.NoError(t, err)
	require.NotNil(t, got)
	require.NotNil(t, created)
	assert.Equal(t, "invitee@test.com", got.RecipientEmail)
	assert.Equal(t, "Invitee Person", got.RecipientName)
	assert.Equal(t, "https", got.ForwardedProto)
	assert.Equal(t, "localhost", got.ForwardedHost)
	assert.Equal(t, now.Add(7*24*time.Hour), got.ExpiresAt)
	assert.Equal(t, created.Sig, got.Signature)
	assert.NotEmpty(t, got.Signature)
}

// TestService_AddNamespaceMember_DirectMembershipFiresNoHook pins the direct-membership
// short-circuit: an existing account is added straight to the namespace, no invitation is created,
// and no delivery hook fires.
func TestService_AddNamespaceMember_DirectMembershipFiresNoHook(t *testing.T) {
	original := membershipInvitedHooks
	t.Cleanup(func() { membershipInvitedHooks = original })
	membershipInvitedHooks = nil

	directMembershipEnabled = true
	t.Cleanup(func() { directMembershipEnabled = false })

	storeMock := storemock.NewMockStore(t)
	ctx := context.TODO()
	now := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	mockClockNow(t, now)

	ns := &models.Namespace{
		TenantID: "00000000-0000-4000-0000-000000000000",
		Owner:    "000000000000000000000000",
		Members:  []models.Member{{ID: "000000000000000000000000", Role: authorizer.RoleOwner}},
	}

	storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, ns.TenantID).Return(ns, nil).Twice()
	storeMock.On("WithTransaction", ctx, mock.AnythingOfType("store.TransactionCb")).
		Return(func(ctx context.Context, cb store.TransactionCb) error { return cb(ctx) }).Once()
	storeMock.On("UserResolve", ctx, store.UserEmailResolver, "invitee@test.com").
		Return(&models.User{ID: "invitee"}, nil).Once()
	storeMock.On("NamespaceCreateMembership", ctx, scope.MustBounded(ns.TenantID), &models.Member{
		ID: "invitee", AddedAt: now, Role: authorizer.RoleObserver,
	}).Return(nil).Once()

	hookCalled := false
	OnMembershipInvited(func(context.Context, *models.MembershipInvitationNotification) error {
		hookCalled = true

		return nil
	})

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache())
	_, err := s.AddNamespaceMember(ctx, &requests.NamespaceAddMember{
		ForwardedHost: "localhost",
		UserID:        "000000000000000000000000",
		TenantID:      ns.TenantID,
		MemberEmail:   "invitee@test.com",
		MemberRole:    authorizer.RoleObserver,
	})

	require.NoError(t, err)
	assert.False(t, hookCalled)
	storeMock.AssertExpectations(t)
}
