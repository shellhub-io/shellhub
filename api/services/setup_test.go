package services

import (
	"context"
	stderrors "errors"
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
	"github.com/shellhub-io/shellhub/pkg/uuid"
	uuidmock "github.com/shellhub-io/shellhub/pkg/uuid/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSetup(t *testing.T) {
	storeMock := mocks.NewMockStore(t)

	clockMock := clockmock.NewMockClock(t)
	prevClockBackend := clock.DefaultBackend
	t.Cleanup(func() { clock.DefaultBackend = prevClockBackend })
	clock.DefaultBackend = clockMock

	tenant := "00000000-0000-4000-0000-000000000000"

	uuidMock := uuidmock.NewMockUUID(t)
	prevUUIDBackend := uuid.DefaultBackend
	t.Cleanup(func() { uuid.DefaultBackend = prevUUIDBackend })
	uuid.DefaultBackend = uuidMock
	uuidMock.On("Generate").Return(tenant)

	now := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	clockMock.On("Now").Return(now)

	// Setup calls envs.IsDevelopment() to decide the namespace tenant; not development here.
	envMock.On("Get", "SHELLHUB_ENV").Return("")

	ctx := context.TODO()

	cases := []struct {
		description   string
		req           requests.Setup
		requiredMocks func()
		expected      error
	}{
		{
			description: "Fail when setup isn't allowed",
			req: requests.Setup{
				Email:     "teste@google.com",
				Name:      "userteste",
				Username:  "userteste",
				Password:  "secret",
				Namespace: "userteste",
			},
			requiredMocks: func() {
				storeMock.On("SystemGet", ctx).Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrSetupForbidden(errors.New("error", "", 0)),
		},
		{
			description: "Fail when cannot hash the password",
			req: requests.Setup{
				Email:     "teste@google.com",
				Name:      "userteste",
				Username:  "userteste",
				Password:  "secret",
				Namespace: "userteste",
			},
			requiredMocks: func() {
				storeMock.On("SystemGet", ctx).Return(&models.System{
					Setup: false,
				}, nil).Once()

				hashMock.
					On("Do", "secret").
					Return("", errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrUserPasswordInvalid(errors.New("error", "", 0)),
		},
		{
			description: "Fail when cannot create the user due to duplicate field",
			req: requests.Setup{
				Email:     "teste@google.com",
				Name:      "userteste",
				Username:  "userteste",
				Password:  "secret",
				Namespace: "userteste",
			},
			requiredMocks: func() {
				storeMock.On("SystemGet", ctx).Return(&models.System{
					Setup: false,
				}, nil).Once()

				hashMock.
					On("Do", "secret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi", nil).
					Once()

				user := &models.User{
					Origin:    models.UserOriginLocal,
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
					MaxNamespaces: -1,
					Preferences: models.UserPreferences{
						AuthMethods: []models.UserAuthMethod{models.UserAuthMethodLocal},
					},
					Admin: true,
				}

				dupErr := stderrors.Join(store.ErrDuplicate, store.DuplicateFieldError{Field: "username"})
				storeMock.On("UserCreate", ctx, user).Return("", dupErr).Once()
			},
			expected: NewErrUserDuplicated(
				[]string{"username"},
				stderrors.Join(store.ErrDuplicate, store.DuplicateFieldError{Field: "username"}),
			),
		},
		{
			description: "Fail when cannot create the user due to unhandled duplicate (no field info)",
			req: requests.Setup{
				Email:     "teste@google.com",
				Name:      "userteste",
				Username:  "userteste",
				Password:  "secret",
				Namespace: "userteste",
			},
			requiredMocks: func() {
				storeMock.On("SystemGet", ctx).Return(&models.System{
					Setup: false,
				}, nil).Once()

				hashMock.
					On("Do", "secret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi", nil).
					Once()

				user := &models.User{
					Origin:    models.UserOriginLocal,
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
					MaxNamespaces: -1,
					Preferences: models.UserPreferences{
						AuthMethods: []models.UserAuthMethod{models.UserAuthMethodLocal},
					},
					Admin: true,
				}

				storeMock.On("UserCreate", ctx, user).Return("", store.ErrDuplicate).Once()
			},
			expected: NewErrUserUnhandledDuplicate(),
		},
		{
			description: "Fail when cannot create the user due to a generic store error",
			req: requests.Setup{
				Email:     "teste@google.com",
				Name:      "userteste",
				Username:  "userteste",
				Password:  "secret",
				Namespace: "userteste",
			},
			requiredMocks: func() {
				storeMock.On("SystemGet", ctx).Return(&models.System{
					Setup: false,
				}, nil).Once()

				hashMock.
					On("Do", "secret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi", nil).
					Once()

				user := &models.User{
					Origin:    models.UserOriginLocal,
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
					MaxNamespaces: -1,
					Preferences: models.UserPreferences{
						AuthMethods: []models.UserAuthMethod{models.UserAuthMethodLocal},
					},
					Admin: true,
				}

				storeMock.On("UserCreate", ctx, user).Return("", store.ErrInternal).Once()
			},
			expected: store.ErrInternal,
		},
		{
			description: "Fail when cannot create namespace, and user deletion fails",
			req: requests.Setup{
				Email:     "teste@google.com",
				Name:      "userteste",
				Username:  "userteste",
				Password:  "secret",
				Namespace: "userteste",
			},
			requiredMocks: func() {
				user := &models.User{
					Origin:    models.UserOriginLocal,
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
					MaxNamespaces: -1,
					Preferences: models.UserPreferences{
						AuthMethods: []models.UserAuthMethod{models.UserAuthMethodLocal},
					},
					Admin: true,
				}
				userWithID := &models.User{
					ID:        "000000000000000000000000",
					Origin:    models.UserOriginLocal,
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
					MaxNamespaces: -1,
					Preferences: models.UserPreferences{
						AuthMethods: []models.UserAuthMethod{models.UserAuthMethodLocal},
					},
					Admin: true,
				}

				storeMock.On("SystemGet", ctx).Return(&models.System{
					Setup: false,
				}, nil).Once()

				hashMock.
					On("Do", "secret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi", nil).
					Once()

				storeMock.On("UserCreate", ctx, user).Return("000000000000000000000000", nil).Once()

				namespace := &models.Namespace{
					Name:       "userteste",
					TenantID:   tenant,
					Owner:      "000000000000000000000000",
					MaxDevices: -1,
					Type:       models.TypePersonal,
					Members: []models.Member{
						{
							ID:      "000000000000000000000000",
							Role:    authorizer.RoleOwner,
							AddedAt: now,
						},
					},
					Settings: &models.NamespaceSettings{
						SessionRecord:          false,
						ConnectionAnnouncement: models.DefaultAnnouncementMessage,
					},
					CreatedAt: now,
				}

				storeMock.On("NamespaceCreate", ctx, namespace).Return("", errors.New("error", "", 0)).Once()
				storeMock.On("UserDelete", ctx, userWithID).Return(errors.New("error", "", 0)).Once()
			},
			expected: NewErrUserDelete(errors.New("error", "", 0)),
		},
		{
			description: "Fail when cannot create namespace, and user deletion fails",
			req: requests.Setup{
				Email:     "teste@google.com",
				Name:      "userteste",
				Username:  "userteste",
				Password:  "secret",
				Namespace: "userteste",
			},
			requiredMocks: func() {
				user := &models.User{
					Origin:    models.UserOriginLocal,
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
					MaxNamespaces: -1,
					Preferences: models.UserPreferences{
						AuthMethods: []models.UserAuthMethod{models.UserAuthMethodLocal},
					},
					Admin: true,
				}
				userWithID := &models.User{
					ID:        "000000000000000000000000",
					Origin:    models.UserOriginLocal,
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
					MaxNamespaces: -1,
					Preferences: models.UserPreferences{
						AuthMethods: []models.UserAuthMethod{models.UserAuthMethodLocal},
					},
					Admin: true,
				}

				storeMock.On("SystemGet", ctx).Return(&models.System{
					Setup: false,
				}, nil).Once()

				hashMock.
					On("Do", "secret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi", nil).
					Once()

				storeMock.On("UserCreate", ctx, user).Return("000000000000000000000000", nil).Once()

				namespace := &models.Namespace{
					Name:       "userteste",
					TenantID:   tenant,
					Owner:      "000000000000000000000000",
					MaxDevices: -1,
					Type:       models.TypePersonal,
					Members: []models.Member{
						{
							ID:      "000000000000000000000000",
							Role:    authorizer.RoleOwner,
							AddedAt: now,
						},
					},
					Settings: &models.NamespaceSettings{
						SessionRecord:          false,
						ConnectionAnnouncement: models.DefaultAnnouncementMessage,
					},
					CreatedAt: now,
				}

				storeMock.On("NamespaceCreate", ctx, namespace).Return("", errors.New("error", "", 0)).Once()
				storeMock.On("UserDelete", ctx, userWithID).Return(errors.New("error", "", 0)).Once()
			},
			expected: NewErrUserDelete(errors.New("error", "", 0)),
		},
		{
			description: "Success to create the user and namespace",
			req: requests.Setup{
				Email:     "teste@google.com",
				Name:      "userteste",
				Username:  "userteste",
				Password:  "secret",
				Namespace: "userteste",
			},
			requiredMocks: func() {
				initialSystem := &models.System{Setup: false}
				finalSystem := &models.System{Setup: true, InstanceTenantID: tenant}

				storeMock.On("SystemGet", ctx).Return(initialSystem, nil).Once()

				hashMock.On("Do", "secret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi", nil).
					Once()

				user := &models.User{
					Origin:    models.UserOriginLocal,
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
					MaxNamespaces: -1,
					Preferences: models.UserPreferences{
						AuthMethods: []models.UserAuthMethod{models.UserAuthMethodLocal},
					},
					Admin: true,
				}

				storeMock.On("UserCreate", ctx, user).Return("000000000000000000000000", nil).Once()

				namespace := &models.Namespace{
					Name:       "userteste",
					TenantID:   tenant,
					Owner:      "000000000000000000000000",
					MaxDevices: -1,
					Type:       models.TypePersonal,
					Members: []models.Member{
						{
							ID:      "000000000000000000000000",
							Role:    authorizer.RoleOwner,
							AddedAt: now,
						},
					},
					Settings: &models.NamespaceSettings{
						SessionRecord:          false,
						ConnectionAnnouncement: models.DefaultAnnouncementMessage,
					},
					CreatedAt: now,
				}
				storeMock.On("NamespaceCreate", ctx, namespace).Return(tenant, nil).Once()
				storeMock.On("SystemSet", ctx, finalSystem).Return(nil).Once()

				// Setup mints an authenticated session for the new admin (auto-login) by
				// delegating to CreateUserToken, which resolves the user and namespace back.
				createdUser := &models.User{
					ID:        "000000000000000000000000",
					Origin:    models.UserOriginLocal,
					Status:    models.UserStatusConfirmed,
					CreatedAt: now,
					UserData: models.UserData{
						Name:     "userteste",
						Email:    "teste@google.com",
						Username: "userteste",
					},
					MaxNamespaces: -1,
					Preferences: models.UserPreferences{
						AuthMethods: []models.UserAuthMethod{models.UserAuthMethodLocal},
					},
					Admin: true,
				}
				storeMock.On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(createdUser, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(&models.Namespace{
						Name:     "userteste",
						TenantID: tenant,
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{ID: "000000000000000000000000", Role: authorizer.RoleOwner, AddedAt: now},
						},
					}, nil).Once()
				storeMock.On("UserUpdatePreferredNamespace", ctx, mock.Anything, mock.Anything).Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

			res, err := service.Setup(ctx, tc.req)
			assert.Equal(t, tc.expected, err)
			if tc.expected == nil {
				assert.NotNil(t, res)
				assert.NotEmpty(t, res.Token)
			}
		})
	}
}
