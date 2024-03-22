package services

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/cli/pkg/inputs"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/password"
	passwordmock "github.com/shellhub-io/shellhub/pkg/password/mocks"
	"github.com/stretchr/testify/assert"
)

func TestUserCreate(t *testing.T) {
	type Expected struct {
		user *models.User
		err  error
	}

	mock := new(mocks.Store)
	passwordMock := &passwordmock.Password{}
	password.Backend = passwordMock
	ctx := context.TODO()
	now := clock.Now()

	mockClock := new(clockmock.Clock)
	clock.DefaultBackend = mockClock
	mockClock.On("Now").Return(now)

	cases := []struct {
		description   string
		requiredMocks func()
		username      string
		password      string
		email         string
		expected      Expected
	}{
		{
			description: "fails when email is invalid",
			username:    "john_doe",
			email:       "invalidmail.com",
			password:    "secret",
			requiredMocks: func() {
			},
			expected: Expected{nil, ErrUserDataInvalid},
		},
		{
			description: "fails when username is invalid",
			username:    "",
			email:       "john.doe@test.com",
			password:    "secret",
			requiredMocks: func() {
			},
			expected: Expected{nil, ErrUserDataInvalid},
		},
		{
			description: "fails when the password is invalid",
			username:    "john_doe",
			email:       "john.doe@test.com",
			password:    "secret",
			requiredMocks: func() {
				passwordMock.
					On("Hash", "secret").
					Return("", errors.New("error")).
					Once()
			},
			expected: Expected{nil, ErrUserPasswordInvalid},
		},
		{
			description: "fails when email is duplicated",
			username:    "john_doe",
			email:       "john.doe@test.com",
			password:    "secret",
			requiredMocks: func() {
				passwordMock.
					On("Hash", "secret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi", nil).
					Once()

				user := &models.User{
					UserData: models.UserData{
						Name:     "john_doe",
						Email:    "john.doe@test.com",
						Username: "john_doe",
					},
					Password: models.UserPassword{
						Plain: "secret",
						Hash:  "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi",
					},
					Confirmed:     true,
					CreatedAt:     clock.Now(),
					MaxNamespaces: MaxNumberNamespacesCommunity,
				}
				mock.On("UserCreate", ctx, user).Return(store.ErrDuplicate).Once()
				currentUser := &models.User{
					UserData: models.UserData{
						Name:     "jane_doe",
						Email:    "john.doe@test.com",
						Username: "jane_doe",
					},
					Password: models.UserPassword{
						Hash: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi",
					},
					Confirmed:     true,
					CreatedAt:     clock.Now(),
					MaxNamespaces: MaxNumberNamespacesCommunity,
				}
				mock.On("UserGetByUsername", ctx, "john_doe").Return(nil, errors.New("error")).Once()
				mock.On("UserGetByEmail", ctx, "john.doe@test.com").Return(currentUser, nil).Once()
			},
			expected: Expected{nil, ErrUserEmailExists},
		},
		{
			description: "fails when username is duplicated",
			username:    "john_doe",
			email:       "john.doe@test.com",
			password:    "secret",
			requiredMocks: func() {
				passwordMock.
					On("Hash", "secret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi", nil).
					Once()

				user := &models.User{
					UserData: models.UserData{
						Name:     "john_doe",
						Email:    "john.doe@test.com",
						Username: "john_doe",
					},
					Password: models.UserPassword{
						Plain: "secret",
						Hash:  "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi",
					},
					Confirmed:     true,
					CreatedAt:     clock.Now(),
					MaxNamespaces: MaxNumberNamespacesCommunity,
				}
				mock.On("UserCreate", ctx, user).Return(store.ErrDuplicate).Once()
				currentUser := &models.User{
					UserData: models.UserData{
						Name:     "john_doe",
						Email:    "jane.doe@test.com",
						Username: "john_doe",
					},
					Password: models.UserPassword{
						Hash: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi",
					},
					Confirmed:     true,
					CreatedAt:     clock.Now(),
					MaxNamespaces: MaxNumberNamespacesCommunity,
				}
				mock.On("UserGetByUsername", ctx, "john_doe").Return(currentUser, nil).Once()
				mock.On("UserGetByEmail", ctx, "john.doe@test.com").Return(nil, errors.New("error")).Once()
			},
			expected: Expected{nil, ErrUserNameExists},
		},
		{
			description: "fails when email and username is duplicated",
			username:    "john_doe",
			email:       "john.doe@test.com",
			password:    "secret",
			requiredMocks: func() {
				passwordMock.
					On("Hash", "secret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi", nil).
					Once()

				user := &models.User{
					UserData: models.UserData{
						Name:     "john_doe",
						Email:    "john.doe@test.com",
						Username: "john_doe",
					},
					Password: models.UserPassword{
						Plain: "secret",
						Hash:  "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi",
					},
					Confirmed:     true,
					CreatedAt:     clock.Now(),
					MaxNamespaces: MaxNumberNamespacesCommunity,
				}
				mock.On("UserCreate", ctx, user).Return(store.ErrDuplicate).Once()
				currentUser := &models.User{
					UserData: models.UserData{
						Name:     "john_doe",
						Email:    "john.doe@test.com",
						Username: "john_doe",
					},
					Password: models.UserPassword{
						Hash: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi",
					},
					Confirmed:     true,
					CreatedAt:     clock.Now(),
					MaxNamespaces: MaxNumberNamespacesCommunity,
				}
				mock.On("UserGetByUsername", ctx, "john_doe").Return(currentUser, nil).Once()
				mock.On("UserGetByEmail", ctx, "john.doe@test.com").Return(currentUser, nil).Once()
			},
			expected: Expected{nil, ErrUserNameAndEmailExists},
		},
		{
			description: "fails when some field is duplicated but unhandled",
			username:    "john_doe",
			email:       "john.doe@test.com",
			password:    "secret",
			requiredMocks: func() {
				passwordMock.
					On("Hash", "secret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi", nil).
					Once()

				user := &models.User{
					UserData: models.UserData{
						Name:     "john_doe",
						Email:    "john.doe@test.com",
						Username: "john_doe",
					},
					Password: models.UserPassword{
						Plain: "secret",
						Hash:  "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi",
					},
					Confirmed:     true,
					CreatedAt:     clock.Now(),
					MaxNamespaces: MaxNumberNamespacesCommunity,
				}
				mock.On("UserCreate", ctx, user).Return(store.ErrDuplicate).Once()
				mock.On("UserGetByUsername", ctx, "john_doe").Return(nil, nil).Once()
				mock.On("UserGetByEmail", ctx, "john.doe@test.com").Return(nil, nil).Once()
			},
			expected: Expected{nil, ErrUserUnhandledDuplicate},
		},
		{
			description: "fails creates a user",
			username:    "john_doe",
			email:       "john.doe@test.com",
			password:    "secret",
			requiredMocks: func() {
				passwordMock.
					On("Hash", "secret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi", nil).
					Once()

				user := &models.User{
					UserData: models.UserData{
						Name:     "john_doe",
						Email:    "john.doe@test.com",
						Username: "john_doe",
					},
					Password: models.UserPassword{
						Plain: "secret",
						Hash:  "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi",
					},
					Confirmed:     true,
					CreatedAt:     clock.Now(),
					MaxNamespaces: MaxNumberNamespacesCommunity,
				}
				mock.On("UserCreate", ctx, user).Return(errors.New("error")).Once()
			},
			expected: Expected{nil, ErrCreateNewUser},
		},
		{
			description: "successfully creates a user",
			username:    "john_doe",
			email:       "john.doe@test.com",
			password:    "secret",
			requiredMocks: func() {
				passwordMock.
					On("Hash", "secret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi", nil).
					Once()

				user := &models.User{
					UserData: models.UserData{
						Name:     "john_doe",
						Email:    "john.doe@test.com",
						Username: "john_doe",
					},
					Password: models.UserPassword{
						Plain: "secret",
						Hash:  "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi",
					},
					Confirmed:     true,
					CreatedAt:     clock.Now(),
					MaxNamespaces: MaxNumberNamespacesCommunity,
				}
				mock.On("UserCreate", ctx, user).Return(nil).Once()
			},
			expected: Expected{&models.User{
				UserData: models.UserData{
					Name:     "john_doe",
					Email:    "john.doe@test.com",
					Username: "john_doe",
				},
				Password: models.UserPassword{
					Plain: "secret",
					Hash:  "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi",
				},
				Confirmed:     true,
				CreatedAt:     clock.Now(),
				MaxNamespaces: MaxNumberNamespacesCommunity,
			}, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(mock))
			user, err := service.UserCreate(ctx, &inputs.UserCreate{Username: tc.username, Password: tc.password, Email: tc.email})

			assert.Equal(t, tc.expected, Expected{user, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestUserDelete(t *testing.T) {
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
			err := service.UserDelete(ctx, &inputs.UserDelete{Username: tc.username})
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestUserResetPassword(t *testing.T) {
	mock := new(mocks.Store)
	passwordMock := &passwordmock.Password{}
	password.Backend = passwordMock

	ctx := context.TODO()

	cases := []struct {
		description   string
		username      string
		password      string
		requiredMocks func()
		expected      error
	}{
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
			description: "fails to reset the user password",
			username:    "john_doe",
			password:    "secret",
			requiredMocks: func() {
				user := &models.User{ID: "507f191e810c19729de860ea"}

				mock.On("UserGetByUsername", ctx, "john_doe").Return(user, nil).Once()

				passwordMock.
					On("Hash", "secret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi", nil).
					Once()

				mock.
					On("UserUpdatePassword", ctx, "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi", "507f191e810c19729de860ea").
					Return(errors.New("error")).
					Once()
			},
			expected: ErrFailedUpdateUser,
		},
		{
			description: "successfully reset the user password",
			username:    "john_doe",
			password:    "secret",
			requiredMocks: func() {
				user := &models.User{ID: "507f191e810c19729de860ea"}

				mock.On("UserGetByUsername", ctx, "john_doe").Return(user, nil).Once()

				passwordMock.
					On("Hash", "secret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi", nil).
					Once()

				mock.
					On("UserUpdatePassword", ctx, "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi", "507f191e810c19729de860ea").
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()
			service := NewService(store.Store(mock))
			err := service.UserUpdate(ctx, &inputs.UserUpdate{Username: tc.username, Password: tc.password})
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}
