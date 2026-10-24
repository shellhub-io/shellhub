package services

import (
	"context"
	stderrors "errors"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/hash"
	hashmock "github.com/shellhub-io/shellhub/pkg/hash/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSetup(t *testing.T) {
	hashMock := hashmock.NewMockHasher(t)
	hash.Set(t, hashMock)

	storeMock := mocks.NewMockStore(t)

	tenant := "00000000-0000-4000-0000-000000000000"
	uuid.Fix(t, tenant)

	now := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	clk := clock.NewFixed(now)

	ctx := context.TODO()

	cases := []struct {
		description   string
		req           requests.Setup
		requiredMocks func()
		expected      error
	}{
		{
			description: "Fail when the system lock cannot be taken",
			req: requests.Setup{
				Email:     "teste@google.com",
				Name:      "userteste",
				Username:  "userteste",
				Password:  "secret",
				Namespace: "userteste",
			},
			requiredMocks: func() {
				storeMock.On("SystemLock", ctx).Return(store.ErrInternal).Once()
			},
			expected: store.ErrInternal,
		},
		{
			description: "Fail when the system cannot be read",
			req: requests.Setup{
				Email:     "teste@google.com",
				Name:      "userteste",
				Username:  "userteste",
				Password:  "secret",
				Namespace: "userteste",
			},
			requiredMocks: func() {
				storeMock.On("SystemLock", ctx).Return(nil).Once()

				storeMock.On("SystemGet", ctx).Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description: "Fail when the instance is already set up",
			req: requests.Setup{
				Email:     "teste@google.com",
				Name:      "userteste",
				Username:  "userteste",
				Password:  "secret",
				Namespace: "userteste",
			},
			requiredMocks: func() {
				storeMock.On("SystemLock", ctx).Return(nil).Once()

				storeMock.On("SystemGet", ctx).Return(&models.System{Setup: true}, nil).Once()
			},
			expected: NewErrSetupCompleted(),
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
				storeMock.On("SystemLock", ctx).Return(nil).Once()

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
				storeMock.On("SystemLock", ctx).Return(nil).Once()

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
				storeMock.On("SystemLock", ctx).Return(nil).Once()

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
				storeMock.On("SystemLock", ctx).Return(nil).Once()

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
			description: "Fail when the namespace name is taken",
			req: requests.Setup{
				Email:     "teste@google.com",
				Name:      "userteste",
				Username:  "userteste",
				Password:  "secret",
				Namespace: "userteste",
			},
			requiredMocks: func() {
				storeMock.On("SystemLock", ctx).Return(nil).Once()

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
						ConnectionAnnouncement: envs.AnnouncementFor(envs.Community),
					},
					CreatedAt: now,
				}

				storeMock.On("NamespaceCreate", ctx, namespace).Return("", store.ErrDuplicate).Once()
			},
			expected: NewErrNamespaceDuplicated(store.ErrDuplicate),
		},
		{
			description: "Fail when the store cannot create the namespace",
			req: requests.Setup{
				Email:     "teste@google.com",
				Name:      "userteste",
				Username:  "userteste",
				Password:  "secret",
				Namespace: "userteste",
			},
			requiredMocks: func() {
				storeMock.On("SystemLock", ctx).Return(nil).Once()

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
						ConnectionAnnouncement: envs.AnnouncementFor(envs.Community),
					},
					CreatedAt: now,
				}

				storeMock.On("NamespaceCreate", ctx, namespace).Return("", errors.New("error", "", 0)).Once()
			},
			expected: NewErrNamespaceCreateStore(errors.New("error", "", 0)),
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
				storeMock.On("SystemLock", ctx).Return(nil).Once()

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
						ConnectionAnnouncement: envs.AnnouncementFor(envs.Community),
					},
					CreatedAt: now,
				}
				storeMock.On("NamespaceCreate", ctx, namespace).Return(tenant, nil).Once()
				storeMock.On("SystemSet", ctx, finalSystem).Return(nil).Once()

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
			storeMock.
				On("WithTransaction", ctx, mock.AnythingOfType("store.TransactionCb")).
				Return(func(ctx context.Context, cb store.TransactionCb) error { return cb(ctx) }).
				Once()
			tc.requiredMocks()

			service := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), WithClock(clk))

			res, err := service.Setup(ctx, tc.req)
			assert.Equal(t, tc.expected, err)
			if tc.expected == nil {
				assert.NotNil(t, res)
				assert.NotEmpty(t, res.Token)
			}
		})
	}
}
