package services

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	uuid_mocks "github.com/shellhub-io/shellhub/pkg/uuid/mocks"
	"github.com/shellhub-io/shellhub/pkg/validator"
	"github.com/stretchr/testify/assert"
)

func TestSetup(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
	uuidMock := &uuid_mocks.Uuid{}
	uuid.DefaultBackend = uuidMock

	ctx := context.TODO()

	Err := errors.New("error", "", 0)

	cases := []struct {
		description   string
		req           requests.Setup
		requiredMocks func()
		expected      error
	}{
		{
			description: "Fail when cannot create the user",
			req: requests.Setup{
				Email:     "teste@google.com",
				Name:      "userteste",
				Username:  "userteste",
				Password:  "123456",
				Namespace: "teste-space",
			},
			requiredMocks: func() {
				clockMock.On("Now").Return(now).Once()
				user := &models.User{
					UserData: models.UserData{
						Name:     "userteste",
						Email:    "teste@google.com",
						Username: "userteste",
					},
					UserPassword: models.UserPassword{
						Password: validator.HashPassword("123456"),
					},
					Confirmed: true,
					CreatedAt: now,
				}
				mock.On("UserCreate", ctx, user).Return(Err).Once()
			},
			expected: NewErrUserDuplicated([]string{"userteste"}, Err),
		},
		{
			description: "Fail when cannot create namespace",
			req: requests.Setup{
				Email:     "teste@google.com",
				Name:      "userteste",
				Username:  "userteste",
				Password:  "123456",
				Namespace: "teste-space",
			},
			requiredMocks: func() {
				clockMock.On("Now").Return(now).Twice()
				uuidMock.On("Generate").Return("random_uuid").Once()
				user := &models.User{
					UserData: models.UserData{
						Name:     "userteste",
						Email:    "teste@google.com",
						Username: "userteste",
					},
					UserPassword: models.UserPassword{
						Password: validator.HashPassword("123456"),
					},
					Confirmed: true,
					CreatedAt: now,
				}
				namespace := &models.Namespace{
					Name:       "teste-space",
					Owner:      user.ID,
					MaxDevices: 0,
					Members: []models.Member{
						{
							ID:   user.ID,
							Role: guard.RoleOwner,
						},
					},
					CreatedAt: now,
				}
				mock.On("UserCreate", ctx, user).Return(nil).Once()
				mock.On("NamespaceCreate", ctx, namespace).Return(namespace, Err).Once()
			},
			expected: NewErrNamespaceDuplicated(Err),
		},
		{
			description: "Success to create the user and namespace",
			req: requests.Setup{
				Email:     "teste@google.com",
				Name:      "userteste",
				Username:  "userteste",
				Password:  "123456",
				Namespace: "teste-space",
			},
			requiredMocks: func() {
				clockMock.On("Now").Return(now).Twice()
				uuidMock.On("Generate").Return("random_uuid").Once()
				user := &models.User{
					UserData: models.UserData{
						Name:     "userteste",
						Email:    "teste@google.com",
						Username: "userteste",
					},
					UserPassword: models.UserPassword{
						Password: validator.HashPassword("123456"),
					},
					Confirmed: true,
					CreatedAt: now,
				}
				namespace := &models.Namespace{
					Name:       "teste-space",
					Owner:      user.ID,
					MaxDevices: 0,
					Members: []models.Member{
						{
							ID:   user.ID,
							Role: guard.RoleOwner,
						},
					},
					CreatedAt: now,
				}
				mock.On("UserCreate", ctx, user).Return(nil).Once()
				mock.On("NamespaceCreate", ctx, namespace).Return(namespace, nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()
			err := s.Setup(ctx, tc.req)
			assert.Equal(t, tc.expected, err)
		})
	}
}
