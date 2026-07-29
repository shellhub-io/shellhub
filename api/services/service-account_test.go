package services

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	storemock "github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
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
			// Guards against deleting a real person through the service-account endpoint.
			description: "fails when the user is a human, not a service account",
			requireMocks: func(storeMock *storemock.MockStore) {
				storeMock.On("UserResolve", ctx, store.UserIDResolver, saID).
					Return(&models.User{ID: saID, Type: models.UserTypeHuman}, nil).Once()
			},
			expectedErr: true,
		},
		{
			// Guards against deleting a service account that lives in another namespace.
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

			service := NewService(storeMock, privateKey, publicKey, nil, clientMock)

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
