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
	uuid_mocks "github.com/shellhub-io/shellhub/pkg/uuid/mocks"
	"github.com/stretchr/testify/assert"
)

func TestSetup(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

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
				Password:  "secret",
				Namespace: "teste-space",
			},
			requiredMocks: func() {
				passwordMock.
					On("Hash", "secret").
					Return("", errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrUserPasswordInvalid(errors.New("error", "", 0)),
		},
		{
			description: "Fail when cannot create the user",
			req: requests.Setup{
				Email:     "teste@google.com",
				Name:      "userteste",
				Username:  "userteste",
				Password:  "secret",
				Namespace: "teste-space",
			},
			requiredMocks: func() {
				clockMock.On("Now").Return(now).Once()

				passwordMock.
					On("Hash", "secret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi", nil).
					Once()

				user := &models.User{
					Confirmed: true,
					CreatedAt: now,
					UserData: models.UserData{
						Name:     "userteste",
						Email:    "teste@google.com",
						Username: "userteste",
					},
					Password: models.UserPassword{
						Plain: "secret",
						Hash:  "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi",
					},
				}
				mock.On("UserCreate", ctx, user).Return(errors.New("error", "", 0)).Once()
			},
			expected: NewErrUserDuplicated([]string{"userteste"}, errors.New("error", "", 0)),
		},
		{
			description: "Fail when cannot create namespace",
			req: requests.Setup{
				Email:     "teste@google.com",
				Name:      "userteste",
				Username:  "userteste",
				Password:  "secret",
				Namespace: "teste-space",
			},
			requiredMocks: func() {
				clockMock.On("Now").Return(now).Twice()

				uuidMock := &uuid_mocks.Uuid{}
				uuidMock.On("Generate").Return("random_uuid").Once()

				passwordMock.
					On("Hash", "secret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi", nil).
					Once()

				user := &models.User{
					Confirmed: true,
					CreatedAt: now,
					UserData: models.UserData{
						Name:     "userteste",
						Email:    "teste@google.com",
						Username: "userteste",
					},
					Password: models.UserPassword{
						Plain: "secret",
						Hash:  "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi",
					},
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
					Settings: &models.NamespaceSettings{
						SessionRecord:          false,
						ConnectionAnnouncement: "",
					},
					CreatedAt: now,
				}
				mock.On("UserCreate", ctx, user).Return(nil).Once()
				mock.On("NamespaceCreate", ctx, namespace).Return(namespace, errors.New("error", "", 0)).Once()
			},
			expected: NewErrNamespaceDuplicated(errors.New("error", "", 0)),
		},
		{
			description: "Success to create the user and namespace",
			req: requests.Setup{
				Email:     "teste@google.com",
				Name:      "userteste",
				Username:  "userteste",
				Password:  "secret",
				Namespace: "teste-space",
			},
			requiredMocks: func() {
				clockMock.On("Now").Return(now).Twice()
				uuidMock := &uuid_mocks.Uuid{}
				uuidMock.On("Generate").Return("random_uuid").Once()

				passwordMock.
					On("Hash", "secret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi", nil).
					Once()

				user := &models.User{
					Confirmed: true,
					CreatedAt: now,
					UserData: models.UserData{
						Name:     "userteste",
						Email:    "teste@google.com",
						Username: "userteste",
					},
					Password: models.UserPassword{
						Plain: "secret",
						Hash:  "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi",
					},
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
					Settings: &models.NamespaceSettings{
						SessionRecord:          false,
						ConnectionAnnouncement: "",
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

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

			err := service.Setup(ctx, tc.req)
			assert.Equal(t, tc.expected, err)
		})
	}
}
