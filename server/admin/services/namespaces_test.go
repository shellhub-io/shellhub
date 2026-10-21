package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	cachemock "github.com/shellhub-io/shellhub/pkg/cache/mocks"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/admin/inputs"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNamespaceCreate(t *testing.T) {
	type Expected struct {
		namespace *models.Namespace
		err       error
	}
	mock := new(mocks.MockStore)
	ctx := context.TODO()
	errStoreDown := errors.New("connection refused")

	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		description   string
		namespace     string
		username      string
		tenant        string
		typeNamespace string
		edition       envs.Edition
		requiredMocks func()
		expected      Expected
		cause         error
	}{
		{
			description:   "fails when could not find a user",
			namespace:     "namespace",
			username:      "john_doe",
			tenant:        "00000000-0000-4000-0000-000000000000",
			typeNamespace: "",
			edition:       envs.Community,
			requiredMocks: func() {
				mock.On("UserResolve", ctx, store.UserUsernameResolver, "john_doe").Return(nil, errors.New("error")).Once()
			},
			expected: Expected{nil, ErrUserNotFound},
		},
		{
			description:   "fails when namespace is duplicated",
			namespace:     "namespace",
			username:      "john_doe",
			tenant:        "00000000-0000-0000-0000-000000000000",
			typeNamespace: "",
			edition:       envs.Community,
			requiredMocks: func() {
				user := &models.User{
					ID: "507f191e810c19729de860ea",
					UserData: models.UserData{
						Name:     "John Doe",
						Email:    "john.doe@test.com",
						Username: "john_doe",
					},
				}
				mock.On("UserResolve", ctx, store.UserUsernameResolver, "john_doe").Return(user, nil).Once()
				namespace := &models.Namespace{
					Name:     "namespace",
					Owner:    "507f191e810c19729de860ea",
					TenantID: "00000000-0000-0000-0000-000000000000",
					Type:     models.TypeTeam,
					Members: []models.Member{
						{
							ID:      "507f191e810c19729de860ea",
							Role:    "owner",
							AddedAt: now,
						},
					},
					Settings: &models.NamespaceSettings{
						SessionRecord:          true,
						ConnectionAnnouncement: envs.AnnouncementFor(envs.Community),
					},
					MaxDevices: MaxNumberDevicesUnlimited,
					CreatedAt:  now,
				}
				mock.On("NamespaceCreate", ctx, namespace).Return("", store.ErrDuplicate).Once()
			},
			expected: Expected{nil, ErrDuplicateNamespace},
		},
		{
			description:   "fails without claiming a duplicate when the store fails for another reason",
			namespace:     "namespace",
			username:      "john_doe",
			tenant:        "00000000-0000-0000-0000-000000000000",
			typeNamespace: "",
			edition:       envs.Community,
			requiredMocks: func() {
				user := &models.User{
					ID: "507f191e810c19729de860ea",
					UserData: models.UserData{
						Name:     "John Doe",
						Email:    "john.doe@test.com",
						Username: "john_doe",
					},
				}
				mock.On("UserResolve", ctx, store.UserUsernameResolver, "john_doe").Return(user, nil).Once()
				namespace := &models.Namespace{
					Name:     "namespace",
					Owner:    "507f191e810c19729de860ea",
					TenantID: "00000000-0000-0000-0000-000000000000",
					Type:     models.TypeTeam,
					Members: []models.Member{
						{
							ID:      "507f191e810c19729de860ea",
							Role:    "owner",
							AddedAt: now,
						},
					},
					Settings: &models.NamespaceSettings{
						SessionRecord:          true,
						ConnectionAnnouncement: envs.AnnouncementFor(envs.Community),
					},
					MaxDevices: MaxNumberDevicesUnlimited,
					CreatedAt:  now,
				}
				mock.On("NamespaceCreate", ctx, namespace).Return("", errStoreDown).Once()
			},
			expected: Expected{nil, ErrCreateNewNamespace},
			cause:    errStoreDown,
		},
		{
			description:   "succeeds in creating a namespace when user and namespace data are valid - Community",
			namespace:     "namespace",
			username:      "john_doe",
			tenant:        "00000000-0000-0000-0000-000000000000",
			typeNamespace: "",
			edition:       envs.Community,
			requiredMocks: func() {
				user := &models.User{
					ID: "507f191e810c19729de860ea",
					UserData: models.UserData{
						Name:     "John Doe",
						Email:    "john.doe@test.com",
						Username: "john_doe",
					},
				}
				mock.On("UserResolve", ctx, store.UserUsernameResolver, "john_doe").Return(user, nil).Once()
				namespace := &models.Namespace{
					Name:     "namespace",
					Owner:    "507f191e810c19729de860ea",
					TenantID: "00000000-0000-0000-0000-000000000000",
					Type:     models.TypeTeam,
					Members: []models.Member{
						{
							ID:      "507f191e810c19729de860ea",
							Role:    "owner",
							AddedAt: now,
						},
					},
					Settings: &models.NamespaceSettings{
						SessionRecord:          true,
						ConnectionAnnouncement: envs.AnnouncementFor(envs.Community),
					},
					MaxDevices: MaxNumberDevicesUnlimited,
					CreatedAt:  now,
				}
				mock.On("NamespaceCreate", ctx, namespace).Return("00000000-0000-0000-0000-000000000000", nil).Once()
			},
			expected: Expected{&models.Namespace{
				Name:     "namespace",
				Owner:    "507f191e810c19729de860ea",
				TenantID: "00000000-0000-0000-0000-000000000000",
				Type:     models.TypeTeam,
				Members: []models.Member{
					{
						ID:      "507f191e810c19729de860ea",
						Role:    "owner",
						AddedAt: now,
					},
				},
				Settings: &models.NamespaceSettings{
					SessionRecord:          true,
					ConnectionAnnouncement: envs.AnnouncementFor(envs.Community),
				},
				MaxDevices: MaxNumberDevicesUnlimited,
				CreatedAt:  now,
			}, nil},
		},
		{
			description:   "succeeds in creating a namespace when user and namespace data are valid - Cloud - type team",
			namespace:     "namespace",
			username:      "john_doe",
			tenant:        "00000000-0000-0000-0000-000000000000",
			typeNamespace: "team",
			edition:       envs.Cloud,
			requiredMocks: func() {
				user := &models.User{
					ID: "507f191e810c19729de860ea",
					UserData: models.UserData{
						Name:     "John Doe",
						Email:    "john.doe@test.com",
						Username: "john_doe",
					},
				}
				mock.On("UserResolve", ctx, store.UserUsernameResolver, "john_doe").Return(user, nil).Once()
				namespace := &models.Namespace{
					Name:     "namespace",
					Owner:    "507f191e810c19729de860ea",
					TenantID: "00000000-0000-0000-0000-000000000000",
					Type:     models.TypeTeam,
					Members: []models.Member{
						{
							ID:      "507f191e810c19729de860ea",
							Role:    "owner",
							AddedAt: now,
						},
					},
					Settings: &models.NamespaceSettings{
						SessionRecord:          true,
						ConnectionAnnouncement: envs.AnnouncementFor(envs.Cloud),
					},
					MaxDevices: MaxNumberDevicesLimited,
					CreatedAt:  now,
				}
				mock.On("NamespaceCreate", ctx, namespace).Return("00000000-0000-0000-0000-000000000000", nil).Once()
			},
			expected: Expected{&models.Namespace{
				Name:     "namespace",
				Owner:    "507f191e810c19729de860ea",
				TenantID: "00000000-0000-0000-0000-000000000000",
				Type:     models.TypeTeam,
				Members: []models.Member{
					{
						ID:      "507f191e810c19729de860ea",
						Role:    "owner",
						AddedAt: now,
					},
				},
				Settings: &models.NamespaceSettings{
					SessionRecord:          true,
					ConnectionAnnouncement: envs.AnnouncementFor(envs.Cloud),
				},
				MaxDevices: MaxNumberDevicesLimited,
				CreatedAt:  now,
			}, nil},
		},
		{
			description:   "succeeds in creating a namespace when user and namespace data are valid - Cloud",
			namespace:     "namespace",
			username:      "john_doe",
			tenant:        "00000000-0000-0000-0000-000000000000",
			typeNamespace: "",
			edition:       envs.Cloud,
			requiredMocks: func() {
				user := &models.User{
					ID: "507f191e810c19729de860ea",
					UserData: models.UserData{
						Name:     "John Doe",
						Email:    "john.doe@test.com",
						Username: "john_doe",
					},
				}
				mock.On("UserResolve", ctx, store.UserUsernameResolver, "john_doe").Return(user, nil).Once()
				namespace := &models.Namespace{
					Name:     "namespace",
					Owner:    "507f191e810c19729de860ea",
					TenantID: "00000000-0000-0000-0000-000000000000",
					Type:     models.TypeTeam,
					Members: []models.Member{
						{
							ID:      "507f191e810c19729de860ea",
							Role:    "owner",
							AddedAt: now,
						},
					},
					Settings: &models.NamespaceSettings{
						SessionRecord:          true,
						ConnectionAnnouncement: envs.AnnouncementFor(envs.Cloud),
					},
					MaxDevices: MaxNumberDevicesLimited,
					CreatedAt:  now,
				}
				mock.On("NamespaceCreate", ctx, namespace).Return("00000000-0000-0000-0000-000000000000", nil).Once()
			},
			expected: Expected{&models.Namespace{
				Name:     "namespace",
				Owner:    "507f191e810c19729de860ea",
				TenantID: "00000000-0000-0000-0000-000000000000",
				Type:     models.TypeTeam,
				Members: []models.Member{
					{
						ID:      "507f191e810c19729de860ea",
						Role:    "owner",
						AddedAt: now,
					},
				},
				Settings: &models.NamespaceSettings{
					SessionRecord:          true,
					ConnectionAnnouncement: envs.AnnouncementFor(envs.Cloud),
				},
				MaxDevices: MaxNumberDevicesLimited,
				CreatedAt:  now,
			}, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			s := NewService(store.Store(mock), WithEdition(tc.edition), WithClock(clock.NewFixed(now)))
			ns, err := s.NamespaceCreate(ctx, &inputs.NamespaceCreate{Namespace: tc.namespace, Owner: tc.username, TenantID: tc.tenant, Type: tc.typeNamespace})
			assert.Equal(t, tc.expected.namespace, ns)
			require.ErrorIs(t, err, tc.expected.err)
			if tc.cause != nil {
				require.ErrorIs(t, err, tc.cause)
			}

			mock.AssertExpectations(t)
		})
	}
}

func TestNamespaceCreateRequiresVerifiedDeviceKeyAsConfigured(t *testing.T) {
	ctx := context.TODO()

	for _, required := range []bool{true, false} {
		t.Run(map[bool]string{true: "required", false: "not required"}[required], func(t *testing.T) {
			st := new(mocks.MockStore)
			st.On("UserResolve", ctx, store.UserUsernameResolver, "john_doe").Return(&models.User{ID: "507f191e810c19729de860ea"}, nil).Once()
			st.On("NamespaceCreate", ctx, testifymock.MatchedBy(func(ns *models.Namespace) bool {
				return ns.Settings.RequireVerifiedDeviceKey == required
			})).Return("00000000-0000-4000-0000-000000000000", nil).Once()

			s := NewService(store.Store(st), WithNamespaceRequireVerifiedDeviceKey(required))
			_, err := s.NamespaceCreate(ctx, &inputs.NamespaceCreate{Namespace: "namespace", Owner: "john_doe"})
			require.NoError(t, err)

			st.AssertExpectations(t)
		})
	}
}

func TestNamespaceAddMember(t *testing.T) {
	type Expected struct {
		namespace *models.Namespace
		err       error
	}

	mock := new(mocks.MockStore)
	ctx := context.TODO()
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		description   string
		username      string
		namespace     string
		role          authorizer.Role
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when could not find a user",
			username:    "john",
			namespace:   "namespace",
			role:        authorizer.RoleObserver,
			requiredMocks: func() {
				mock.On("UserResolve", ctx, store.UserUsernameResolver, "john").Return(nil, errors.New("error")).Once()
			},
			expected: Expected{nil, ErrUserNotFound},
		},
		{
			description: "fails when could not find a namespace",
			username:    "john",
			namespace:   "invalid_namespace",
			role:        authorizer.RoleObserver,
			requiredMocks: func() {
				user := &models.User{
					ID: "507f191e810c19729de860ea",
					UserData: models.UserData{
						Name:     "John Doe",
						Email:    "john.doe@test.com",
						Username: "john",
					},
				}
				mock.On("UserResolve", ctx, store.UserUsernameResolver, "john").Return(user, nil).Once()
				mock.On("NamespaceResolve", ctx, store.NamespaceNameResolver, "invalid_namespace").Return(nil, errors.New("error")).Once()
			},
			expected: Expected{nil, ErrNamespaceNotFound},
		},
		{
			description: "successfully add user to the Namespace",
			username:    "john",
			namespace:   "namespace",
			role:        authorizer.RoleObserver,
			requiredMocks: func() {
				user := &models.User{
					ID: "507f191e810c19729de860ea",
					UserData: models.UserData{
						Name:     "John Doe",
						Email:    "john.doe@test.com",
						Username: "john",
					},
				}
				mock.On("UserResolve", ctx, store.UserUsernameResolver, "john").Return(user, nil).Once()
				namespace := &models.Namespace{
					Name:     "namespace",
					Owner:    "507f191e810c19729de860ea",
					TenantID: "00000000-0000-0000-0000-000000000000",
					Members:  []models.Member{{ID: "507f191e810c19729de860ea", Role: "owner"}},
					Settings: &models.NamespaceSettings{
						SessionRecord: true,
					},
					CreatedAt: now,
				}
				mock.On("NamespaceResolve", ctx, store.NamespaceNameResolver, "namespace").Return(namespace, nil).Once()
				mock.On("NamespaceCreateMembership", ctx, scope.MustBounded("00000000-0000-0000-0000-000000000000"), &models.Member{
					ID:      "507f191e810c19729de860ea",
					Role:    authorizer.RoleObserver,
					AddedAt: now,
				}).Return(nil).Once()
			},
			expected: Expected{&models.Namespace{
				Name:     "namespace",
				Owner:    "507f191e810c19729de860ea",
				TenantID: "00000000-0000-0000-0000-000000000000",
				Members:  []models.Member{{ID: "507f191e810c19729de860ea", Role: "owner"}},
				Settings: &models.NamespaceSettings{
					SessionRecord: true,
				},
				CreatedAt: now,
			}, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			s := NewService(store.Store(mock), WithClock(clock.NewFixed(now)))
			ns, err := s.NamespaceAddMember(ctx, &inputs.MemberAdd{Username: tc.username, Namespace: tc.namespace, Role: tc.role})
			assert.Equal(t, tc.expected, Expected{ns, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestNamespaceRemoveMember(t *testing.T) {
	type Expected struct {
		user *models.Namespace
		err  error
	}

	mock := new(mocks.MockStore)

	ctx := context.TODO()
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		description   string
		username      string
		namespace     string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when could not find a user",
			username:    "john_doe",
			namespace:   "namespace",
			requiredMocks: func() {
				mock.On("UserResolve", ctx, store.UserUsernameResolver, "john_doe").Return(nil, errors.New("error")).Once()
			},
			expected: Expected{nil, ErrUserNotFound},
		},
		{
			description: "fails when could not find a namespace",
			username:    "john_doe",
			namespace:   "namespace",
			requiredMocks: func() {
				user := &models.User{
					ID: "507f191e810c19729de860ea",
					UserData: models.UserData{
						Name:     "John Doe",
						Email:    "john.doe@test.com",
						Username: "john_doe",
					},
				}
				mock.On("UserResolve", ctx, store.UserUsernameResolver, "john_doe").Return(user, nil).Once()
				mock.On("NamespaceResolve", ctx, store.NamespaceNameResolver, "namespace").Return(nil, errors.New("error")).Once()
			},
			expected: Expected{nil, ErrNamespaceNotFound},
		},
		{
			description: "fails remove member from the namespace",
			username:    "john_doe",
			namespace:   "namespace",
			requiredMocks: func() {
				user := &models.User{
					ID: "507f191e810c19729de860ea",
					UserData: models.UserData{
						Name:     "John Doe",
						Email:    "john.doe@test.com",
						Username: "john_doe",
					},
				}
				mock.On("UserResolve", ctx, store.UserUsernameResolver, "john_doe").Return(user, nil).Once()
				namespace := &models.Namespace{
					Name:     "namespace",
					Owner:    "608f191e810c19729de860eb",
					TenantID: "00000000-0000-0000-0000-000000000000",
					Members: []models.Member{
						{ID: "608f191e810c19729de860eb", Role: "owner"},
						{ID: "507f191e810c19729de860ea", Role: "operator"},
					},
					Settings: &models.NamespaceSettings{
						SessionRecord: true,
					},
					CreatedAt: now,
				}
				mock.On("NamespaceResolve", ctx, store.NamespaceNameResolver, "namespace").Return(namespace, nil).Once()
				mock.On("NamespaceDepartMember", ctx, scope.MustBounded("00000000-0000-0000-0000-000000000000"), "507f191e810c19729de860ea", store.MemberDeparture{}).Return(nil, errors.New("error")).Once()
			},
			expected: Expected{nil, ErrFailedNamespaceRemoveMember},
		},
		{
			description: "refuses to remove the namespace owner",
			username:    "john_doe",
			namespace:   "namespace",
			requiredMocks: func() {
				user := &models.User{
					ID: "507f191e810c19729de860ea",
					UserData: models.UserData{
						Name:     "John Doe",
						Email:    "john.doe@test.com",
						Username: "john_doe",
					},
				}
				mock.On("UserResolve", ctx, store.UserUsernameResolver, "john_doe").Return(user, nil).Once()
				namespace := &models.Namespace{
					Name:     "namespace",
					Owner:    "507f191e810c19729de860ea",
					TenantID: "00000000-0000-0000-0000-000000000000",
					Members:  []models.Member{{ID: "507f191e810c19729de860ea", Role: "owner"}},
					Settings: &models.NamespaceSettings{
						SessionRecord: true,
					},
					CreatedAt: now,
				}
				mock.On("NamespaceResolve", ctx, store.NamespaceNameResolver, "namespace").Return(namespace, nil).Once()
			},
			expected: Expected{nil, ErrNamespaceRemoveOwner},
		},
		{
			description: "successfully remove member from the namespace",
			username:    "john_doe",
			namespace:   "namespace",
			requiredMocks: func() {
				user := &models.User{
					ID: "507f191e810c19729de860ea",
					UserData: models.UserData{
						Name:     "John Doe",
						Email:    "john.doe@test.com",
						Username: "john_doe",
					},
				}
				mock.On("UserResolve", ctx, store.UserUsernameResolver, "john_doe").Return(user, nil).Once()
				namespace := &models.Namespace{
					Name:     "namespace",
					Owner:    "608f191e810c19729de860eb",
					TenantID: "00000000-0000-0000-0000-000000000000",
					Members: []models.Member{
						{ID: "608f191e810c19729de860eb", Role: "owner"},
						{ID: "507f191e810c19729de860ea", Role: "operator"},
					},
					Settings: &models.NamespaceSettings{
						SessionRecord: true,
					},
					CreatedAt: now,
				}
				mock.On("NamespaceResolve", ctx, store.NamespaceNameResolver, "namespace").Return(namespace, nil).Once()
				mock.On("NamespaceDepartMember", ctx, scope.MustBounded("00000000-0000-0000-0000-000000000000"), "507f191e810c19729de860ea", store.MemberDeparture{}).Return(&models.MemberDeparted{}, nil).Once()
			},
			expected: Expected{&models.Namespace{
				Name:     "namespace",
				Owner:    "608f191e810c19729de860eb",
				TenantID: "00000000-0000-0000-0000-000000000000",
				Members: []models.Member{
					{ID: "608f191e810c19729de860eb", Role: "owner"},
					{ID: "507f191e810c19729de860ea", Role: "operator"},
				},
				Settings: &models.NamespaceSettings{
					SessionRecord: true,
				},
				CreatedAt: now,
			}, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			s := NewService(store.Store(mock), WithClock(clock.NewFixed(now)))
			ns, err := s.NamespaceRemoveMember(ctx, &inputs.MemberRemove{Username: tc.username, Namespace: tc.namespace})
			assert.Equal(t, tc.expected, Expected{ns, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestNamespaceLimit(t *testing.T) {
	type Expected struct {
		namespace *models.Namespace
		err       error
	}

	mock := new(mocks.MockStore)
	ctx := context.TODO()

	cases := []struct {
		description   string
		input         inputs.NamespaceLimit
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when could not find a namespace",
			input:       inputs.NamespaceLimit{Namespace: "namespace", MaxDevices: 1},
			requiredMocks: func() {
				mock.On("NamespaceResolve", ctx, store.NamespaceNameResolver, "namespace").Return(nil, errors.New("error")).Once()
			},
			expected: Expected{nil, ErrNamespaceNotFound},
		},
		{
			description: "fails when the store cannot save the limit",
			input:       inputs.NamespaceLimit{Namespace: "namespace", MaxDevices: 1},
			requiredMocks: func() {
				mock.On("NamespaceResolve", ctx, store.NamespaceNameResolver, "namespace").
					Return(&models.Namespace{Name: "namespace", MaxDevices: MaxNumberDevicesUnlimited}, nil).Once()
				mock.On("NamespaceUpdate", ctx, &models.Namespace{Name: "namespace", MaxDevices: 1}).Return(errors.New("error")).Once()
			},
			expected: Expected{nil, ErrFailedLimitNamespace},
		},
		{
			description: "sets the limit",
			input:       inputs.NamespaceLimit{Namespace: "namespace", MaxDevices: 1},
			requiredMocks: func() {
				mock.On("NamespaceResolve", ctx, store.NamespaceNameResolver, "namespace").
					Return(&models.Namespace{Name: "namespace", MaxDevices: MaxNumberDevicesUnlimited}, nil).Once()
				mock.On("NamespaceUpdate", ctx, &models.Namespace{Name: "namespace", MaxDevices: 1}).Return(nil).Once()
			},
			expected: Expected{&models.Namespace{Name: "namespace", MaxDevices: 1}, nil},
		},
		{
			description: "lifts the limit",
			input:       inputs.NamespaceLimit{Namespace: "namespace", MaxDevices: MaxNumberDevicesUnlimited},
			requiredMocks: func() {
				mock.On("NamespaceResolve", ctx, store.NamespaceNameResolver, "namespace").
					Return(&models.Namespace{Name: "namespace", MaxDevices: 1}, nil).Once()
				mock.On("NamespaceUpdate", ctx, &models.Namespace{Name: "namespace", MaxDevices: MaxNumberDevicesUnlimited}).Return(nil).Once()
			},
			expected: Expected{&models.Namespace{Name: "namespace", MaxDevices: MaxNumberDevicesUnlimited}, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			s := NewService(store.Store(mock))
			ns, err := s.NamespaceLimit(ctx, &tc.input)

			assert.Equal(t, tc.expected.namespace, ns)
			assert.ErrorIs(t, err, tc.expected.err)
		})
	}

	mock.AssertExpectations(t)
}

func TestNamespaceDelete(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

	mock := new(mocks.MockStore)
	cacheMock := cachemock.NewMockCache(t)

	ctx := context.TODO()

	cases := []struct {
		description   string
		namespace     string
		requiredMocks func()
		expected      error
	}{
		{
			description: "fails when could not find a namespace",
			namespace:   "namespace",
			requiredMocks: func() {
				mock.On("NamespaceResolve", ctx, store.NamespaceNameResolver, "namespace").Return(nil, errors.New("error")).Once()
			},
			expected: ErrNamespaceNotFound,
		},
		{
			description: "fails to delete the namespace",
			namespace:   "namespace",
			requiredMocks: func() {
				namespace := &models.Namespace{
					Name:     "namespace",
					Owner:    "507f191e810c19729de860ea",
					TenantID: "00000000-0000-0000-0000-000000000000",
					Members:  []models.Member{{ID: "507f191e810c19729de860ea", Role: "owner"}},
					Settings: &models.NamespaceSettings{
						SessionRecord: true,
					},
					MaxDevices: MaxNumberDevicesUnlimited,
					CreatedAt:  now,
				}
				mock.On("NamespaceResolve", ctx, store.NamespaceNameResolver, "namespace").Return(namespace, nil).Once()
				mock.On("DeviceList", ctx, scope.MustBounded("00000000-0000-0000-0000-000000000000"), store.DeviceAcceptableAsFalse).
					Return([]models.Device{{UID: "uid-1"}}, int64(1), nil).
					Once()
				mock.On("NamespaceDelete", ctx, namespace).Return(errors.New("error")).Once()
			},
			expected: ErrFailedDeleteNamespace,
		},
		{
			description: "fails without deleting when the namespace's devices cannot be listed",
			namespace:   "namespace",
			requiredMocks: func() {
				namespace := &models.Namespace{Name: "namespace", TenantID: "00000000-0000-0000-0000-000000000000"}
				mock.On("NamespaceResolve", ctx, store.NamespaceNameResolver, "namespace").Return(namespace, nil).Once()
				mock.On("DeviceList", ctx, scope.MustBounded("00000000-0000-0000-0000-000000000000"), store.DeviceAcceptableAsFalse).
					Return(nil, int64(0), errors.New("error")).
					Once()
			},
			expected: ErrFailedDeleteNamespace,
		},
		{
			description: "success to delete the namespace",
			namespace:   "namespace",
			requiredMocks: func() {
				namespace := &models.Namespace{
					Name:     "namespace",
					Owner:    "507f191e810c19729de860ea",
					TenantID: "00000000-0000-0000-0000-000000000000",
					Members:  []models.Member{{ID: "507f191e810c19729de860ea", Role: "owner"}},
					Settings: &models.NamespaceSettings{
						SessionRecord: true,
					},
					MaxDevices: MaxNumberDevicesUnlimited,
					CreatedAt:  now,
				}
				mock.On("NamespaceResolve", ctx, store.NamespaceNameResolver, "namespace").Return(namespace, nil).Once()
				mock.On("DeviceList", ctx, scope.MustBounded("00000000-0000-0000-0000-000000000000"), store.DeviceAcceptableAsFalse).
					Return([]models.Device{{UID: "uid-1"}, {UID: "uid-2"}}, int64(2), nil).
					Once()
				deleted := mock.On("NamespaceDelete", ctx, namespace).Return(nil).Once()
				cacheMock.On("Delete", ctx, "auth_device/uid-1").Return(nil).Once().NotBefore(deleted)
				cacheMock.On("Delete", ctx, "auth_device/uid-2").Return(nil).Once().NotBefore(deleted)
			},
			expected: nil,
		},
		{
			description: "deletes the namespace when evicting a device's authorization fails",
			namespace:   "namespace",
			requiredMocks: func() {
				namespace := &models.Namespace{Name: "namespace", TenantID: "00000000-0000-0000-0000-000000000000"}
				mock.On("NamespaceResolve", ctx, store.NamespaceNameResolver, "namespace").Return(namespace, nil).Once()
				mock.On("DeviceList", ctx, scope.MustBounded("00000000-0000-0000-0000-000000000000"), store.DeviceAcceptableAsFalse).
					Return([]models.Device{{UID: "uid-1"}, {UID: "uid-2"}}, int64(2), nil).
					Once()
				mock.On("NamespaceDelete", ctx, namespace).Return(nil).Once()
				cacheMock.On("Delete", ctx, "auth_device/uid-1").Return(errors.New("error")).Once()
				cacheMock.On("Delete", ctx, "auth_device/uid-2").Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			s := NewService(store.Store(mock), WithClock(clock.NewFixed(now)), WithCache(cacheMock))
			err := s.NamespaceDelete(ctx, &inputs.NamespaceDelete{Namespace: tc.namespace})
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestNamespaceList(t *testing.T) {
	type Expected struct {
		namespaces []models.Namespace
		err        error
	}

	mock := new(mocks.MockStore)
	ctx := context.TODO()

	cases := []struct {
		description   string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "returns an error when store fails",
			requiredMocks: func() {
				mock.On("NamespaceList", ctx).Return(nil, int64(0), errors.New("store error")).Once()
			},
			expected: Expected{nil, ErrFailedListNamespaces},
		},
		{
			description: "returns namespaces successfully",
			requiredMocks: func() {
				namespaces := []models.Namespace{
					{Name: "namespace1"},
					{Name: "namespace2"},
				}
				mock.On("NamespaceList", ctx).Return(namespaces, int64(len(namespaces)), nil).Once()
			},
			expected: Expected{
				[]models.Namespace{
					{Name: "namespace1"},
					{Name: "namespace2"},
				},
				nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			s := NewService(store.Store(mock))
			namespaces, err := s.NamespaceList(ctx)
			assert.Equal(t, tc.expected.namespaces, namespaces)
			assert.Equal(t, tc.expected.err, err)

			mock.AssertExpectations(t)
		})
	}
}

func TestNamespaceResolve(t *testing.T) {
	type Expected struct {
		namespace *models.Namespace
		err       error
	}

	mock := new(mocks.MockStore)
	ctx := context.TODO()

	cases := []struct {
		description   string
		resolver      NamespaceResolver
		value         string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when store returns error using name resolver",
			resolver:    NamespaceResolverName,
			value:       "namespace",
			requiredMocks: func() {
				mock.On("NamespaceResolve", ctx, store.NamespaceNameResolver, "namespace").
					Return(nil, errors.New("error")).Once()
			},
			expected: Expected{nil, ErrNamespaceNotFound},
		},
		{
			description: "succeeds using name resolver",
			resolver:    NamespaceResolverName,
			value:       "namespace",
			requiredMocks: func() {
				ns := &models.Namespace{Name: "namespace"}
				mock.On("NamespaceResolve", ctx, store.NamespaceNameResolver, "namespace").
					Return(ns, nil).Once()
			},
			expected: Expected{&models.Namespace{Name: "namespace"}, nil},
		},
		{
			description: "succeeds using tenant-id resolver",
			resolver:    NamespaceResolverTenantID,
			value:       "tenant-id",
			requiredMocks: func() {
				ns := &models.Namespace{TenantID: "tenant-id"}
				mock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant-id").
					Return(ns, nil).Once()
			},
			expected: Expected{&models.Namespace{TenantID: "tenant-id"}, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			s := NewService(store.Store(mock))
			ns, err := s.NamespaceResolve(ctx, tc.resolver, tc.value)

			assert.Equal(t, tc.expected.namespace, ns)
			assert.Equal(t, tc.expected.err, err)
		})
	}

	mock.AssertExpectations(t)
}
