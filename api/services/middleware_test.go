package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"

	storecache "github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestCheckPermission(t *testing.T) {
	mock := &mocks.Store{}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	s := NewService(store.Store(mock), privateKey, &privateKey.PublicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()

	Err := errors.New("")

	userOwner := &models.User{
		ID: "user0ID",
	}
	userObserver := &models.User{
		ID: "user1ID",
	}
	namespace := &models.Namespace{
		Owner:    userOwner.ID,
		TenantID: "tenantID",
		Members: []models.Member{
			{
				ID:   userOwner.ID,
				Type: authorizer.MemberTypeOwner,
			},
			{
				ID:   userObserver.ID,
				Type: authorizer.MemberTypeObserver,
			},
		},
	}

	tests := []struct {
		description   string
		tenantID      string
		userID        string
		action        int
		callback      error
		requiredMocks func()
		expected      error
	}{
		{
			description: "Fail when the user was not found",
			tenantID:    namespace.TenantID,
			userID:      "userNotFound",
			action:      authorizer.Actions.Namespace.Rename,
			callback:    nil,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, "userNotFound", false).Return(nil, 0, Err).Once()
			},
			expected: ErrForbidden,
		},
		{
			description: "Fail when the namespace was not found",
			tenantID:    "namespaceNotFound",
			userID:      userOwner.ID,
			action:      authorizer.Actions.Namespace.Rename,
			callback:    nil,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, userOwner.ID, false).Return(userOwner, 0, nil).Once()
				mock.On("NamespaceGet", ctx, "namespaceNotFound").Return(nil, Err).Once()
			},
			expected: ErrForbidden,
		},
		{
			description: "Fail when the user does not have the permission to execute an action",
			tenantID:    namespace.TenantID,
			userID:      userObserver.ID,
			action:      authorizer.Actions.Namespace.Rename,
			callback:    nil,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, userObserver.ID, false).Return(userObserver, 0, nil).Once()
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
			},
			expected: ErrForbidden,
		},
		{
			description: "Success when the user has permission to execute an action",
			tenantID:    namespace.TenantID,
			userID:      userOwner.ID,
			action:      authorizer.Actions.Namespace.Rename,
			callback:    nil,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, userOwner.ID, false).Return(userOwner, 0, nil).Once()
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()
			err := s.CheckPermission(ctx, tc.tenantID, tc.userID, tc.action, func() error {
				return tc.callback
			})

			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}
