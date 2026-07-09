package services

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	storemock "github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	cachemock "github.com/shellhub-io/shellhub/pkg/cache/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	uuidmock "github.com/shellhub-io/shellhub/pkg/uuid/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_CreateUserActivationToken(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	cacheMock := cachemock.NewMockCache(t)

	uuidMock := uuidmock.NewMockUUID(t)
	prevUUID := uuid.DefaultBackend
	t.Cleanup(func() { uuid.DefaultBackend = prevUUID })
	uuid.DefaultBackend = uuidMock

	const actorID = "000000000000000000000001"
	const targetID = "000000000000000000000000"

	cases := []struct {
		description   string
		req           *requests.CreateUserActivation
		requiredMocks func(ctx context.Context)
		expectedErr   error
	}{
		{
			description: "fails when a non-admin mints for an unapproved account",
			req:         &requests.CreateUserActivation{UserParam: requests.UserParam{ID: targetID}, UserID: actorID},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, actorID).
					Return(&models.User{ID: actorID, Admin: false}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, targetID).
					Return(&models.User{ID: targetID, Status: models.UserStatusNotConfirmed, AwaitingApproval: true}, nil).
					Once()
			},
			expectedErr: NewErrAuthForbidden(),
		},
		{
			description: "fails when a non-admin mints for an approved account they don't manage",
			req:         &requests.CreateUserActivation{UserParam: requests.UserParam{ID: targetID}, UserID: actorID},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, actorID).
					Return(&models.User{ID: actorID, Admin: false}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, targetID).
					Return(&models.User{ID: targetID, Status: models.UserStatusNotConfirmed, AwaitingApproval: false}, nil).
					Once()
				queryOptionsMock := new(storemock.MockQueryOptions)
				storeMock.On("Options").Return(queryOptionsMock).Once()
				queryOptionsMock.On("WithMember", targetID).Return(nil).Once()
				storeMock.
					On("NamespaceList", ctx, mock.Anything).
					Return([]models.Namespace{{TenantID: "00000000-0000-4000-0000-000000000000"}}, 1, nil).
					Once()
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Members:  []models.Member{{ID: actorID, Role: authorizer.RoleObserver}},
					}, nil).
					Once()
			},
			expectedErr: NewErrAuthForbidden(),
		},
		{
			description: "lets a namespace admin mint for an approved account they manage",
			req:         &requests.CreateUserActivation{UserParam: requests.UserParam{ID: targetID}, UserID: actorID},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, actorID).
					Return(&models.User{ID: actorID, Admin: false}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, targetID).
					Return(&models.User{ID: targetID, Status: models.UserStatusNotConfirmed, AwaitingApproval: false}, nil).
					Once()
				queryOptionsMock := new(storemock.MockQueryOptions)
				storeMock.On("Options").Return(queryOptionsMock).Once()
				queryOptionsMock.On("WithMember", targetID).Return(nil).Once()
				storeMock.
					On("NamespaceList", ctx, mock.Anything).
					Return([]models.Namespace{{TenantID: "00000000-0000-4000-0000-000000000000"}}, 1, nil).
					Once()
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Members:  []models.Member{{ID: actorID, Role: authorizer.RoleAdministrator}},
					}, nil).
					Once()
				uuidMock.On("Generate").Return("00000000-0000-0000-0000-000000000000").Once()
				cacheMock.
					On("Set", ctx, "recover-password={"+targetID+"}", "00000000-0000-0000-0000-000000000000", userActivationTokenTTL).
					Return(nil).
					Once()
				clockMock.On("Now").Return(now)
			},
			expectedErr: nil,
		},
		{
			description: "fails when the target user does not exist",
			req:         &requests.CreateUserActivation{UserParam: requests.UserParam{ID: targetID}, UserID: actorID},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, actorID).
					Return(&models.User{ID: actorID, Admin: true}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, targetID).
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expectedErr: NewErrUserNotFound(targetID, store.ErrNoDocuments),
		},
		{
			description: "fails when an admin mints for an already-activated account",
			req:         &requests.CreateUserActivation{UserParam: requests.UserParam{ID: targetID}, UserID: actorID},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, actorID).
					Return(&models.User{ID: actorID, Admin: true}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, targetID).
					Return(&models.User{ID: targetID, Status: models.UserStatusConfirmed}, nil).
					Once()
			},
			expectedErr: NewErrAuthForbidden(),
		},
		{
			description: "mints a token when an admin requests it for an existing user",
			req:         &requests.CreateUserActivation{UserParam: requests.UserParam{ID: targetID}, UserID: actorID},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, actorID).
					Return(&models.User{ID: actorID, Admin: true}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, targetID).
					Return(&models.User{ID: targetID, Status: models.UserStatusNotConfirmed}, nil).
					Once()
				uuidMock.On("Generate").Return("00000000-0000-0000-0000-000000000000").Once()
				cacheMock.
					On("Set", ctx, "recover-password={"+targetID+"}", "00000000-0000-0000-0000-000000000000", userActivationTokenTTL).
					Return(nil).
					Once()
				clockMock.On("Now").Return(now)
			},
			expectedErr: nil,
		},
	}

	s := NewService(store.Store(storeMock), privateKey, publicKey, cacheMock, clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.TODO()
			tc.requiredMocks(ctx)

			token, expiresAt, err := s.CreateUserActivationToken(ctx, tc.req)
			assert.Equal(t, tc.expectedErr, err)

			if tc.expectedErr == nil {
				assert.NotEmpty(t, token)
				assert.False(t, expiresAt.IsZero())
			}
		})
	}

	storeMock.AssertExpectations(t)
	cacheMock.AssertExpectations(t)
}

func TestService_ActivateUser(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	cacheMock := cachemock.NewMockCache(t)

	cases := []struct {
		description   string
		req           *requests.ActivateUser
		requiredMocks func(ctx context.Context)
		expectedErr   error
	}{
		{
			description: "fails when the user does not exist",
			req:         &requests.ActivateUser{UserParam: requests.UserParam{ID: "000000000000000000000000"}, Token: "token", Password: "newpassword"},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expectedErr: NewErrUserNotFound("000000000000000000000000", store.ErrNoDocuments),
		},
		{
			description: "fails when the account is already activated",
			req:         &requests.ActivateUser{UserParam: requests.UserParam{ID: "000000000000000000000000"}, Token: "stored-token", Password: "newpassword"},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{ID: "000000000000000000000000", Status: models.UserStatusConfirmed}, nil).
					Once()
			},
			expectedErr: NewErrAuthForbidden(),
		},
		{
			description: "fails when the token does not match",
			req:         &requests.ActivateUser{UserParam: requests.UserParam{ID: "000000000000000000000000"}, Token: "wrong-token", Password: "newpassword"},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{ID: "000000000000000000000000", Status: models.UserStatusNotConfirmed}, nil).
					Once()
				cacheMock.
					On("Get", ctx, "recover-password={000000000000000000000000}", mock.Anything).
					Run(func(args mock.Arguments) { *args.Get(2).(*string) = "stored-token" }).
					Return(nil).
					Once()
			},
			expectedErr: NewErrAuthUnathorized(nil),
		},
		{
			description: "sets the password and confirms the account with a valid token",
			req:         &requests.ActivateUser{UserParam: requests.UserParam{ID: "000000000000000000000000"}, Token: "stored-token", Password: "newpassword"},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{ID: "000000000000000000000000", Status: models.UserStatusNotConfirmed, AwaitingApproval: true}, nil).
					Once()
				cacheMock.
					On("Get", ctx, "recover-password={000000000000000000000000}", mock.Anything).
					Run(func(args mock.Arguments) { *args.Get(2).(*string) = "stored-token" }).
					Return(nil).
					Once()
				hashMock.
					On("Do", "newpassword").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yv", nil).
					Once()
				storeMock.
					On("UserUpdate", ctx, mock.MatchedBy(func(u *models.User) bool {
						return u.ID == "000000000000000000000000" &&
							u.Status == models.UserStatusConfirmed &&
							!u.AwaitingApproval &&
							u.Password.Hash != ""
					})).
					Return(nil).
					Once()
				cacheMock.
					On("Delete", ctx, "recover-password={000000000000000000000000}").
					Return(nil).
					Once()
			},
			expectedErr: nil,
		},
	}

	s := NewService(store.Store(storeMock), privateKey, publicKey, cacheMock, clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.TODO()
			tc.requiredMocks(ctx)

			err := s.ActivateUser(ctx, tc.req)
			assert.Equal(t, tc.expectedErr, err)
		})
	}

	storeMock.AssertExpectations(t)
	cacheMock.AssertExpectations(t)
}
