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

func TestUpdateUser(t *testing.T) {
	type Expected struct {
		conflicts []string
		err       error
	}

	storeMock := new(mocks.Store)

	cases := []struct {
		description   string
		req           *requests.UpdateUser
		requiredMocks func(context.Context)
		expected      Expected
	}{
		{
			description: "Fail when user is not found",
			req: &requests.UpdateUser{
				UserID:        "000000000000000000000000",
				Name:          "John Doe",
				Username:      "john_doe",
				Email:         "john.doe@test.com",
				RecoveryEmail: "recovery@test.com",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(nil, NewErrUserNotFound("000000000000000000000000", nil)).
					Once()
			},
			expected: Expected{
				conflicts: []string{},
				err:       NewErrUserNotFound("000000000000000000000000", nil),
			},
		},
		{
			description: "Fail when recovery email is same as req's email",
			req: &requests.UpdateUser{
				UserID:        "000000000000000000000000",
				Name:          "John Doe",
				Username:      "john_doe",
				Email:         "john.doe@test.com",
				RecoveryEmail: "john.doe@test.com",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
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
			req: &requests.UpdateUser{
				UserID:        "000000000000000000000000",
				Name:          "John Doe",
				Username:      "john_doe",
				Email:         "john.doe@test.com",
				RecoveryEmail: "james.smith@test.com",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
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
			description: "Fail when conflict fields exists",
			req: &requests.UpdateUser{
				UserID:        "000000000000000000000000",
				Name:          "John Doe",
				Username:      "john_doe",
				Email:         "john.doe@test.com",
				RecoveryEmail: "recovery@test.com",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
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
						nil,
					).
					Once()
				storeMock.
					On("UserConflicts", ctx, &models.UserConflicts{Username: "john_doe", Email: "john.doe@test.com"}).
					Return([]string{"email"}, true, nil).
					Once()
			},
			expected: Expected{
				conflicts: []string{"email"},
				err:       NewErrUserDuplicated([]string{"email"}, nil),
			},
		},
		{
			description: "fails when the current password doesn't match with user's password",
			req: &requests.UpdateUser{
				UserID:          "000000000000000000000000",
				Name:            "John Doe",
				Username:        "john_doe",
				Email:           "john.doe@test.com",
				RecoveryEmail:   "recovery@test.com",
				Password:        "new-secret",
				CurrentPassword: "secret",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(
						&models.User{
							ID: "000000000000000000000000",
							UserData: models.UserData{
								Name:          "James Smith",
								Username:      "james_smith",
								Email:         "james.smith@shellhub.io",
								RecoveryEmail: "recover@test.com",
							},
							Password: models.UserPassword{
								Hash: "$2a$10$V/6N1wsjheBVvWosVVVV2uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi",
							},
						},
						nil,
					).
					Once()
				storeMock.
					On("UserConflicts", ctx, &models.UserConflicts{Username: "john_doe", Email: "john.doe@test.com"}).
					Return([]string{}, false, nil).
					Once()
				hashMock.
					On("CompareWith", "secret", "$2a$10$V/6N1wsjheBVvWosVVVV2uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi").
					Return(false).
					Once()
			},
			expected: Expected{
				conflicts: []string{},
				err:       NewErrUserPasswordNotMatch(nil),
			},
		},
		{
			description: "Fail when could not update user",
			req: &requests.UpdateUser{
				UserID:        "000000000000000000000000",
				Name:          "John Doe",
				Username:      "john_doe",
				Email:         "john.doe@test.com",
				RecoveryEmail: "recovery@test.com",
			},
			requiredMocks: func(ctx context.Context) {
				user := &models.User{
					ID: "000000000000000000000000",
					UserData: models.UserData{
						Name:          "James Smith",
						Username:      "james_smith",
						Email:         "james.smith@shellhub.io",
						RecoveryEmail: "recover@test.com",
					},
				}
				updatedUser := &models.User{
					ID: "000000000000000000000000",
					UserData: models.UserData{
						Name:          "John Doe",
						Username:      "john_doe",
						Email:         "john.doe@test.com",
						RecoveryEmail: "recovery@test.com",
					},
				}

				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(user, nil).
					Once()
				storeMock.
					On("UserConflicts", ctx, &models.UserConflicts{Username: "john_doe", Email: "john.doe@test.com"}).
					Return([]string{}, false, nil).
					Once()
				storeMock.
					On("UserUpdate", ctx, updatedUser).
					Return(errors.New("error", "", 0)).
					Once()
			},
			expected: Expected{
				conflicts: []string{},
				err: NewErrUserUpdate(
					&models.User{
						ID: "000000000000000000000000",
						UserData: models.UserData{
							Name:          "James Smith",
							Username:      "james_smith",
							Email:         "james.smith@shellhub.io",
							RecoveryEmail: "recover@test.com",
						},
					},
					errors.New("error", "", 0),
				),
			},
		},
		{
			description: "succeeds when only password fields are submitted",
			req: &requests.UpdateUser{
				UserID:          "000000000000000000000000",
				Password:        "new-secret",
				CurrentPassword: "secret",
			},
			requiredMocks: func(ctx context.Context) {
				user := &models.User{
					ID: "000000000000000000000000",
					UserData: models.UserData{
						Name:          "James Smith",
						Username:      "james_smith",
						Email:         "james.smith@shellhub.io",
						RecoveryEmail: "recover@test.com",
					},
					Password: models.UserPassword{
						Hash: "$2a$10$V/6N1wsjheBVvWosVVVV2uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi",
					},
				}

				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(user, nil).
					Once()
				storeMock.
					On("UserConflicts", ctx, &models.UserConflicts{}).
					Return([]string{}, false, nil).
					Once()
				hashMock.
					On("CompareWith", "secret", "$2a$10$V/6N1wsjheBVvWosVVVV2uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi").
					Return(true).
					Once()
				hashMock.
					On("Do", "new-secret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yv", nil).
					Once()

				updatedUser := *user
				updatedUser.Password = models.UserPassword{
					Plain: "new-secret",
					Hash:  "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yv",
				}

				storeMock.
					On("UserUpdate", ctx, &updatedUser).
					Return(nil).
					Once()
			},
			expected: Expected{
				conflicts: []string{},
				err:       nil,
			},
		},
		{
			description: "Success to update user",
			req: &requests.UpdateUser{
				UserID:        "000000000000000000000000",
				Name:          "John Doe",
				Username:      "john_doe",
				Email:         "john.doe@test.com",
				RecoveryEmail: "recovery@test.com",
			},
			requiredMocks: func(ctx context.Context) {
				user := &models.User{
					ID: "000000000000000000000000",
					UserData: models.UserData{
						Name:          "James Smith",
						Username:      "james_smith",
						Email:         "james.smith@shellhub.io",
						RecoveryEmail: "recover@test.com",
					},
				}
				updatedUser := &models.User{
					ID: "000000000000000000000000",
					UserData: models.UserData{
						Name:          "John Doe",
						Username:      "john_doe",
						Email:         "john.doe@test.com",
						RecoveryEmail: "recovery@test.com",
					},
				}

				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(user, nil).
					Once()
				storeMock.
					On("UserConflicts", ctx, &models.UserConflicts{Username: "john_doe", Email: "john.doe@test.com"}).
					Return([]string{}, false, nil).
					Once()
				storeMock.
					On("UserUpdate", ctx, updatedUser).
					Return(nil).
					Once()
			},
			expected: Expected{
				conflicts: []string{},
				err:       nil,
			},
		},
	}

	service := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()
			tc.requiredMocks(ctx)

			conflicts, err := service.UpdateUser(ctx, tc.req)
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
					On("UserResolve", ctx, store.UserIDResolver, "65fde3a72c4c7507c7f53c43").
					Return(nil, errors.New("error", "", 0)).
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
					On("UserResolve", ctx, store.UserIDResolver, "1").
					Return(user, nil).
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
					On("UserResolve", ctx, store.UserIDResolver, "65fde3a72c4c7507c7f53c43").
					Return(user, nil).
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
					On("UserResolve", ctx, store.UserIDResolver, "65fde3a72c4c7507c7f53c43").
					Return(user, nil).
					Once()
				hashMock.
					On("CompareWith", "secret", "$2a$10$V/6N1wsjheBVvWosVVVV2uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi").
					Return(true).
					Once()
				hashMock.
					On("Do", "newSecret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yv", nil).
					Once()

				expectedUser := *user
				expectedUser.Password.Plain = "newSecret"
				expectedUser.Password.Hash = "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yv"

				mock.
					On("UserUpdate", ctx, &expectedUser).
					Return(errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrUserUpdate(
				&models.User{
					Password: models.UserPassword{
						Plain: "newSecret",
						Hash:  "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yv",
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
					On("UserResolve", ctx, store.UserIDResolver, "65fde3a72c4c7507c7f53c43").
					Return(user, nil).
					Once()
				hashMock.
					On("CompareWith", "secret", "$2a$10$V/6N1wsjheBVvWosVVVV2uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi").
					Return(true).
					Once()
				hashMock.
					On("Do", "newSecret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yv", nil).
					Once()

				expectedUser := *user
				expectedUser.Password.Plain = "newSecret"
				expectedUser.Password.Hash = "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yv"

				mock.
					On("UserUpdate", ctx, &expectedUser).
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			services := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock)
			err := services.UpdatePasswordUser(ctx, tc.id, tc.currentPassword, tc.newPassword)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}
