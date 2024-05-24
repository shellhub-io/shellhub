package services

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestUpdateDataUser(t *testing.T) {
	type Expected struct {
		conflicts []string
		err       error
	}

	storeMock := new(mocks.Store)

	cases := []struct {
		description   string
		userID        string
		req           *requests.UserDataUpdate
		requiredMocks func(context.Context)
		expected      Expected
	}{
		{
			description: "Fail when user is not found",
			userID:      "000000000000000000000000",
			req: &requests.UserDataUpdate{
				Name:          "John Doe",
				Username:      "john_doe",
				Email:         "john.doe@test.com",
				RecoveryEmail: "recovery@test.com",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(nil, 0, NewErrUserNotFound("000000000000000000000000", nil)).
					Once()
			},
			expected: Expected{
				conflicts: nil,
				err:       NewErrUserNotFound("000000000000000000000000", nil),
			},
		},
		{
			description: "Fail when recovery email is same as req's email",
			userID:      "000000000000000000000000",
			req: &requests.UserDataUpdate{
				Name:          "John Doe",
				Username:      "john_doe",
				Email:         "john.doe@test.com",
				RecoveryEmail: "john.doe@test.com",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(
						&models.User{
							ID: "000000000000000000000000",
							UserData: models.UserData{
								Name:          "James Smith",
								Username:      "james_smith",
								Email:         "james.smith@test.com",
								RecoveryEmail: "recover@test.com",
							},
						},
						0,
						nil,
					).
					Once()
			},
			expected: Expected{
				conflicts: []string{"email", "recovery_email"},
				err:       NewErrBadRequest(nil),
			},
		},
		{
			description: "Fail when recovery email is same as user's email",
			userID:      "000000000000000000000000",
			req: &requests.UserDataUpdate{
				Name:          "John Doe",
				Username:      "john_doe",
				Email:         "john.doe@test.com",
				RecoveryEmail: "james.smith@test.com",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(
						&models.User{
							ID: "000000000000000000000000",
							UserData: models.UserData{
								Name:          "James Smith",
								Username:      "james_smith",
								Email:         "james.smith@test.com",
								RecoveryEmail: "recover@test.com",
							},
						},
						0,
						nil,
					).
					Once()
			},
			expected: Expected{
				conflicts: []string{"email", "recovery_email"},
				err:       NewErrBadRequest(nil),
			},
		},
		{
			description: "Fail when username already exists",
			userID:      "000000000000000000000000",
			req: &requests.UserDataUpdate{
				Name:          "John Doe",
				Username:      "james_smith",
				Email:         "john.doe@test.com",
				RecoveryEmail: "recovery@test.com",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(
						&models.User{
							ID: "000000000000000000000000",
							UserData: models.UserData{
								Name:          "James Smith",
								Username:      "james_smith",
								Email:         "james.smith@test.com",
								RecoveryEmail: "recover@test.com",
							},
						},
						0,
						nil,
					).
					Once()
				storeMock.
					On("UserConflicts", ctx, &models.UserConflicts{Username: "james_smith", Email: "john.doe@test.com"}).
					Return([]string{"username"}, true, nil).
					Once()
			},
			expected: Expected{
				conflicts: []string{"username"},
				err:       NewErrUserDuplicated([]string{"username"}, nil),
			},
		},
		{
			description: "Fail when email already exists",
			userID:      "000000000000000000000000",
			req: &requests.UserDataUpdate{
				Name:          "John Doe",
				Username:      "john_doe",
				Email:         "james.smith@test.com",
				RecoveryEmail: "recovery@test.com",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(
						&models.User{
							ID: "000000000000000000000000",
							UserData: models.UserData{
								Name:          "James Smith",
								Username:      "james_smith",
								Email:         "james.smith@test.com",
								RecoveryEmail: "recover@test.com",
							},
						},
						0,
						nil,
					).
					Once()
				storeMock.
					On("UserConflicts", ctx, &models.UserConflicts{Username: "john_doe", Email: "james.smith@test.com"}).
					Return([]string{"email"}, true, nil).
					Once()
			},
			expected: Expected{
				conflicts: []string{"email"},
				err:       NewErrUserDuplicated([]string{"email"}, nil),
			},
		},
		{
			description: "Fail when could not update user",
			userID:      "000000000000000000000000",
			req: &requests.UserDataUpdate{
				Name:          "John Doe",
				Username:      "john_doe",
				Email:         "john.doe@test.com",
				RecoveryEmail: "recovery@test.com",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(
						&models.User{
							ID: "000000000000000000000000",
							UserData: models.UserData{
								Name:          "James Smith",
								Username:      "james_smith",
								Email:         "james.smith@shellhub.io",
								RecoveryEmail: "recover@test.com",
							},
						},
						0,
						nil,
					).
					Once()
				storeMock.
					On("UserConflicts", ctx, &models.UserConflicts{Username: "john_doe", Email: "john.doe@test.com"}).
					Return([]string{}, false, nil).
					Once()
				storeMock.
					On("UserUpdate", ctx, "000000000000000000000000", &models.UserChanges{
						Name:          "John Doe",
						Username:      "john_doe",
						Email:         "john.doe@test.com",
						RecoveryEmail: "recovery@test.com",
					}).
					Return(errors.New("error", "", 0)).
					Once()
			},
			expected: Expected{
				conflicts: nil,
				err:       errors.New("error", "", 0),
			},
		},
		{
			description: "Success to update user",
			userID:      "000000000000000000000000",
			req: &requests.UserDataUpdate{
				Name:          "John Doe",
				Username:      "john_doe",
				Email:         "john.doe@test.com",
				RecoveryEmail: "recovery@test.com",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(
						&models.User{
							ID: "000000000000000000000000",
							UserData: models.UserData{
								Name:          "James Smith",
								Username:      "james_smith",
								Email:         "james.smith@shellhub.io",
								RecoveryEmail: "recover@test.com",
							},
						},
						0,
						nil,
					).
					Once()
				storeMock.
					On("UserConflicts", ctx, &models.UserConflicts{Username: "john_doe", Email: "john.doe@test.com"}).
					Return([]string{}, false, nil).
					Once()
				storeMock.
					On("UserUpdate", ctx, "000000000000000000000000", &models.UserChanges{
						Name:          "John Doe",
						Username:      "john_doe",
						Email:         "john.doe@test.com",
						RecoveryEmail: "recovery@test.com",
					}).
					Return(nil).
					Once()
			},
			expected: Expected{
				conflicts: nil,
				err:       nil,
			},
		},
	}

	service := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()
			tc.requiredMocks(ctx)

			conflicts, err := service.UpdateDataUser(ctx, tc.userID, tc.req)
			assert.Equal(t, tc.expected, Expected{conflicts, err})
		})
	}

	storeMock.AssertExpectations(t)
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
			description: "fails when user is not found",
			id:          "65fde3a72c4c7507c7f53c43",
			requiredMocks: func() {
				mock.
					On("UserGetByID", ctx, "65fde3a72c4c7507c7f53c43", false).
					Return(nil, 0, errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrUserNotFound("65fde3a72c4c7507c7f53c43", errors.New("error", "", 0)),
		},
		{
			description:     "fails when the current password doesn't match with user's password",
			id:              "1",
			currentPassword: "wrong_password",
			newPassword:     "newSecret",
			requiredMocks: func() {
				user := &models.User{
					Password: models.UserPassword{
						Hash: "$2a$10$V/6N1wsjheBVvWosVVVV2uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi",
					},
				}

				mock.
					On("UserGetByID", ctx, "1", false).
					Return(user, 1, nil).
					Once()
				hashMock.
					On("CompareWith", "wrong_password", "$2a$10$V/6N1wsjheBVvWosVVVV2uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi").
					Return(false).
					Once()
			},
			expected: NewErrUserPasswordNotMatch(nil),
		},
		{
			description:     "fail when unable to hash the new password",
			id:              "65fde3a72c4c7507c7f53c43",
			currentPassword: "secret",
			newPassword:     "newSecret",
			requiredMocks: func() {
				user := &models.User{
					Password: models.UserPassword{
						Plain: "secret",
						Hash:  "$2a$10$V/6N1wsjheBVvWosVVVV2uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi",
					},
				}

				mock.
					On("UserGetByID", ctx, "65fde3a72c4c7507c7f53c43", false).
					Return(user, 1, nil).
					Once()
				hashMock.
					On("CompareWith", "secret", "$2a$10$V/6N1wsjheBVvWosVVVV2uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi").
					Return(true).
					Once()
				hashMock.
					On("Do", "newSecret").
					Return("", errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrUserPasswordInvalid(errors.New("error", "", 0)),
		},
		{
			description:     "fail to update the user's password",
			id:              "65fde3a72c4c7507c7f53c43",
			currentPassword: "secret",
			newPassword:     "newSecret",
			requiredMocks: func() {
				user := &models.User{
					Password: models.UserPassword{
						Plain: "secret",
						Hash:  "$2a$10$V/6N1wsjheBVvWosVVVV2uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi",
					},
				}

				mock.
					On("UserGetByID", ctx, "65fde3a72c4c7507c7f53c43", false).
					Return(user, 1, nil).
					Once()
				hashMock.
					On("CompareWith", "secret", "$2a$10$V/6N1wsjheBVvWosVVVV2uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi").
					Return(true).
					Once()
				hashMock.
					On("Do", "newSecret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi", nil).
					Once()
				mock.
					On("UserUpdate", ctx, "65fde3a72c4c7507c7f53c43", &models.UserChanges{Password: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi"}).
					Return(errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrUserUpdate(
				&models.User{
					Password: models.UserPassword{
						Plain: "secret",
						Hash:  "$2a$10$V/6N1wsjheBVvWosVVVV2uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi",
					},
				},
				errors.New("error", "", 0),
			),
		},
		{
			description:     "succeeds to update the password",
			id:              "65fde3a72c4c7507c7f53c43",
			currentPassword: "secret",
			newPassword:     "newSecret",
			requiredMocks: func() {
				user := &models.User{
					Password: models.UserPassword{
						Plain: "secret",
						Hash:  "$2a$10$V/6N1wsjheBVvWosVVVV2uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi",
					},
				}

				mock.
					On("UserGetByID", ctx, "65fde3a72c4c7507c7f53c43", false).
					Return(user, 1, nil).
					Once()
				hashMock.
					On("CompareWith", "secret", "$2a$10$V/6N1wsjheBVvWosVVVV2uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi").
					Return(true).
					Once()
				hashMock.
					On("Do", "newSecret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi", nil).
					Once()
				mock.
					On("UserUpdate", ctx, "65fde3a72c4c7507c7f53c43", &models.UserChanges{Password: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi"}).
					Return(nil).
					Once()
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
