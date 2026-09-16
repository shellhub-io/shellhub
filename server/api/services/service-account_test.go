package services

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDeleteServiceAccount(t *testing.T) {
	ctx := context.TODO()

	const (
		tenantID = "00000000-0000-4000-0000-000000000000"
		saID     = "00000000-0000-0000-0000-00000000000a"
	)

	req := &requests.ServiceAccountDelete{
		ServiceAccountIDParam: requests.ServiceAccountIDParam{ID: saID},
		TenantID:              tenantID,
	}

	cases := []struct {
		description  string
		requireMocks func(storeMock *storemock.MockStore)
		expectedErr  bool
	}{
		{
			description: "deletes a service account that belongs to the namespace",
			requireMocks: func(storeMock *storemock.MockStore) {
				storeMock.On("UserResolve", ctx, store.UserIDResolver, saID).
					Return(&models.User{ID: saID, Type: models.UserTypeService}, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(&models.Namespace{TenantID: tenantID, Members: []models.Member{{ID: saID, Role: authorizer.RoleService}}}, nil).Once()
				storeMock.On("UserDelete", ctx, &models.User{ID: saID}).Return(nil).Once()
			},
			expectedErr: false,
		},
		{
			description: "fails when the user does not exist",
			requireMocks: func(storeMock *storemock.MockStore) {
				storeMock.On("UserResolve", ctx, store.UserIDResolver, saID).
					Return(nil, store.ErrNoDocuments).Once()
			},
			expectedErr: true,
		},
		{
			description: "fails when the user is a human, not a service account",
			requireMocks: func(storeMock *storemock.MockStore) {
				storeMock.On("UserResolve", ctx, store.UserIDResolver, saID).
					Return(&models.User{ID: saID, Type: models.UserTypeHuman}, nil).Once()
			},
			expectedErr: true,
		},
		{
			description: "fails when the service account is not a member of the namespace",
			requireMocks: func(storeMock *storemock.MockStore) {
				storeMock.On("UserResolve", ctx, store.UserIDResolver, saID).
					Return(&models.User{ID: saID, Type: models.UserTypeService}, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(&models.Namespace{TenantID: tenantID}, nil).Once()
			},
			expectedErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			storeMock := new(storemock.MockStore)
			tc.requireMocks(storeMock)

			service := NewService(storeMock, privateKey, publicKey, nil)

			err := service.DeleteServiceAccount(ctx, req)
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			storeMock.AssertExpectations(t)
		})
	}
}

func TestListServiceAccountsGivesEveryAccountAnIdentityList(t *testing.T) {
	ctx := context.TODO()

	const (
		tenantID = "00000000-0000-4000-0000-000000000000"
		withKey  = "00000000-0000-0000-0000-00000000000a"
		without  = "00000000-0000-0000-0000-00000000000b"
	)

	storeMock := new(storemock.MockStore)

	storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
		Return(&models.Namespace{TenantID: tenantID}, nil).Once()
	storeMock.On("ServiceAccountList", ctx, tenantID).
		Return([]models.ServiceAccount{{ID: withKey, Name: "deployer"}, {ID: without, Name: "auditor"}}, 2, nil).Once()
	storeMock.On("SSHIdentityList", ctx, mock.Anything).
		Return([]models.SSHIdentity{{ID: "identity-1", PrincipalID: withKey}}, 1, nil).Once()

	service := NewService(storeMock, privateKey, publicKey, nil)

	accounts, err := service.ListServiceAccounts(ctx, &requests.ServiceAccountList{TenantID: tenantID})
	require.NoError(t, err)
	require.Len(t, accounts, 2)

	require.Len(t, accounts[0].Identities, 1)
	require.NotNil(t, accounts[1].Identities, "an account with no identity serializes [], not null")
	require.Empty(t, accounts[1].Identities)

	storeMock.AssertExpectations(t)
}
