package services

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestDelUser(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	Err := errors.New("error")
	userNotFound := &models.User{
		UserData: models.UserData{
			Name:     "userNotFound",
			Email:    "userNotFound@userNotFound.com",
			Username: "usernameNotFound",
		},
	}
	user := &models.User{
		ID: "userID",
		UserData: models.UserData{
			Name:     "user",
			Email:    "user@user.com",
			Username: "username",
		},
	}
	namespaceOwned := []*models.Namespace{
		{
			Name:     "namespace1",
			Owner:    user.ID,
			TenantID: "tenantID1",
			Members:  []models.Member{{ID: user.ID, Role: "owner"}},
			Settings: &models.NamespaceSettings{
				SessionRecord: true,
			},
			CreatedAt: clock.Now(),
		},
		{
			Name:     "namespace2",
			Owner:    user.ID,
			TenantID: "tenantID2",
			Members:  []models.Member{{ID: user.ID, Role: "owner"}},
			Settings: &models.NamespaceSettings{
				SessionRecord: true,
			},
			CreatedAt: clock.Now(),
		},
	}
	namespaceMember := []*models.Namespace{
		{
			Name:     "namespace3",
			Owner:    "ownerID",
			TenantID: "tenantID3",
			Members:  []models.Member{{ID: "ownerID", Role: guard.RoleObserver}, {ID: user.ID, Role: guard.RoleObserver}},
			Settings: &models.NamespaceSettings{
				SessionRecord: true,
			},
			CreatedAt: clock.Now(),
		},
		{
			Name:     "namespace1",
			Owner:    "ownerID",
			TenantID: "tenantID1",
			Members:  []models.Member{{ID: "ownerID", Role: guard.RoleObserver}, {ID: user.ID, Role: guard.RoleObserver}},
			Settings: &models.NamespaceSettings{
				SessionRecord: true,
			},
			CreatedAt: clock.Now(),
		},
	}

	tests := []struct {
		description   string
		username      string
		requiredMocks func()
		expected      error
	}{
		{
			description: "Fails to find the user",
			username:    userNotFound.Username,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, userNotFound.Username).Return(nil, Err).Once()
			},
			expected: ErrUserNotFound,
		},
		{
			description: "Fail to delete the user and associated data",
			username:    user.Username,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, user.Username).Return(user, nil).Once()
				mock.On("UserDetachInfo", ctx, user.ID).Return(nil, Err).Once()
			},
			expected: ErrNamespaceNotFound,
		},
		{
			description: "Successfully delete the user and associated data",
			username:    user.Username,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, user.Username).Return(user, nil).Once()
				mock.On("UserDetachInfo", ctx, user.ID).Return(map[string][]*models.Namespace{
					"owner":  namespaceOwned,
					"member": namespaceMember,
				}, nil)
				for _, v := range namespaceOwned {
					mock.On("NamespaceDelete", ctx, v.TenantID).Return(nil).Once()
				}
				for _, v := range namespaceMember {
					mock.On("NamespaceRemoveMember", ctx, v.TenantID, user.ID).Return(nil, nil).Once()
				}
				mock.On("UserDelete", ctx, user.ID).Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, ts := range tests {
		test := ts
		t.Run(test.description, func(t *testing.T) {
			test.requiredMocks()
			err := s.UserDelete(ctx, test.username)
			assert.Equal(t, test.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestResetUserPassword(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	Err := errors.New("error")

	userPasswordInvalid := models.UserPassword{
		Password: "ab",
	}
	userPassword := models.UserPassword{
		Password: "password",
	}
	user := &models.User{
		ID: "userID",
		UserData: models.UserData{
			Name:     "user",
			Email:    "user@user.com",
			Username: "username",
		},
	}
	userNotFound := &models.User{
		UserData: models.UserData{
			Name:     "userNotFound",
			Email:    "userNotFound@userNotFound.com",
			Username: "usernameNotFound",
		},
	}

	tests := []struct {
		description   string
		username      string
		password      string
		requiredMocks func()
		expected      error
	}{
		{
			description: "Fails when the field is invalid",
			username:    user.Username,
			password:    userPasswordInvalid.Password,
			requiredMocks: func() {
			},
			expected: ErrUserPasswordInvalid,
		},
		{
			description: "Fails to find the user",
			username:    userNotFound.Username,
			password:    userPassword.Password,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, userNotFound.Username).Return(nil, Err).Once()
			},
			expected: ErrUserNotFound,
		},
		{
			description: "Fail reset the user password",
			username:    user.Username,
			password:    userPassword.Password,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, user.Username).Return(user, nil).Once()
				mock.On("UserUpdatePassword", ctx, hashPassword(userPassword.Password), user.ID).Return(Err).Once()
			},
			expected: ErrFailedUpdateUser,
		},
		{
			description: "Successfully reset the user password",
			username:    user.Username,
			password:    userPassword.Password,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, user.Username).Return(user, nil).Once()
				mock.On("UserUpdatePassword", ctx, hashPassword(userPassword.Password), user.ID).Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, ts := range tests {
		test := ts
		t.Run(test.description, func(t *testing.T) {
			test.requiredMocks()
			err := s.UserUpdate(ctx, test.username, test.password)
			assert.Equal(t, test.expected, err)
		})
	}

	mock.AssertExpectations(t)
}
