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
	mock := new(mocks.Store)
	ctx := context.TODO()

	cases := []struct {
		description   string
		username      string
		requiredMocks func()
		expected      error
	}{
		{
			description: "fails when could not find a user",
			username:    "john_doe",
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, "john_doe").Return(nil, errors.New("error")).Once()
			},
			expected: ErrUserNotFound,
		},
		{
			description: "fails to delete the user and associated data when namespace not found",
			username:    "john_doe",
			requiredMocks: func() {
				user := &models.User{
					ID: "507f191e810c19729de860ea",
					UserData: models.UserData{
						Name:     "John Doe",
						Email:    "john.doe@test.com",
						Username: "john_doe",
					},
				}
				mock.On("UserGetByUsername", ctx, "john_doe").Return(user, nil).Once()
				mock.On("UserDetachInfo", ctx, "507f191e810c19729de860ea").Return(nil, errors.New("error")).Once()
			},
			expected: ErrNamespaceNotFound,
		},
		{
			description: "successfully delete the user and associated data",
			username:    "john_doe",
			requiredMocks: func() {
				user := &models.User{
					ID: "507f191e810c19729de860ea",
					UserData: models.UserData{
						Name:     "John Doe",
						Email:    "john.doe@test.com",
						Username: "john_doe",
					},
				}
				mock.On("UserGetByUsername", ctx, "john_doe").Return(user, nil).Once()

				namespaceOwned := []*models.Namespace{
					{
						Name:     "namespace1",
						Owner:    "507f191e810c19729de860ea",
						TenantID: "10000000-0000-0000-0000-000000000000",
						Members:  []models.Member{{ID: "507f191e810c19729de860ea", Role: "owner"}},
						Settings: &models.NamespaceSettings{
							SessionRecord: true,
						},
						CreatedAt: clock.Now(),
					},
					{
						Name:     "namespace2",
						Owner:    "507f191e810c19729de860ea",
						TenantID: "20000000-0000-0000-0000-000000000000",
						Members:  []models.Member{{ID: "507f191e810c19729de860ea", Role: "owner"}},
						Settings: &models.NamespaceSettings{
							SessionRecord: true,
						},
						CreatedAt: clock.Now(),
					},
				}
				namespaceMember := []*models.Namespace{
					{
						Name:     "namespace3",
						Owner:    "507f191e810c19729de86000",
						TenantID: "30000000-0000-0000-0000-000000000000",
						Members: []models.Member{
							{ID: "507f191e810c19729de86000", Role: guard.RoleObserver},
							{ID: "507f191e810c19729de860ea", Role: guard.RoleObserver},
						},
						Settings: &models.NamespaceSettings{
							SessionRecord: true,
						},
						CreatedAt: clock.Now(),
					},
					{
						Name:     "namespace1",
						Owner:    "507f191e810c19729de86000",
						TenantID: "tenantID1",
						Members: []models.Member{
							{ID: "507f191e810c19729de86000", Role: guard.RoleObserver},
							{ID: "507f191e810c19729de860ea", Role: guard.RoleObserver},
						},
						Settings: &models.NamespaceSettings{
							SessionRecord: true,
						},
						CreatedAt: clock.Now(),
					},
				}

				mock.On("UserDetachInfo", ctx, "507f191e810c19729de860ea").Return(map[string][]*models.Namespace{
					"owner":  namespaceOwned,
					"member": namespaceMember,
				}, nil)

				for _, v := range namespaceOwned {
					mock.On("NamespaceDelete", ctx, v.TenantID).Return(nil).Once()
				}
				for _, v := range namespaceMember {
					mock.On("NamespaceRemoveMember", ctx, v.TenantID, "507f191e810c19729de860ea").Return(nil, nil).Once()
				}

				mock.On("UserDelete", ctx, "507f191e810c19729de860ea").Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(mock))
			err := service.UserDelete(ctx, tc.username)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestResetUserPassword(t *testing.T) {
	mock := new(mocks.Store)
	ctx := context.TODO()

	cases := []struct {
		description   string
		username      string
		password      string
		requiredMocks func()
		expected      error
	}{
		{
			description: "fails when the password is invalid",
			username:    "john_doe",
			password:    "ab",
			requiredMocks: func() {
			},
			expected: ErrUserPasswordInvalid,
		},
		{
			description: "fails when could not find a user",
			username:    "john_doe",
			password:    "password",
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, "john_doe").Return(nil, errors.New("error")).Once()
			},
			expected: ErrUserNotFound,
		},
		{
			description: "Fail reset the user password",
			username:    "john_doe",
			password:    "password",
			requiredMocks: func() {
				user := &models.User{
					ID: "507f191e810c19729de860ea",
					UserData: models.UserData{
						Name:     "John Doe",
						Email:    "john.doe@test.com",
						Username: "john_doe",
					},
				}
				mock.On("UserGetByUsername", ctx, "john_doe").Return(user, nil).Once()
				mock.On("UserUpdatePassword", ctx, hashPassword("password"), "507f191e810c19729de860ea").Return(errors.New("error")).Once()
			},
			expected: ErrFailedUpdateUser,
		},
		{
			description: "Successfully reset the user password",
			username:    "john_doe",
			password:    "password",
			requiredMocks: func() {
				user := &models.User{
					ID: "507f191e810c19729de860ea",
					UserData: models.UserData{
						Name:     "John Doe",
						Email:    "john.doe@test.com",
						Username: "john_doe",
					},
				}
				mock.On("UserGetByUsername", ctx, "john_doe").Return(user, nil).Once()
				mock.On("UserUpdatePassword", ctx, hashPassword("password"), "507f191e810c19729de860ea").Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()
			service := NewService(store.Store(mock))
			err := service.UserUpdate(ctx, tc.username, tc.password)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}
