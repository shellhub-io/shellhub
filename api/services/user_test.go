package services

import (
	"context"
	"testing"

	storecache "github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/validator"
	"github.com/stretchr/testify/assert"
)

func TestUpdateDataUser(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()

	Err := errors.New("conflict", "", 0)

	user1 := &models.User{UserData: models.UserData{Name: "name", Email: "user1@email.com", Username: "username1"}, UserPassword: models.UserPassword{Password: "hash1"}, ID: "id1"}

	user2 := &models.User{UserData: models.UserData{Name: "name", Email: "user2@email.com", Username: "username2"}, UserPassword: models.UserPassword{Password: "hash2"}, ID: "id2"}

	type Expected struct {
		fields []string
		err    error
	}

	tests := []struct {
		description   string
		user          *models.User
		updateUser    *models.User
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "Fails to find the user by the ID",
			user:        user1,
			updateUser:  &models.User{UserData: models.UserData{Name: "name", Email: "user1@email2.com", Username: user2.Username}, ID: "id1"},
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, user1.ID, false).Return(nil, 0, Err).Once()
			},
			expected: Expected{nil, Err},
		},
		{
			description: "Fails conflict username",
			user:        &models.User{UserData: models.UserData{Name: "name", Email: "user1@email.com", Username: "username1"}, UserPassword: models.UserPassword{Password: "hash1"}, ID: "id1"},
			updateUser:  &models.User{UserData: models.UserData{Name: "name", Email: "user1@email2.com", Username: user2.Username}, ID: "id1"},
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
				mock.On("UserGetByUsername", ctx, user1.Username).Return(user2, nil).Once()
				mock.On("UserGetByEmail", ctx, user1.Email).Return(user1, nil).Once()
			},
			expected: Expected{[]string{"username"}, ErrConflict},
		},
		{
			description: "Fails conflict email and username",
			user:        user1,
			updateUser:  &models.User{UserData: models.UserData{Name: "name", Email: "user1@email2.com", Username: user2.Username}, ID: "id1"},
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
				mock.On("UserGetByUsername", ctx, user1.Username).Return(user2, nil).Once()
				mock.On("UserGetByEmail", ctx, user1.Email).Return(user2, nil).Once()
			},
			expected: Expected{[]string{"username", "email"}, ErrConflict},
		},
		{
			description: "Fails invalid username",
			user:        &models.User{UserData: models.UserData{Name: "newname", Email: "user1@email2.com", Username: "invalid_name"}, ID: "id1"},
			updateUser:  &models.User{UserData: models.UserData{Name: "newname", Email: "user1@email2.com", Username: "invalid_name"}, ID: "id1"},
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
			},
			expected: Expected{[]string{"username"}, ErrBadRequest},
		},
		{
			description: "Fails invalid email",
			user:        &models.User{UserData: models.UserData{Name: "newname", Email: "invalid.email", Username: "newusername"}, ID: "id1"},
			updateUser:  &models.User{UserData: models.UserData{Name: "newname", Email: "invalid.email", Username: "newusername"}, ID: "id1"},
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
			},
			expected: Expected{[]string{"email"}, ErrBadRequest},
		},
		{
			description: "Fails invalid email and username",
			user:        &models.User{UserData: models.UserData{Name: "newname", Email: "invalid.email", Username: "us"}, ID: "id1"},
			updateUser:  &models.User{UserData: models.UserData{Name: "newname", Email: "invalid.email", Username: "us"}, ID: "id1"},
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
			},
			expected: Expected{[]string{"email", "username"}, ErrBadRequest},
		},
		{
			description: "Fails empty username",
			user:        &models.User{UserData: models.UserData{Name: "", Email: "new@email.com", Username: "newusername"}, ID: "id1"},
			updateUser:  &models.User{UserData: models.UserData{Name: "", Email: "new@email.com", Username: "newusername"}, ID: "id1"},
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
			},
			expected: Expected{[]string{"name"}, ErrBadRequest},
		},
		{
			description: "Fails empty email",
			user:        &models.User{UserData: models.UserData{Name: "newname", Email: "", Username: "newusername"}, ID: "id1"},
			updateUser:  &models.User{UserData: models.UserData{Name: "newname", Email: "", Username: "newusername"}, ID: "id1"},
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
			},
			expected: Expected{[]string{"email"}, ErrBadRequest},
		},
		{
			description: "Successful update user data",
			user:        user1,
			updateUser:  &models.User{UserData: models.UserData{Name: "name", Email: "user1@email2.com", Username: user2.Username}, ID: "id1"},
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
				mock.On("UserGetByUsername", ctx, user1.Username).Return(nil, Err).Once()
				mock.On("UserGetByEmail", ctx, user1.Email).Return(nil, Err).Once()
				mock.On("UserUpdateData", ctx, user1, user1.ID).Return(nil).Once()
			},
			expected: Expected{nil, nil},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			returnedFields, err := s.UpdateDataUser(ctx, tc.user, tc.updateUser.ID)
			assert.Equal(t, tc.expected, Expected{returnedFields, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestUpdatePasswordUser(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()

	type updatePassword struct {
		currentPassword string
		newPassword     string
	}

	cases := []struct {
		name          string
		data          updatePassword
		id            string
		requiredMocks func()
		tenantID      string
		expected      error
	}{
		{
			name:          "Fail when current password is invalid",
			data:          updatePassword{currentPassword: "1234", newPassword: "1234567"},
			id:            "1",
			requiredMocks: func() {},
			expected:      ErrBadRequest,
		},
		{
			name:          "Fail when new password is invalid",
			data:          updatePassword{currentPassword: "123456", newPassword: "123"},
			id:            "1",
			requiredMocks: func() {},
			expected:      ErrBadRequest,
		},
		{
			name:          "Fail when current and new password are equals",
			data:          updatePassword{currentPassword: "123456", newPassword: "123456"},
			id:            "1",
			requiredMocks: func() {},
			expected:      ErrBadRequest,
		},
		{
			name: "Fails when ID is not valid",
			data: updatePassword{currentPassword: "123456", newPassword: "123567"},
			id:   "2",
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, "2", false).Return(nil, 0, errors.New("error", "", 0))
			},
			expected: ErrUnauthorized,
		},
		{
			name: "Fails when user's password and current password is not equal",
			data: updatePassword{currentPassword: "123456", newPassword: "123567"},
			id:   "1",
			requiredMocks: func() {
				user := &models.User{UserData: models.UserData{Name: "name", Email: "user1@email.com", Username: "username1"}, UserPassword: models.UserPassword{Password: validator.HashPassword("password")}, ID: "1"}
				mock.On("UserGetByID", ctx, "1", false).Return(user, 0, nil).Once()
			},
			expected: ErrUnauthorized,
		},
		{
			name: "Success to update user's password",
			data: updatePassword{currentPassword: "password", newPassword: "newpassword"},
			id:   "1",
			requiredMocks: func() {
				user := &models.User{UserData: models.UserData{Name: "name", Email: "user1@email.com", Username: "username1"}, UserPassword: models.UserPassword{Password: validator.HashPassword("password")}, ID: "1"}
				mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
				mock.On("UserUpdatePassword", ctx, validator.HashPassword("newpassword"), user.ID).Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			test := tc

			test.requiredMocks()
			err := s.UpdatePasswordUser(ctx, test.data.currentPassword, test.data.newPassword, test.id)
			assert.Equal(t, err, test.expected)
		})
	}

	mock.AssertExpectations(t)
}
