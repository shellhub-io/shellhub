package services

import (
	"context"
	"errors"
	"testing"

	storecache "github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestUpdateDataUser(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()

	Err := errors.New("conflict")

	user1 := &models.User{UserData: models.UserData{Name: "name", Email: "user1@email.com", Username: "username1"}, UserPassword: models.UserPassword{Password: "hash1"}, ID: "id1"}

	user2 := &models.User{UserData: models.UserData{Name: "name", Email: "user2@email.com", Username: "username2"}, UserPassword: models.UserPassword{Password: "hash2"}, ID: "id2"}

	updateUser1 := &models.User{UserData: models.UserData{Name: "name", Email: "user1@email2.com", Username: user2.Username}, ID: "id1"}

	updateInvalidUsername := &models.User{UserData: models.UserData{Name: "newname", Email: "user1@email2.com", Username: "invalid_name"}, ID: "id1"}

	updateInvalidEmail := &models.User{UserData: models.UserData{Name: "newname", Email: "invalid.email", Username: "newusername"}, ID: "id1"}

	updateInvalidUsernameEmail := &models.User{UserData: models.UserData{Name: "newname", Email: "invalid.email", Username: "us"}, ID: "id1"}

	updateEmptyUsername := &models.User{UserData: models.UserData{Name: "", Email: "new@email.com", Username: "newusername"}, ID: "id1"}

	updateEmptyEmail := &models.User{UserData: models.UserData{Name: "newname", Email: "", Username: "newusername"}, ID: "id1"}

	conflictedUsername := []string{"username"}

	conflictedUserEmail := []string{"username", "email"}

	invalidUsername := []string{"username"}

	invalidEmail := []string{"email"}

	invalidUsernameEmail := []string{"email", "username"}

	emptyUsername := []string{"name"}

	emptyEmail := []string{"email"}

	type Expected struct {
		fields []string
		err    error
	}

	tests := []struct {
		description                string
		user1                      *models.User
		user2                      *models.User
		updateUser                 *models.User
		updateInvalidUsername      *models.User
		updateInvalidEmail         *models.User
		updateInvalidUsernameEmail *models.User
		updateEmptyUsername        *models.User
		updateEmptyEmail           *models.User
		requiredMocks              func()
		expected                   Expected
	}{
		{
			description: "Fails to find the user by the ID",
			user1:       user1,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, user1.ID, false).Return(nil, 0, Err).Once()
			},
			expected: Expected{nil, Err},
		},
		{
			description: "Fails conflict username",
			user1:       user1,
			user2:       user2,
			updateUser:  updateUser1,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, updateUser1.ID, false).Return(user1, 0, nil).Once()
				mock.On("UserGetByUsername", ctx, updateUser1.Username).Return(user2, nil).Once()
				mock.On("UserGetByEmail", ctx, updateUser1.Email).Return(nil, Err).Once()
			},
			expected: Expected{conflictedUsername, Err},
		},
		{
			description: "Fails conflict email and username",
			user1:       user1,
			user2:       user2,
			updateUser:  updateUser1,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, updateUser1.ID, false).Return(user1, 0, nil).Once()
				mock.On("UserGetByUsername", ctx, updateUser1.Username).Return(user2, nil).Once()
				mock.On("UserGetByEmail", ctx, updateUser1.Email).Return(user2, nil).Once()
			},
			expected: Expected{conflictedUserEmail, Err},
		},
		{
			description:           "Fails invalid username",
			user1:                 user1,
			updateInvalidUsername: updateInvalidUsername,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
			},
			expected: Expected{invalidUsername, ErrBadRequest},
		},
		{
			description:        "Fails invalid email",
			user1:              user1,
			updateInvalidEmail: updateInvalidEmail,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
			},
			expected: Expected{invalidEmail, ErrBadRequest},
		},
		{
			description:                "Fails invalid email and username",
			user1:                      user1,
			updateInvalidUsernameEmail: updateInvalidUsernameEmail,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
			},
			expected: Expected{invalidUsernameEmail, ErrBadRequest},
		},
		{
			description:         "Fails empty username",
			user1:               user1,
			updateEmptyUsername: updateEmptyUsername,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
			},
			expected: Expected{emptyUsername, ErrBadRequest},
		},
		{
			description:      "Fails empty email",
			user1:            user1,
			updateEmptyEmail: updateEmptyEmail,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
			},
			expected: Expected{emptyEmail, ErrBadRequest},
		},
		{
			description: "Successful update user data",
			user1:       user1,
			updateUser:  updateUser1,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
				mock.On("UserGetByUsername", ctx, updateUser1.Username).Return(nil, Err).Once()
				mock.On("UserGetByEmail", ctx, updateUser1.Email).Return(nil, Err).Once()
				mock.On("UserUpdateData", ctx, updateUser1, updateUser1.ID).Return(nil).Once()
			},
			expected: Expected{nil, nil},
		},
	}

	returnedFields := new([]string)

	for _, test := range tests {
		t.Log("PASS:  ", test.description)
		test.requiredMocks()
		switch test.description {
		case "Fails to find the user by the ID":
			*returnedFields, Err = s.UpdateDataUser(ctx, test.user1, test.user1.ID)
		case "Fails invalid username":
			*returnedFields, Err = s.UpdateDataUser(ctx, test.updateInvalidUsername, test.updateInvalidUsername.ID)
		case "Fails invalid email":
			*returnedFields, Err = s.UpdateDataUser(ctx, test.updateInvalidEmail, test.updateInvalidEmail.ID)
		case "Fails invalid email and username":
			*returnedFields, Err = s.UpdateDataUser(ctx, test.updateInvalidUsernameEmail, test.updateInvalidUsernameEmail.ID)
		case "Fails empty username":
			*returnedFields, Err = s.UpdateDataUser(ctx, test.updateEmptyUsername, test.updateEmptyUsername.ID)
		case "Fails empty email":
			*returnedFields, Err = s.UpdateDataUser(ctx, test.updateEmptyEmail, test.updateEmptyEmail.ID)
		case "Successful update user data":
			*returnedFields, Err = s.UpdateDataUser(ctx, test.updateUser, test.updateUser.ID)
		default:
			*returnedFields, Err = s.UpdateDataUser(ctx, test.updateUser, test.updateUser.ID)
		}
		assert.Equal(t, test.expected, Expected{*returnedFields, Err})
	}

	mock.AssertExpectations(t)
}

func TestUpdatePasswordUser(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()

	user1 := &models.User{UserData: models.UserData{Name: "name", Email: "user1@email.com", Username: "username1"}, UserPassword: models.UserPassword{Password: "hash1"}, ID: "id1"}

	type updatePassword struct {
		currentPassword string
		newPassword     string
		expected        error
	}

	tests := []updatePassword{
		{
			"hiadoshioasc",
			"hashnew",
			ErrUnauthorized,
		},
		{
			"pass123",
			"hashnew",
			ErrUnauthorized,
		},
		{
			"askdhkasd",
			"hashnew",
			ErrUnauthorized,
		},
		{
			"pass890",
			"hashnew",
			ErrUnauthorized,
		},
		{
			"hash1",
			"hashnew",
			nil,
		},
	}

	for _, test := range tests {
		mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
		if test.expected == nil {
			mock.On("UserUpdatePassword", ctx, test.newPassword, user1.ID).Return(nil).Once()
		}
		err := s.UpdatePasswordUser(ctx, test.currentPassword, test.newPassword, user1.ID)
		assert.Equal(t, err, test.expected)
	}

	mock.AssertExpectations(t)
}
