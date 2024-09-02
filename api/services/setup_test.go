package services

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
	uuid_mocks "github.com/shellhub-io/shellhub/pkg/uuid/mocks"
	"github.com/stretchr/testify/assert"
)

func TestSetup(t *testing.T) {
	storeMock := new(mocks.Store)
	clockMock := new(clockmock.Clock)
	clock.DefaultBackend = clockMock

	now := time.Now()
	clockMock.On("Now").Return(now)

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
				hashMock.
					On("Do", "secret").
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

				hashMock.
					On("Do", "secret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi", nil).
					Once()

				user := &models.User{
					Status:    models.UserStatusConfirmed,
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
				storeMock.On("UserCreate", ctx, user).Return("", errors.New("error", "", 0)).Once()
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

				hashMock.
					On("Do", "secret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi", nil).
					Once()

				user := &models.User{
					Status:    models.UserStatusConfirmed,
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
					Owner:      "000000000000000000000000",
					MaxDevices: 0,
					Members: []models.Member{
						{
							ID:   "000000000000000000000000",
							Role: authorizer.RoleOwner,
						},
					},
					Settings: &models.NamespaceSettings{
						SessionRecord:          false,
						ConnectionAnnouncement: "",
					},
					CreatedAt: now,
				}
				storeMock.On("UserCreate", ctx, user).Return("000000000000000000000000", nil).Once()
				storeMock.On("NamespaceCreate", ctx, namespace).Return(namespace, errors.New("error", "", 0)).Once()
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

				hashMock.
					On("Do", "secret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi", nil).
					Once()

				user := &models.User{
					Status:    models.UserStatusConfirmed,
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
					Owner:      "000000000000000000000000",
					MaxDevices: 0,
					Members: []models.Member{
						{
							ID:   "000000000000000000000000",
							Role: authorizer.RoleOwner,
						},
					},
					Settings: &models.NamespaceSettings{
						SessionRecord:          false,
						ConnectionAnnouncement: "",
					},
					CreatedAt: now,
				}
				storeMock.On("UserCreate", ctx, user).Return("000000000000000000000000", nil).Once()
				storeMock.On("NamespaceCreate", ctx, namespace).Return(namespace, nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

			err := service.Setup(ctx, tc.req)
			assert.Equal(t, tc.expected, err)
		})
	}
}
