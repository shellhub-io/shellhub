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

func TestAddNamespaceMember(t *testing.T) {
	type Expected struct {
		namespace *models.Namespace
		err       error
	}

	envMock := new(envmock.Backend)
	storeMock := new(storemock.Store)
	clockMock := new(clockmock.Clock)

	envs.DefaultBackend = envMock
	clock.DefaultBackend = clockMock

	now := time.Now()
	clockMock.On("Now").Return(now)

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
			description: "fails when the passive role's is owner",
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
								ID:     "000000000000000000000000",
								Role:   authorizer.RoleOperator,
								Status: models.MemberStatusAccepted,
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
								ID:     "000000000000000000000000",
								Role:   authorizer.RoleOperator,
								Status: models.MemberStatusAccepted,
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
			description: "fails when passive member was not found",
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
								ID:     "000000000000000000000000",
								Role:   authorizer.RoleOwner,
								Status: models.MemberStatusAccepted,
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
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrUserNotFound("john.doe@test.com", errors.New("error")),
			},
		},
		{
			description: "fails when the member is duplicated without 'pending' status and expiration date not reached",
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
								ID:     "000000000000000000000000",
								Role:   authorizer.RoleOwner,
								Status: models.MemberStatusAccepted,
							},
							{
								ID:     "000000000000000000000001",
								Role:   authorizer.RoleAdministrator,
								Status: models.MemberStatusAccepted,
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
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceMemberDuplicated("000000000000000000000001", nil),
			},
		},
		{
			description: "[cloud] fails when the member is duplicated without 'pending' status and expiration date not reached",
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
								ID:     "000000000000000000000000",
								Role:   authorizer.RoleOwner,
								Status: models.MemberStatusAccepted,
							},
							{
								ID:     "000000000000000000000001",
								Role:   authorizer.RoleAdministrator,
								Status: models.MemberStatusAccepted,
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
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceMemberDuplicated("000000000000000000000001", nil),
			},
		},
		{
			description: "[cloud] fails when the member is duplicated with 'pending' status and expiration date not reached",
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
								ID:     "000000000000000000000000",
								Role:   authorizer.RoleOwner,
								Status: models.MemberStatusAccepted,
							},
							{
								ID:        "000000000000000000000001",
								Role:      authorizer.RoleAdministrator,
								Status:    models.MemberStatusPending,
								ExpiresAt: time.Now().Add(7 * (24 * time.Hour)),
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
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceMemberDuplicated("000000000000000000000001", nil),
			},
		},
		{
			description: "[cloud] succeeds to resend the invite",
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
								ID:     "000000000000000000000000",
								Role:   authorizer.RoleOwner,
								Status: models.MemberStatusAccepted,
							},
							{
								ID:        "000000000000000000000001",
								Role:      authorizer.RoleAdministrator,
								Status:    models.MemberStatusPending,
								ExpiresAt: time.Date(2023, 0o1, 0o1, 12, 0o0, 0o0, 0o0, time.UTC),
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
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				storeMock.
					On("WithTransaction", ctx, mock.Anything).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-4000-0000-000000000000",
							Name:     "namespace",
							Owner:    "000000000000000000000000",
							Members: []models.Member{
								{
									ID:     "000000000000000000000000",
									Role:   authorizer.RoleOwner,
									Status: models.MemberStatusAccepted,
								},
								{
									ID:     "000000000000000000000001",
									Role:   authorizer.RoleObserver,
									Status: models.MemberStatusPending,
								},
							},
						},
						nil,
					).
					Once()
			},
			expected: Expected{
				namespace: &models.Namespace{
					TenantID: "00000000-0000-4000-0000-000000000000",
					Name:     "namespace",
					Owner:    "000000000000000000000000",
					Members: []models.Member{
						{
							ID:     "000000000000000000000000",
							Role:   authorizer.RoleOwner,
							Status: models.MemberStatusAccepted,
						},
						{
							ID:     "000000000000000000000001",
							Role:   authorizer.RoleObserver,
							Status: models.MemberStatusPending,
						},
					},
				},
				err: nil,
			},
		},
		{
			description: "[cloud] succeeds to create the member when not found",
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
								ID:     "000000000000000000000000",
								Role:   authorizer.RoleOwner,
								Status: models.MemberStatusAccepted,
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
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				storeMock.
					On("UserCreateInvited", ctx, "john.doe@test.com").
					Return("000000000000000000000001", nil).
					Once()
				storeMock.
					On("WithTransaction", ctx, mock.Anything).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-4000-0000-000000000000",
							Name:     "namespace",
							Owner:    "000000000000000000000000",
							Members: []models.Member{
								{
									ID:     "000000000000000000000000",
									Role:   authorizer.RoleOwner,
									Status: models.MemberStatusAccepted,
								},
								{
									ID:     "000000000000000000000001",
									Role:   authorizer.RoleObserver,
									Status: models.MemberStatusPending,
								},
							},
						},
						nil,
					).
					Once()
			},
			expected: Expected{
				namespace: &models.Namespace{
					TenantID: "00000000-0000-4000-0000-000000000000",
					Name:     "namespace",
					Owner:    "000000000000000000000000",
					Members: []models.Member{
						{
							ID:     "000000000000000000000000",
							Role:   authorizer.RoleOwner,
							Status: models.MemberStatusAccepted,
						},
						{
							ID:     "000000000000000000000001",
							Role:   authorizer.RoleObserver,
							Status: models.MemberStatusPending,
						},
					},
				},
				err: nil,
			},
		},
		{
			description: "fails when cannot add the member",
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
								ID:     "000000000000000000000000",
								Role:   authorizer.RoleOwner,
								Status: models.MemberStatusAccepted,
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
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
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
			description: "succeeds",
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
								ID:     "000000000000000000000000",
								Role:   authorizer.RoleOwner,
								Status: models.MemberStatusAccepted,
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
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
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
								ID:     "000000000000000000000000",
								Role:   authorizer.RoleOwner,
								Status: models.MemberStatusAccepted,
							},
							{
								ID:     "000000000000000000000000",
								Role:   authorizer.RoleObserver,
								Status: models.MemberStatusAccepted,
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
							ID:     "000000000000000000000000",
							Role:   authorizer.RoleOwner,
							Status: models.MemberStatusAccepted,
						},
						{
							ID:     "000000000000000000000000",
							Role:   authorizer.RoleObserver,
							Status: models.MemberStatusAccepted,
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

func TestService_addMember(t *testing.T) {
	envMock = new(envmock.Backend)
	clockMock := new(clockmock.Clock)
	storeMock := new(storemock.Store)

	envs.DefaultBackend = envMock
	clock.DefaultBackend = clockMock

	now := time.Now()
	clockMock.On("Now").Return(now)

	cases := []struct {
		description   string
		memberID      string
		req           *requests.NamespaceAddMember
		requiredMocks func(context.Context)
		expected      error
	}{
		{
			description: "fails cannot add the member",
			memberID:    "000000000000000000000000",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				storeMock.
					On("NamespaceAddMember", ctx, "00000000-0000-4000-0000-000000000000", &models.Member{ID: "000000000000000000000000", Role: authorizer.RoleObserver, Status: models.MemberStatusAccepted, AddedAt: now, ExpiresAt: time.Time{}}).
					Return(errors.New("error")).
					Once()
			},
			expected: errors.New("error"),
		},
		{
			description: "succeeds",
			memberID:    "000000000000000000000000",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				storeMock.
					On("NamespaceAddMember", ctx, "00000000-0000-4000-0000-000000000000", &models.Member{ID: "000000000000000000000000", Role: authorizer.RoleObserver, Status: models.MemberStatusAccepted, AddedAt: now, ExpiresAt: time.Time{}}).
					Return(nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
			},
			expected: nil,
		},
		{
			description: "[cloud] fails cannot add the member",
			memberID:    "000000000000000000000000",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				storeMock.
					On("NamespaceAddMember", ctx, "00000000-0000-4000-0000-000000000000", &models.Member{ID: "000000000000000000000000", Role: authorizer.RoleObserver, Status: models.MemberStatusPending, AddedAt: now, ExpiresAt: now.Add(7 * (24 * time.Hour))}).
					Return(errors.New("error")).
					Once()
			},
			expected: errors.New("error"),
		},
		{
			description: "[cloud] fails cannot send the invite",
			memberID:    "000000000000000000000000",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				storeMock.
					On("NamespaceAddMember", ctx, "00000000-0000-4000-0000-000000000000", &models.Member{ID: "000000000000000000000000", Role: authorizer.RoleObserver, Status: models.MemberStatusPending, AddedAt: now, ExpiresAt: now.Add(7 * (24 * time.Hour))}).
					Return(nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				clientMock.
					On("InviteMember", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000000", "localhost").
					Return(errors.New("error")).
					Once()
			},
			expected: errors.New("error"),
		},
		{
			description: "[cloud] succeeds",
			memberID:    "000000000000000000000000",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				storeMock.
					On("NamespaceAddMember", ctx, "00000000-0000-4000-0000-000000000000", &models.Member{ID: "000000000000000000000000", Role: authorizer.RoleObserver, Status: models.MemberStatusPending, AddedAt: now, ExpiresAt: now.Add(7 * (24 * time.Hour))}).
					Return(nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				clientMock.
					On("InviteMember", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000000", "localhost").
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	s := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.Background()
			tc.requiredMocks(ctx)

			cb := s.addMember(tc.memberID, tc.req)
			assert.Equal(tt, tc.expected, cb(ctx))

			storeMock.AssertExpectations(tt)
			envMock.AssertExpectations(tt)
		})
	}
}

func TestService_resendMemberInvite(t *testing.T) {
	envMock = new(envmock.Backend)
	storeMock := new(storemock.Store)
	clockMock := new(clockmock.Clock)

	envs.DefaultBackend = envMock
	clock.DefaultBackend = clockMock

	now := time.Now()
	clockMock.On("Now").Return(now)

	cases := []struct {
		description   string
		memberID      string
		req           *requests.NamespaceAddMember
		requiredMocks func(context.Context)
		expected      error
	}{
		{
			description: "fails cannot update the member",
			memberID:    "000000000000000000000000",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				expiresAt := now.Add(7 * (24 * time.Hour))
				storeMock.
					On("NamespaceUpdateMember", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000000", &models.MemberChanges{Role: authorizer.RoleObserver, ExpiresAt: &expiresAt}).
					Return(errors.New("error")).
					Once()
			},
			expected: errors.New("error"),
		},
		{
			description: "fails when cannot send the invite",
			memberID:    "000000000000000000000000",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				expiresAt := now.Add(7 * (24 * time.Hour))
				storeMock.
					On("NamespaceUpdateMember", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000000", &models.MemberChanges{Role: authorizer.RoleObserver, ExpiresAt: &expiresAt}).
					Return(nil).
					Once()
				clientMock.
					On("InviteMember", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000000", "localhost").
					Return(errors.New("error")).
					Once()
			},
			expected: errors.New("error"),
		},
		{
			description: "succeeds",
			memberID:    "000000000000000000000000",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				expiresAt := now.Add(7 * (24 * time.Hour))
				storeMock.
					On("NamespaceUpdateMember", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000000", &models.MemberChanges{Role: authorizer.RoleObserver, ExpiresAt: &expiresAt}).
					Return(nil).
					Once()
				clientMock.
					On("InviteMember", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000000", "localhost").
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	s := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.Background()
			tc.requiredMocks(ctx)

			cb := s.resendMemberInvite(tc.memberID, tc.req)
			assert.Equal(tt, tc.expected, cb(ctx))

			envMock.AssertExpectations(tt)
			storeMock.AssertExpectations(tt)
			clockMock.AssertExpectations(tt)
		})
	}
}

func TestUpdateNamespaceMember(t *testing.T) {
	storeMock := new(storemock.Store)

	cases := []struct {
		description   string
		req           *requests.NamespaceUpdateMember
		requiredMocks func(context.Context)
		expected      error
	}{
		{
			description: "fails when the namespace was not found",
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
			description: "fails when the active member was not found",
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
			description: "fails when the active member is not on the namespace",
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
			description: "fails when the passive member is not on the namespace",
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
			description: "fails when the passive role's is owner",
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
			description: "fails when the active member's role cannot act over passive member's role",
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
			description: "fails when cannot update the member",
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
					On("NamespaceUpdateMember", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000001", &models.MemberChanges{Role: authorizer.RoleAdministrator}).
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

func TestRemoveNamespaceMember(t *testing.T) {
	type Expected struct {
		namespace *models.Namespace
		err       error
	}

	storeMock := new(storemock.Store)

	cases := []struct {
		description   string
		req           *requests.NamespaceRemoveMember
		requiredMocks func(context.Context)
		expected      Expected
	}{
		{
			description: "fails when the namespace was not found",
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
			description: "fails when the active member was not found",
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
			description: "fails when the active member is not on the namespace",
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
			description: "fails when the passive member is not on the namespace",
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
			description: "fails when the active member's role cannot act over passive member's role",
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
			description: "fails when cannot remove the member",
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
					On("NamespaceRemoveMember", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000001").
					Return(errors.New("error")).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       errors.New("error"),
			},
		},
		{
			description: "succeeds",
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
					On("NamespaceRemoveMember", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000001").
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
					On("NamespaceRemoveMember", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000000").
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
					On("NamespaceRemoveMember", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000000").
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
					On("NamespaceRemoveMember", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000000").
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
