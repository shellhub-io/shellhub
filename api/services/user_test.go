package services

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/validator"
	"github.com/stretchr/testify/assert"
)

func TestUpdateDataUser(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.Background()

	type Expected struct {
		fields []string
		err    error
	}

	cases := []struct {
		description   string
		id            string
		data          models.UserData
		requiredMocks func()
		expected      Expected
	}{
		{
			description:   "Fail when user data is invalid",
			id:            "1",
			requiredMocks: func() {},
			expected: Expected{
				fields: nil,
				err:    NewErrUserInvalid(nil, validator.ErrStructureInvalid),
			},
		},
		{
			description: "Fail when user is not found",
			id:          "1",
			data: models.UserData{
				Name:     "test",
				Username: "test",
				Email:    "test@shellhub.io",
			},
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, "1", false).Return(nil, 0, NewErrUserNotFound("1", nil)).Once()
			},
			expected: Expected{
				fields: nil,
				err:    NewErrUserNotFound("1", nil),
			},
		},
		{
			description: "Fail when username already exists",
			id:          "1",
			data: models.UserData{
				Name:     "test",
				Username: "new",
				Email:    "test@test.com",
			},
			requiredMocks: func() {
				user := &models.User{
					ID: "1",
					UserData: models.UserData{
						Name:     "test",
						Username: "test",
						Email:    "test@test.com",
					},
				}
				exist := &models.User{
					ID: "2",
					UserData: models.UserData{
						Username: "new",
					},
				}

				mock.On("UserGetByID", ctx, "1", false).Return(user, 1, nil).Once()
				mock.On("UserGetByUsername", ctx, "new").Return(exist, nil).Once()
				mock.On("UserGetByEmail", ctx, "test@test.com").Return(nil, nil).Once()
			},
			expected: Expected{
				fields: []string{"username"},
				err:    NewErrUserDuplicated([]string{"username"}, nil),
			},
		},
		{
			description: "Fail when email already exists",
			id:          "1",
			data: models.UserData{
				Name:     "test",
				Username: "test",
				Email:    "new@test.com",
			},
			requiredMocks: func() {
				user := &models.User{
					ID: "1",
					UserData: models.UserData{
						Email: "test@test.com",
					},
				}
				exist := &models.User{
					ID: "2",
					UserData: models.UserData{
						Email: "new@test.com",
					},
				}

				mock.On("UserGetByID", ctx, "1", false).Return(user, 1, nil).Once()
				mock.On("UserGetByUsername", ctx, "test").Return(nil, nil).Once()
				mock.On("UserGetByEmail", ctx, "new@test.com").Return(exist, nil).Once()
			},
			expected: Expected{
				fields: []string{"email"},
				err:    NewErrUserDuplicated([]string{"email"}, nil),
			},
		},
		{
			description: "Fail when username and email already exists",
			id:          "1",
			data: models.UserData{
				Name:     "test",
				Username: "new",
				Email:    "new@test.com",
			},
			requiredMocks: func() {
				user := &models.User{
					ID: "1",
					UserData: models.UserData{
						Username: "test",
						Email:    "test@test.com",
					},
				}
				exist := &models.User{
					ID: "2",
					UserData: models.UserData{
						Username: "new",
						Email:    "new@test.com",
					},
				}

				mock.On("UserGetByID", ctx, "1", false).Return(user, 1, nil).Once()
				mock.On("UserGetByUsername", ctx, "new").Return(exist, nil).Once()
				mock.On("UserGetByEmail", ctx, "new@test.com").Return(exist, nil).Once()
			},
			expected: Expected{
				fields: []string{"username", "email"},
				err:    NewErrUserDuplicated([]string{"username", "email"}, nil),
			},
		},
		{
			description: "Fail when could not update user",
			id:          "1",
			data: models.UserData{
				Name:     "test",
				Username: "new",
				Email:    "new@test.com",
			},
			requiredMocks: func() {
				user := &models.User{
					ID: "1",
					UserData: models.UserData{
						Username: "test",
						Email:    "test@test.com",
					},
				}

				data := models.User{
					UserData: models.UserData{
						Name:     "test",
						Username: "new",
						Email:    "new@test.com",
					},
				}

				mock.On("UserGetByID", ctx, "1", false).Return(user, 1, nil).Once()
				mock.On("UserGetByUsername", ctx, "new").Return(nil, nil).Once()
				mock.On("UserGetByEmail", ctx, "new@test.com").Return(nil, nil).Once()
				mock.On("UserUpdateData", ctx, "1", data).Return(errors.New("error", "", 0)).Once()
			},
			expected: Expected{
				fields: nil,
				err:    errors.New("error", "", 0),
			},
		},
		{
			description: "Success to update user",
			id:          "1",
			data: models.UserData{
				Name:     "test",
				Username: "new",
				Email:    "new@test.com",
			},
			requiredMocks: func() {
				user := &models.User{
					ID: "1",
					UserData: models.UserData{
						Username: "test",
						Email:    "test@test.com",
					},
				}

				data := models.User{
					UserData: models.UserData{
						Name:     "test",
						Username: "new",
						Email:    "new@test.com",
					},
				}

				mock.On("UserGetByID", ctx, "1", false).Return(user, 1, nil).Once()
				mock.On("UserGetByUsername", ctx, "new").Return(nil, nil).Once()
				mock.On("UserGetByEmail", ctx, "new@test.com").Return(nil, nil).Once()
				mock.On("UserUpdateData", ctx, "1", data).Return(nil).Once()
			},
			expected: Expected{
				fields: nil,
				err:    nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			services := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
			fields, err := services.UpdateDataUser(ctx, tc.id, tc.data)
			assert.Equal(t, tc.expected, Expected{fields, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestUpdatePasswordUser(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.Background()

	cases := []struct {
		description     string
		id              string
		currentPassword string
		newPassword     string
		requiredMocks   func()
		expected        error
	}{
		{
			description: "Fail when user is not found",
			id:          "1",
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, "1", false).Return(nil, 0, errors.New("error", "", 0)).Once()
			},
			expected: NewErrUserNotFound("1", errors.New("error", "", 0)),
		},
		{
			description:     "Fail when the current password doesn't match with user's password",
			id:              "1",
			currentPassword: "password",
			newPassword:     "newPassword",
			requiredMocks: func() {
				user := &models.User{
					UserPassword: models.NewUserPassword("passwordNoMatch"),
				}

				mock.On("UserGetByID", ctx, "1", false).Return(user, 1, nil).Once()
			},
			expected: NewErrUserPasswordNotMatch(nil),
		},
		{
			description:     "Fail to update user's password",
			id:              "1",
			currentPassword: "password",
			newPassword:     "newPassword",
			requiredMocks: func() {
				user := &models.User{
					UserPassword: models.NewUserPassword("password"),
				}

				password := models.NewUserPassword("newPassword")

				mock.On("UserGetByID", ctx, "1", false).Return(user, 1, nil).Once()
				mock.On("UserUpdatePassword", ctx, password.HashedPassword, "1").Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			services := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
			err := services.UpdatePasswordUser(ctx, tc.id, tc.currentPassword, tc.newPassword)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}
