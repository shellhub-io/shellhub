package services

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/envs"
	envmock "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	uuid_mocks "github.com/shellhub-io/shellhub/pkg/uuid/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListNamespaces(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	type Expected struct {
		namespaces []models.Namespace
		count      int
		err        error
	}

	cases := []struct {
		description   string
		req           *requests.NamespaceList
		ctx           context.Context
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fail when could not get the namespace list",
			req: &requests.NamespaceList{
				Paginator: query.Paginator{Page: 1, PerPage: 10},
				Filters:   query.Filters{},
			},
			ctx: ctx,
			requiredMocks: func() {
				mock.On("NamespaceList", ctx, query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, false).Return(nil, 0, errors.New("error")).Once()
			},
			expected: Expected{
				namespaces: nil,
				count:      0,
				err:        NewErrNamespaceList(errors.New("error")),
			},
		},
		{
			description: "success to get the namespace list",
			req: &requests.NamespaceList{
				Paginator: query.Paginator{Page: 1, PerPage: 10},
				Filters:   query.Filters{},
			},
			ctx: ctx,
			requiredMocks: func() {
				namespaces := []models.Namespace{
					{
						Name:     "group1",
						Owner:    "66ffe0745a82ba5c4fe842ac",
						TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
						Members: []models.Member{
							{
								ID:   "66ffe0745a82ba5c4fe842ac",
								Role: authorizer.RoleOwner,
							},
						},
					},
					{
						Name:     "group2",
						Owner:    "66ffe0745a82ba5c4fe842ac",
						TenantID: "a736a52b-5777-4f92-b0b8-e359bf48471i4",
						Members: []models.Member{
							{
								ID:   "66ffe0745a82ba5c4fe842ac",
								Role: authorizer.RoleOwner,
							},
							{
								ID:   "66ffe0232da6d319c9769afb",
								Role: authorizer.RoleObserver,
							},
						},
					},
				}

				user := &models.User{
					ID: "66ffe0745a82ba5c4fe842ac",
					UserData: models.UserData{
						Name:     "John Doe",
						Username: "john_doe",
						Email:    "john.doe@test.com",
					},
				}

				user1 := &models.User{
					ID: "66ffe0232da6d319c9769afb",
					UserData: models.UserData{
						Name:     "Jane Smith",
						Username: "jane_smith",
						Email:    "jane.smith@test.com",
					},
				}

				// TODO: Add mock to fillMembersData what will replace the three call to UserGetByID.
				mock.On("NamespaceList", ctx, query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, false).Return(namespaces, len(namespaces), nil).Once()
				mock.On("UserGetByID", ctx, "66ffe0745a82ba5c4fe842ac", false).Return(user, 0, nil).Once()
				mock.On("UserGetByID", ctx, "66ffe0232da6d319c9769afb", false).Return(user1, 0, nil).Once()
				mock.On("UserGetByID", ctx, "66ffe0745a82ba5c4fe842ac", false).Return(user, 0, nil).Once()
			},
			expected: Expected{
				namespaces: []models.Namespace{
					{
						Name:     "group1",
						Owner:    "66ffe0745a82ba5c4fe842ac",
						TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
						Members: []models.Member{
							{
								ID:    "66ffe0745a82ba5c4fe842ac",
								Email: "john.doe@test.com",
								Role:  authorizer.RoleOwner,
							},
						},
					},
					{
						Name:     "group2",
						Owner:    "66ffe0745a82ba5c4fe842ac",
						TenantID: "a736a52b-5777-4f92-b0b8-e359bf48471i4",
						Members: []models.Member{
							{
								ID:    "66ffe0745a82ba5c4fe842ac",
								Email: "john.doe@test.com",
								Role:  authorizer.RoleOwner,
							},
							{
								ID:    "66ffe0232da6d319c9769afb",
								Email: "jane.smith@test.com",
								Role:  authorizer.RoleObserver,
							},
						}},
				},
				count: 2,
				err:   nil,
			},
		},
	}

	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			nss, count, err := s.ListNamespaces(tc.ctx, tc.req)
			assert.Equal(t, tc.expected, Expected{nss, count, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestGetNamespace(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	type Expected struct {
		namespace *models.Namespace
		err       error
	}

	cases := []struct {
		description   string
		tenantID      string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when could not get the namespace",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			requiredMocks: func() {
				mock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(nil, errors.New("error")).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceNotFound("00000000-0000-4000-0000-000000000000", errors.New("error")),
			},
		},
		{
			description: "succeeds",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			requiredMocks: func() {
				mock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(
						&models.Namespace{
							Name:     "group1",
							Owner:    "66ffe21f76d5207a38a056d5",
							TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
							Members: []models.Member{
								{
									ID:   "66ffe21f76d5207a38a056d5",
									Role: authorizer.RoleOwner,
								},
							},
						},
						nil,
					).
					Once()
				mock.
					On("UserGetByID", ctx, "66ffe21f76d5207a38a056d5", false).
					Return(
						&models.User{
							ID: "66ffe21f76d5207a38a056d5",
							UserData: models.UserData{
								Name:     "John Doe",
								Username: "john_doe",
								Email:    "john.doe@test.com",
							},
						},
						0,
						nil,
					).
					Once()
			},
			expected: Expected{
				namespace: &models.Namespace{
					Name:     "group1",
					Owner:    "66ffe21f76d5207a38a056d5",
					TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
					Members: []models.Member{
						{
							ID:    "66ffe21f76d5207a38a056d5",
							Role:  authorizer.RoleOwner,
							Email: "john.doe@test.com",
						},
					},
				},
				err: nil,
			},
		},
	}

	s := NewService(mock, privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			ns, err := s.GetNamespace(ctx, tc.tenantID)
			assert.Equal(t, tc.expected, Expected{ns, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestSetMemberData(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	type Expected struct {
		members []models.Member
		err     error
	}

	cases := []struct {
		description   string
		members       []models.Member
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "success to fill member data",
			members: []models.Member{
				{
					ID:   "76a8bf2812dabc4fae5d1a42",
					Role: authorizer.RoleObserver,
				},
				{
					ID:   "95e4bdc9023a5f9b8d5b8c9a",
					Role: authorizer.RoleObserver,
				},
				{
					ID:   "b7c2fbb745fe9c6d7b648832",
					Role: authorizer.RoleObserver,
				},
				{
					ID:   "f8a1d4e8752f6a3b9c7b2d14",
					Role: authorizer.RoleOwner,
				},
			},
			requiredMocks: func() {
				mock.
					On("UserGetByID", ctx, "76a8bf2812dabc4fae5d1a42", false).
					Return(
						&models.User{
							ID: "76a8bf2812dabc4fae5d1a42",
							UserData: models.UserData{
								Username: "john_doe",
								Email:    "john.doe@example.com",
							},
						},
						0,
						nil,
					).
					Once()
				mock.
					On("UserGetByID", ctx, "95e4bdc9023a5f9b8d5b8c9a", false).
					Return(
						&models.User{
							ID: "95e4bdc9023a5f9b8d5b8c9a",
							UserData: models.UserData{
								Username: "jane_smith",
								Email:    "jane.smith@example.com",
							},
						},
						0,
						nil,
					).
					Once()
				mock.
					On("UserGetByID", ctx, "b7c2fbb745fe9c6d7b648832", false).
					Return(
						&models.User{
							ID: "b7c2fbb745fe9c6d7b648832",
							UserData: models.UserData{
								Username: "michael_brown",
								Email:    "michael.brown@example.com",
							},
						},
						0,
						nil,
					).
					Once()
				mock.
					On("UserGetByID", ctx, "f8a1d4e8752f6a3b9c7b2d14", false).
					Return(
						&models.User{
							ID: "f8a1d4e8752f6a3b9c7b2d14",
							UserData: models.UserData{
								Username: "sarah_connor",
								Email:    "sarah.connor@example.com",
							},
						},
						0,
						nil,
					).
					Once()
			},
			expected: Expected{
				members: []models.Member{
					{
						ID:    "76a8bf2812dabc4fae5d1a42",
						Email: "john.doe@example.com",
						Role:  authorizer.RoleObserver,
					},
					{
						ID:    "95e4bdc9023a5f9b8d5b8c9a",
						Email: "jane.smith@example.com",
						Role:  authorizer.RoleObserver,
					},
					{
						ID:    "b7c2fbb745fe9c6d7b648832",
						Email: "michael.brown@example.com",
						Role:  authorizer.RoleObserver,
					},
					{
						ID:    "f8a1d4e8752f6a3b9c7b2d14",
						Email: "sarah.connor@example.com",
						Role:  authorizer.RoleOwner,
					},
				},
				err: nil,
			},
		},
	}

	s := NewService(mock, privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			members, err := s.fillMembersData(ctx, tc.members)
			assert.Equal(t, tc.expected, Expected{members, err})
		})
	}
}

func TestCreateNamespace(t *testing.T) {
	storeMock := new(mocks.Store)
	clockMock := new(clockmock.Clock)
	clock.DefaultBackend = clockMock

	now := time.Now()
	clockMock.On("Now").Return(now)

	ctx := context.TODO()

	uuidMock := &uuid_mocks.Uuid{}
	uuid.DefaultBackend = uuidMock

	type Expected struct {
		ns  *models.Namespace
		err error
	}

	cases := []struct {
		description   string
		requiredMocks func()
		req           *requests.NamespaceCreate
		expected      Expected
	}{
		{
			description: "fails when store user get has no documents",
			req: &requests.NamespaceCreate{
				UserID:   "000000000000000000000000",
				Name:     "namespace",
				TenantID: "00000000-0000-4000-0000-000000000000",
			},
			requiredMocks: func() {
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(nil, 0, store.ErrNoDocuments).
					Once()
			},
			expected: Expected{
				ns:  nil,
				err: NewErrUserNotFound("000000000000000000000000", store.ErrNoDocuments),
			},
		},
		{
			description: "fails when user reachs the max namespaces",
			req: &requests.NamespaceCreate{
				UserID:   "000000000000000000000000",
				Name:     "namespace",
				TenantID: "00000000-0000-4000-0000-000000000000",
			},
			requiredMocks: func() {
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{ID: "000000000000000000000000", MaxNamespaces: 1}, 0, nil).
					Once()
				storeMock.
					On("UserGetInfo", ctx, "000000000000000000000000").
					Return(
						&models.UserInfo{
							OwnedNamespaces:      []models.Namespace{{}},
							AssociatedNamespaces: []models.Namespace{},
						},
						nil,
					).
					Once()
			},
			expected: Expected{
				ns:  nil,
				err: NewErrNamespaceLimitReached(1, nil),
			},
		},
		{
			description: "fails when a namespace already exists",
			req: &requests.NamespaceCreate{
				UserID:   "000000000000000000000000",
				Name:     "namespace",
				TenantID: "00000000-0000-4000-0000-000000000000",
			},
			requiredMocks: func() {
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{ID: "000000000000000000000000"}, 0, nil).
					Once()
				storeMock.
					On("NamespaceGetByName", ctx, "namespace").
					Return(&models.Namespace{Name: "namespace"}, nil).
					Once()
			},
			expected: Expected{
				ns:  nil,
				err: NewErrNamespaceDuplicated(nil),
			},
		},
		{
			description: "fails retrieve namespace fails without ErrNoDocuments",
			req: &requests.NamespaceCreate{
				UserID:   "000000000000000000000000",
				Name:     "namespace",
				TenantID: "00000000-0000-4000-0000-000000000000",
			},
			requiredMocks: func() {
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{ID: "000000000000000000000000"}, 0, nil).
					Once()
				storeMock.
					On("NamespaceGetByName", ctx, "namespace").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: Expected{
				ns:  nil,
				err: NewErrNamespaceDuplicated(errors.New("error")),
			},
		},
		{
			description: "fails when store namespace create fails",
			req: &requests.NamespaceCreate{
				UserID:   "000000000000000000000000",
				Name:     "namespace",
				TenantID: "00000000-0000-4000-0000-000000000000",
			},
			requiredMocks: func() {
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{ID: "000000000000000000000000"}, 0, nil).
					Once()
				storeMock.
					On("NamespaceGetByName", ctx, "namespace").
					Return(nil, store.ErrNoDocuments).
					Once()
				// envs.IsCommunity = true
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("false").
					Once()
				// --
				// envs.IsCloud = false
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				// --
				storeMock.
					On(
						"NamespaceCreate",
						ctx,
						&models.Namespace{
							TenantID: "00000000-0000-4000-0000-000000000000",
							Name:     "namespace",
							Owner:    "000000000000000000000000",
							Members: []models.Member{
								{
									ID:      "000000000000000000000000",
									Role:    authorizer.RoleOwner,
									Status:  models.MemberStatusAccepted,
									AddedAt: now,
								},
							},
							Settings: &models.NamespaceSettings{
								SessionRecord:          true,
								ConnectionAnnouncement: models.DefaultAnnouncementMessage,
							},
							MaxDevices: -1,
						},
					).
					Return(nil, errors.New("error")).
					Once()
			},
			expected: Expected{
				ns:  nil,
				err: NewErrNamespaceCreateStore(errors.New("error")),
			},
		},
		{
			description: "succeeds to create a namespace",
			req: &requests.NamespaceCreate{
				UserID:   "000000000000000000000000",
				Name:     "namespace",
				TenantID: "00000000-0000-4000-0000-000000000000",
			},
			requiredMocks: func() {
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{ID: "000000000000000000000000"}, 0, nil).
					Once()
				storeMock.
					On("NamespaceGetByName", ctx, "namespace").
					Return(nil, store.ErrNoDocuments).
					Once()
				// envs.IsCommunity = true
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("false").
					Once()
				// --
				// envs.IsCloud = false
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				// --
				storeMock.
					On(
						"NamespaceCreate",
						ctx,
						&models.Namespace{
							TenantID: "00000000-0000-4000-0000-000000000000",
							Name:     "namespace",
							Owner:    "000000000000000000000000",
							Members: []models.Member{
								{
									ID:      "000000000000000000000000",
									Role:    authorizer.RoleOwner,
									Status:  models.MemberStatusAccepted,
									AddedAt: now,
								},
							},
							Settings: &models.NamespaceSettings{
								SessionRecord:          true,
								ConnectionAnnouncement: models.DefaultAnnouncementMessage,
							},
							MaxDevices: -1,
						},
					).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-4000-0000-000000000000",
							Name:     "namespace",
							Owner:    "000000000000000000000000",
							Members: []models.Member{
								{
									ID:      "000000000000000000000000",
									Role:    authorizer.RoleOwner,
									Status:  models.MemberStatusAccepted,
									AddedAt: now,
								},
							},
							Settings: &models.NamespaceSettings{
								SessionRecord:          true,
								ConnectionAnnouncement: models.DefaultAnnouncementMessage,
							},
							MaxDevices: -1,
						},
						nil,
					).
					Once()
			},
			expected: Expected{
				ns: &models.Namespace{
					TenantID: "00000000-0000-4000-0000-000000000000",
					Name:     "namespace",
					Owner:    "000000000000000000000000",
					Members: []models.Member{
						{
							ID:      "000000000000000000000000",
							Role:    authorizer.RoleOwner,
							Status:  models.MemberStatusAccepted,
							AddedAt: now,
						},
					},
					Settings: &models.NamespaceSettings{
						SessionRecord:          true,
						ConnectionAnnouncement: models.DefaultAnnouncementMessage,
					},
					MaxDevices: -1,
				},
				err: nil,
			},
		},
		{
			description: "succeeds to create a namespace-:-without tenant id",
			req: &requests.NamespaceCreate{
				UserID:   "000000000000000000000000",
				Name:     "namespace",
				TenantID: "",
			},
			requiredMocks: func() {
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{ID: "000000000000000000000000"}, 0, nil).
					Once()
				storeMock.
					On("NamespaceGetByName", ctx, "namespace").
					Return(nil, store.ErrNoDocuments).
					Once()
				// envs.IsCommunity = true
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("false").
					Once()
				// --
				// envs.IsCloud = false
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				// --
				uuidMock.
					On("Generate").
					Return("4de9253f-4a2a-49e7-a748-26e7a009bd2e").
					Once()
				storeMock.
					On(
						"NamespaceCreate",
						ctx,
						&models.Namespace{
							TenantID: "4de9253f-4a2a-49e7-a748-26e7a009bd2e",
							Name:     "namespace",
							Owner:    "000000000000000000000000",
							Members: []models.Member{
								{
									ID:      "000000000000000000000000",
									Role:    authorizer.RoleOwner,
									Status:  models.MemberStatusAccepted,
									AddedAt: now,
								},
							},
							Settings: &models.NamespaceSettings{
								SessionRecord:          true,
								ConnectionAnnouncement: models.DefaultAnnouncementMessage,
							},
							MaxDevices: -1,
						},
					).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-4000-0000-000000000000",
							Name:     "namespace",
							Owner:    "000000000000000000000000",
							Members: []models.Member{
								{
									ID:      "000000000000000000000000",
									Role:    authorizer.RoleOwner,
									Status:  models.MemberStatusAccepted,
									AddedAt: now,
								},
							},
							Settings: &models.NamespaceSettings{
								SessionRecord:          true,
								ConnectionAnnouncement: models.DefaultAnnouncementMessage,
							},
							MaxDevices: -1,
						},
						nil,
					).
					Once()
			},
			expected: Expected{
				ns: &models.Namespace{
					TenantID: "4de9253f-4a2a-49e7-a748-26e7a009bd2e",
					Name:     "namespace",
					Owner:    "000000000000000000000000",
					Members: []models.Member{
						{
							ID:      "000000000000000000000000",
							Role:    authorizer.RoleOwner,
							Status:  models.MemberStatusAccepted,
							AddedAt: now,
						},
					},
					Settings: &models.NamespaceSettings{
						SessionRecord:          true,
						ConnectionAnnouncement: models.DefaultAnnouncementMessage,
					},
					MaxDevices: -1,
				},
				err: nil,
			},
		},
		{
			description: "succeeds to create a namespace-:-env=cloud",
			req: &requests.NamespaceCreate{
				UserID:   "000000000000000000000000",
				Name:     "namespace",
				TenantID: "00000000-0000-4000-0000-000000000000",
			},
			requiredMocks: func() {
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{ID: "000000000000000000000000"}, 0, nil).
					Once()
				storeMock.
					On("NamespaceGetByName", ctx, "namespace").
					Return(nil, store.ErrNoDocuments).
					Once()
				// envs.IsCommunity = false
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("true").
					Once()
				// --
				// envs.IsCloud = true
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				// --
				storeMock.
					On(
						"NamespaceCreate",
						ctx,
						&models.Namespace{
							TenantID: "00000000-0000-4000-0000-000000000000",
							Name:     "namespace",
							Owner:    "000000000000000000000000",
							Members: []models.Member{
								{
									ID:      "000000000000000000000000",
									Role:    authorizer.RoleOwner,
									Status:  models.MemberStatusAccepted,
									AddedAt: now,
								},
							},
							Settings: &models.NamespaceSettings{
								SessionRecord:          true,
								ConnectionAnnouncement: "",
							},
							MaxDevices: 3,
						},
					).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-4000-0000-000000000000",
							Name:     "namespace",
							Owner:    "000000000000000000000000",
							Members: []models.Member{
								{
									ID:      "000000000000000000000000",
									Role:    authorizer.RoleOwner,
									Status:  models.MemberStatusAccepted,
									AddedAt: now,
								},
							},
							Settings: &models.NamespaceSettings{
								SessionRecord:          true,
								ConnectionAnnouncement: models.DefaultAnnouncementMessage,
							},
							MaxDevices: -1,
						},
						nil,
					).
					Once()
			},
			expected: Expected{
				ns: &models.Namespace{
					TenantID: "00000000-0000-4000-0000-000000000000",
					Name:     "namespace",
					Owner:    "000000000000000000000000",
					Members: []models.Member{
						{
							ID:      "000000000000000000000000",
							Role:    authorizer.RoleOwner,
							Status:  models.MemberStatusAccepted,
							AddedAt: now,
						},
					},
					Settings: &models.NamespaceSettings{
						SessionRecord:          true,
						ConnectionAnnouncement: "",
					},
					MaxDevices: 3,
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)
			returnedNamespace, err := service.CreateNamespace(ctx, tc.req)

			assert.Equal(t, tc.expected, Expected{returnedNamespace, err})

			storeMock.AssertExpectations(t)
		})
	}
}

func TestEditNamespace(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	type Expected struct {
		namespace *models.Namespace
		err       error
	}

	cases := []struct {
		description   string
		requiredMocks func()
		tenantID      string
		namespaceName string
		expected      Expected
	}{
		{
			description:   "fails when namespace does not exist",
			tenantID:      "xxxxx",
			namespaceName: "newname",
			requiredMocks: func() {
				mock.On("NamespaceEdit", ctx, "xxxxx", &models.NamespaceChanges{Name: "newname"}).
					Return(store.ErrNoDocuments).
					Once()
			},
			expected: Expected{
				nil,
				NewErrNamespaceNotFound("xxxxx", store.ErrNoDocuments),
			},
		},
		{
			description:   "fails when the store namespace rename fails",
			tenantID:      "xxxxx",
			namespaceName: "newname",
			requiredMocks: func() {
				mock.On("NamespaceEdit", ctx, "xxxxx", &models.NamespaceChanges{Name: "newname"}).
					Return(errors.New("error")).
					Once()
			},
			expected: Expected{
				nil,
				errors.New("error"),
			},
		},
		{
			description:   "succeeds changing the name to lowercase",
			namespaceName: "newName",
			tenantID:      "xxxxx",
			requiredMocks: func() {
				mock.On("NamespaceEdit", ctx, "xxxxx", &models.NamespaceChanges{Name: "newname"}).
					Return(nil).
					Once()

				namespace := &models.Namespace{
					TenantID: "xxxxx",
					Name:     "newname",
				}

				mock.On("NamespaceGet", ctx, "xxxxx", true).
					Return(namespace, nil).
					Once()
			},
			expected: Expected{
				&models.Namespace{
					TenantID: "xxxxx",
					Name:     "newname",
				},
				nil,
			},
		},
		{
			description:   "succeeds",
			namespaceName: "newname",
			tenantID:      "xxxxx",
			requiredMocks: func() {
				mock.On("NamespaceEdit", ctx, "xxxxx", &models.NamespaceChanges{Name: "newname"}).
					Return(nil).
					Once()

				namespace := &models.Namespace{
					TenantID: "xxxxx",
					Name:     "newname",
				}

				mock.On("NamespaceGet", ctx, "xxxxx", true).
					Return(namespace, nil).
					Once()
			},
			expected: Expected{
				&models.Namespace{
					TenantID: "xxxxx",
					Name:     "newname",
				},
				nil,
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

			req := &requests.NamespaceEdit{
				TenantParam: requests.TenantParam{Tenant: tc.tenantID},
				Name:        tc.namespaceName,
			}
			namespace, err := service.EditNamespace(ctx, req)

			assert.Equal(t, tc.expected, Expected{namespace, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestDeleteNamespace(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	cases := []struct {
		description   string
		namespace     *models.Namespace
		requiredMocks func(namespace *models.Namespace)
		expected      error
	}{
		{
			description: "fails when namespace does not exist",
			namespace:   &models.Namespace{Name: "oldname", Owner: "ID1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Members: []models.Member{{ID: "user1", Role: authorizer.RoleOwner}}},
			requiredMocks: func(namespace *models.Namespace) {
				mock.On("NamespaceGet", ctx, namespace.TenantID, true).Return(nil, errors.New("error")).Once()
			},
			expected: NewErrNamespaceNotFound("a736a52b-5777-4f92-b0b8-e359bf484713", errors.New("error")),
		},
		{
			description: "fails when store delete fails",
			namespace:   &models.Namespace{Name: "oldname", Owner: "ID1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Members: []models.Member{{ID: "user1", Role: authorizer.RoleOwner}}},
			requiredMocks: func(namespace *models.Namespace) {
				mock.On("NamespaceGet", ctx, namespace.TenantID, true).Return(namespace, nil).Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Once()
				envMock.On("Get", "SHELLHUB_BILLING").Return("false").Once()
				mock.On("NamespaceDelete", ctx, namespace.TenantID).Return(errors.New("error")).Once()
			},
			expected: errors.New("error"),
		},
		{
			description: "succeeds",
			namespace:   &models.Namespace{Name: "oldname", Owner: "ID1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Members: []models.Member{{ID: "user1", Role: authorizer.RoleOwner}}},
			requiredMocks: func(namespace *models.Namespace) {
				mock.On("NamespaceGet", ctx, namespace.TenantID, true).Return(namespace, nil).Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Once()
				envMock.On("Get", "SHELLHUB_BILLING").Return("false").Once()
				mock.On("NamespaceDelete", ctx, namespace.TenantID).Return(nil).Once()
			},
			expected: nil,
		},
		{
			description: "reports delete",
			namespace:   &models.Namespace{Name: "oldname", Owner: "ID1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Members: []models.Member{{ID: "user1", Role: authorizer.RoleOwner}}},
			requiredMocks: func(namespace *models.Namespace) {
				user1 := &models.User{
					UserData: models.UserData{
						Name:     "user1",
						Username: "user1",
						Email:    "user1@email.com",
					},
					ID: "ID1",
				}

				ns := &models.Namespace{
					TenantID: namespace.TenantID,
					Owner:    user1.ID,
					Members: []models.Member{
						{ID: user1.ID, Role: authorizer.RoleOwner},
					},
					Billing: &models.Billing{
						Active: true,
					},
					MaxDevices: -1,
				}
				mock.On("NamespaceGet", ctx, namespace.TenantID, true).Return(ns, nil).Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return(strconv.FormatBool(true)).Once()
				envMock.On("Get", "SHELLHUB_BILLING").Return(strconv.FormatBool(true)).Once()
				clientMock.On("ReportDelete", ns).Return(200, nil).Once()
				mock.On("NamespaceDelete", ctx, namespace.TenantID).Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks(tc.namespace)

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock)
			err := service.DeleteNamespace(ctx, tc.namespace.TenantID)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestAddNamespaceMember(t *testing.T) {
	type Expected struct {
		namespace *models.Namespace
		err       error
	}

	storeMock := new(mocks.Store)
	clockMock := new(clockmock.Clock)
	clock.DefaultBackend = clockMock

	now := time.Now()
	clockMock.On("Now").Return(now)

	cases := []struct {
		description   string
		req           *requests.NamespaceAddMember
		requiredMocks func(context.Context)
		expected      Expected
	}{
		{
			description: "fails when the namespace was not found",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(nil, ErrNamespaceNotFound).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceNotFound("00000000-0000-4000-0000-000000000000", ErrNamespaceNotFound),
			},
		},
		{
			description: "fails when the active member was not found",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members:  []models.Member{},
					}, nil).
					Once()
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(nil, 0, ErrUserNotFound).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrUserNotFound("000000000000000000000000", ErrUserNotFound),
			},
		},
		{
			description: "fails when the active member is not on the namespace",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members:  []models.Member{},
					}, nil).
					Once()
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, 0, nil).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceMemberNotFound("000000000000000000000000", nil),
			},
		},
		{
			description: "fails when the passive role's is owner",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleOwner,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:     "000000000000000000000000",
								Role:   authorizer.RoleOperator,
								Status: models.MemberStatusAccepted,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, 0, nil).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrRoleInvalid(),
			},
		},
		{
			description: "fails when the active member's role cannot act over passive member's role",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleAdministrator,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:     "000000000000000000000000",
								Role:   authorizer.RoleOperator,
								Status: models.MemberStatusAccepted,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, 0, nil).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrRoleInvalid(),
			},
		},
		{
			description: "fails when passive member was not found",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:     "000000000000000000000000",
								Role:   authorizer.RoleOwner,
								Status: models.MemberStatusAccepted,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, 0, nil).
					Once()
				storeMock.
					On("UserGetByEmail", ctx, "john.doe@test.com").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrUserNotFound("john.doe@test.com", errors.New("error")),
			},
		},
		{
			description: "fails when the member is duplicated without 'pending' status and expiration date not reached",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:     "000000000000000000000000",
								Role:   authorizer.RoleOwner,
								Status: models.MemberStatusAccepted,
							},
							{
								ID:     "000000000000000000000001",
								Role:   authorizer.RoleAdministrator,
								Status: models.MemberStatusAccepted,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, 0, nil).
					Once()
				storeMock.
					On("UserGetByEmail", ctx, "john.doe@test.com").
					Return(&models.User{
						ID:       "000000000000000000000001",
						UserData: models.UserData{Username: "john_doe"},
					}, nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceMemberDuplicated("000000000000000000000001", nil),
			},
		},
		{
			description: "[cloud] fails when the member is duplicated without 'pending' status and expiration date not reached",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:     "000000000000000000000000",
								Role:   authorizer.RoleOwner,
								Status: models.MemberStatusAccepted,
							},
							{
								ID:     "000000000000000000000001",
								Role:   authorizer.RoleAdministrator,
								Status: models.MemberStatusAccepted,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, 0, nil).
					Once()
				storeMock.
					On("UserGetByEmail", ctx, "john.doe@test.com").
					Return(&models.User{
						ID:       "000000000000000000000001",
						UserData: models.UserData{Username: "john_doe"},
					}, nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceMemberDuplicated("000000000000000000000001", nil),
			},
		},
		{
			description: "[cloud] fails when the member is duplicated with 'pending' status and expiration date not reached",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:     "000000000000000000000000",
								Role:   authorizer.RoleOwner,
								Status: models.MemberStatusAccepted,
							},
							{
								ID:        "000000000000000000000001",
								Role:      authorizer.RoleAdministrator,
								Status:    models.MemberStatusPending,
								ExpiresAt: time.Now().Add(7 * (24 * time.Hour)),
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, 0, nil).
					Once()
				storeMock.
					On("UserGetByEmail", ctx, "john.doe@test.com").
					Return(&models.User{
						ID:       "000000000000000000000001",
						UserData: models.UserData{Username: "john_doe"},
					}, nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceMemberDuplicated("000000000000000000000001", nil),
			},
		},
		{
			description: "[cloud] succeeds to resend the invite",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:     "000000000000000000000000",
								Role:   authorizer.RoleOwner,
								Status: models.MemberStatusAccepted,
							},
							{
								ID:        "000000000000000000000001",
								Role:      authorizer.RoleAdministrator,
								Status:    models.MemberStatusPending,
								ExpiresAt: time.Date(2023, 01, 01, 12, 00, 00, 00, time.UTC),
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, 0, nil).
					Once()
				storeMock.
					On("UserGetByEmail", ctx, "john.doe@test.com").
					Return(&models.User{
						ID:       "000000000000000000000001",
						UserData: models.UserData{Username: "john_doe"},
					}, nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				storeMock.
					On("WithTransaction", ctx, mock.Anything).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-4000-0000-000000000000",
							Name:     "namespace",
							Owner:    "000000000000000000000000",
							Members: []models.Member{
								{
									ID:     "000000000000000000000000",
									Role:   authorizer.RoleOwner,
									Status: models.MemberStatusAccepted,
								},
								{
									ID:     "000000000000000000000001",
									Role:   authorizer.RoleObserver,
									Status: models.MemberStatusPending,
								},
							},
						},
						nil,
					).
					Once()
			},
			expected: Expected{
				namespace: &models.Namespace{
					TenantID: "00000000-0000-4000-0000-000000000000",
					Name:     "namespace",
					Owner:    "000000000000000000000000",
					Members: []models.Member{
						{
							ID:     "000000000000000000000000",
							Role:   authorizer.RoleOwner,
							Status: models.MemberStatusAccepted,
						},
						{
							ID:     "000000000000000000000001",
							Role:   authorizer.RoleObserver,
							Status: models.MemberStatusPending,
						},
					},
				},
				err: nil,
			},
		},
		{
			description: "fails when cannot add the member",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:     "000000000000000000000000",
								Role:   authorizer.RoleOwner,
								Status: models.MemberStatusAccepted,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, 0, nil).
					Once()
				storeMock.
					On("UserGetByEmail", ctx, "john.doe@test.com").
					Return(&models.User{
						ID:       "000000000000000000000001",
						UserData: models.UserData{Username: "john_doe"},
					}, nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				storeMock.
					On("WithTransaction", ctx, mock.Anything).
					Return(errors.New("error")).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       errors.New("error"),
			},
		},
		{
			description: "succeeds",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:     "000000000000000000000000",
								Role:   authorizer.RoleOwner,
								Status: models.MemberStatusAccepted,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, 0, nil).
					Once()
				storeMock.
					On("UserGetByEmail", ctx, "john.doe@test.com").
					Return(&models.User{
						ID:       "000000000000000000000001",
						UserData: models.UserData{Username: "john_doe"},
					}, nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				storeMock.
					On("WithTransaction", ctx, mock.Anything).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:     "000000000000000000000000",
								Role:   authorizer.RoleOwner,
								Status: models.MemberStatusAccepted,
							},
							{
								ID:     "000000000000000000000000",
								Role:   authorizer.RoleObserver,
								Status: models.MemberStatusAccepted,
							},
						},
					}, nil).
					Once()
			},
			expected: Expected{
				namespace: &models.Namespace{
					TenantID: "00000000-0000-4000-0000-000000000000",
					Name:     "namespace",
					Owner:    "000000000000000000000000",
					Members: []models.Member{
						{
							ID:     "000000000000000000000000",
							Role:   authorizer.RoleOwner,
							Status: models.MemberStatusAccepted,
						},
						{
							ID:     "000000000000000000000000",
							Role:   authorizer.RoleObserver,
							Status: models.MemberStatusAccepted,
						},
					},
				},
				err: nil,
			},
		},
	}

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.TODO()
			tc.requiredMocks(ctx)
			ns, err := s.AddNamespaceMember(ctx, tc.req)
			assert.Equal(t, tc.expected, Expected{ns, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestService_addMember(t *testing.T) {
	envMock = new(envmock.Backend)
	clockMock := new(clockmock.Clock)
	storeMock := new(mocks.Store)

	envs.DefaultBackend = envMock
	clock.DefaultBackend = clockMock

	now := time.Now()
	clockMock.On("Now").Return(now)

	cases := []struct {
		description   string
		memberID      string
		req           *requests.NamespaceAddMember
		requiredMocks func(context.Context)
		expected      error
	}{
		{
			description: "fails cannot add the member",
			memberID:    "000000000000000000000000",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				storeMock.
					On("NamespaceAddMember", ctx, "00000000-0000-4000-0000-000000000000", &models.Member{ID: "000000000000000000000000", Role: authorizer.RoleObserver, Status: models.MemberStatusAccepted, AddedAt: now, ExpiresAt: time.Time{}}).
					Return(errors.New("error")).
					Once()
			},
			expected: errors.New("error"),
		},
		{
			description: "succeeds",
			memberID:    "000000000000000000000000",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				storeMock.
					On("NamespaceAddMember", ctx, "00000000-0000-4000-0000-000000000000", &models.Member{ID: "000000000000000000000000", Role: authorizer.RoleObserver, Status: models.MemberStatusAccepted, AddedAt: now, ExpiresAt: time.Time{}}).
					Return(nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
			},
			expected: nil,
		},
		{
			description: "[cloud] fails cannot add the member",
			memberID:    "000000000000000000000000",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				storeMock.
					On("NamespaceAddMember", ctx, "00000000-0000-4000-0000-000000000000", &models.Member{ID: "000000000000000000000000", Role: authorizer.RoleObserver, Status: models.MemberStatusPending, AddedAt: now, ExpiresAt: now.Add(7 * (24 * time.Hour))}).
					Return(errors.New("error")).
					Once()
			},
			expected: errors.New("error"),
		},
		{
			description: "[cloud] fails cannot send the invite",
			memberID:    "000000000000000000000000",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				storeMock.
					On("NamespaceAddMember", ctx, "00000000-0000-4000-0000-000000000000", &models.Member{ID: "000000000000000000000000", Role: authorizer.RoleObserver, Status: models.MemberStatusPending, AddedAt: now, ExpiresAt: now.Add(7 * (24 * time.Hour))}).
					Return(nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				clientMock.
					On("InviteMember", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000000", "localhost").
					Return(errors.New("error")).
					Once()
			},
			expected: errors.New("error"),
		},
		{
			description: "[cloud] succeeds",
			memberID:    "000000000000000000000000",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				UserID:       "000000000000000000000000",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberEmail:  "john.doe@test.com",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				storeMock.
					On("NamespaceAddMember", ctx, "00000000-0000-4000-0000-000000000000", &models.Member{ID: "000000000000000000000000", Role: authorizer.RoleObserver, Status: models.MemberStatusPending, AddedAt: now, ExpiresAt: now.Add(7 * (24 * time.Hour))}).
					Return(nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				clientMock.
					On("InviteMember", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000000", "localhost").
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	s := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.Background()
			tc.requiredMocks(ctx)

			cb := s.addMember(tc.memberID, tc.req)
			assert.Equal(tt, tc.expected, cb(ctx))

			storeMock.AssertExpectations(tt)
			envMock.AssertExpectations(tt)
		})
	}
}

func TestService_resendMemberInvite(t *testing.T) {
	envMock = new(envmock.Backend)
	storeMock := new(mocks.Store)
	clockMock := new(clockmock.Clock)

	envs.DefaultBackend = envMock
	clock.DefaultBackend = clockMock

	now := time.Now()
	clockMock.On("Now").Return(now)

	cases := []struct {
		description   string
		memberID      string
		req           *requests.NamespaceAddMember
		requiredMocks func(context.Context)
		expected      error
	}{
		{
			description: "fails cannot update the member",
			memberID:    "000000000000000000000000",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				expiresAt := now.Add(7 * (24 * time.Hour))
				storeMock.
					On("NamespaceUpdateMember", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000000", &models.MemberChanges{Role: authorizer.RoleObserver, ExpiresAt: &expiresAt}).
					Return(errors.New("error")).
					Once()
			},
			expected: errors.New("error"),
		},
		{
			description: "fails when cannot send the invite",
			memberID:    "000000000000000000000000",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				expiresAt := now.Add(7 * (24 * time.Hour))
				storeMock.
					On("NamespaceUpdateMember", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000000", &models.MemberChanges{Role: authorizer.RoleObserver, ExpiresAt: &expiresAt}).
					Return(nil).
					Once()
				clientMock.
					On("InviteMember", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000000", "localhost").
					Return(errors.New("error")).
					Once()
			},
			expected: errors.New("error"),
		},
		{
			description: "succeeds",
			memberID:    "000000000000000000000000",
			req: &requests.NamespaceAddMember{
				FowardedHost: "localhost",
				TenantID:     "00000000-0000-4000-0000-000000000000",
				MemberRole:   authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				expiresAt := now.Add(7 * (24 * time.Hour))
				storeMock.
					On("NamespaceUpdateMember", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000000", &models.MemberChanges{Role: authorizer.RoleObserver, ExpiresAt: &expiresAt}).
					Return(nil).
					Once()
				clientMock.
					On("InviteMember", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000000", "localhost").
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	s := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.Background()
			tc.requiredMocks(ctx)

			cb := s.resendMemberInvite(tc.memberID, tc.req)
			assert.Equal(tt, tc.expected, cb(ctx))

			envMock.AssertExpectations(tt)
			storeMock.AssertExpectations(tt)
			clockMock.AssertExpectations(tt)
		})
	}
}

func TestUpdateNamespaceMember(t *testing.T) {
	storeMock := new(mocks.Store)

	cases := []struct {
		description   string
		req           *requests.NamespaceUpdateMember
		requiredMocks func(context.Context)
		expected      error
	}{
		{
			description: "fails when the namespace was not found",
			req: &requests.NamespaceUpdateMember{
				UserID:     "000000000000000000000000",
				TenantID:   "00000000-0000-4000-0000-000000000000",
				MemberID:   "000000000000000000000001",
				MemberRole: authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(nil, ErrNamespaceNotFound).
					Once()
			},
			expected: NewErrNamespaceNotFound("00000000-0000-4000-0000-000000000000", ErrNamespaceNotFound),
		},
		{
			description: "fails when the active member was not found",
			req: &requests.NamespaceUpdateMember{
				UserID:     "000000000000000000000000",
				TenantID:   "00000000-0000-4000-0000-000000000000",
				MemberID:   "000000000000000000000001",
				MemberRole: authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members:  []models.Member{},
					}, nil).
					Once()
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(nil, 0, ErrUserNotFound).
					Once()
			},
			expected: NewErrUserNotFound("000000000000000000000000", ErrUserNotFound),
		},
		{
			description: "fails when the active member is not on the namespace",
			req: &requests.NamespaceUpdateMember{
				UserID:     "000000000000000000000000",
				TenantID:   "00000000-0000-4000-0000-000000000000",
				MemberID:   "000000000000000000000001",
				MemberRole: authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members:  []models.Member{},
					}, nil).
					Once()
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, 0, nil).
					Once()
			},
			expected: NewErrNamespaceMemberNotFound("000000000000000000000000", nil),
		},
		{
			description: "fails when the passive member is not on the namespace",
			req: &requests.NamespaceUpdateMember{
				UserID:     "000000000000000000000000",
				TenantID:   "00000000-0000-4000-0000-000000000000",
				MemberID:   "000000000000000000000001",
				MemberRole: authorizer.RoleObserver,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOwner,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, 0, nil).
					Once()
			},
			expected: NewErrNamespaceMemberNotFound("000000000000000000000001", nil),
		},
		{
			description: "fails when the passive role's is owner",
			req: &requests.NamespaceUpdateMember{
				UserID:     "000000000000000000000000",
				TenantID:   "00000000-0000-4000-0000-000000000000",
				MemberID:   "000000000000000000000001",
				MemberRole: authorizer.RoleOwner,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOwner,
							},
							{
								ID:   "000000000000000000000001",
								Role: authorizer.RoleOwner,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, 0, nil).
					Once()
			},
			expected: NewErrRoleInvalid(),
		},
		{
			description: "fails when the active member's role cannot act over passive member's role",
			req: &requests.NamespaceUpdateMember{
				UserID:     "000000000000000000000000",
				TenantID:   "00000000-0000-4000-0000-000000000000",
				MemberID:   "000000000000000000000001",
				MemberRole: authorizer.RoleAdministrator,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOperator,
							},
							{
								ID:   "000000000000000000000001",
								Role: authorizer.RoleAdministrator,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, 0, nil).
					Once()
			},
			expected: NewErrRoleInvalid(),
		},
		{
			description: "fails when cannot update the member",
			req: &requests.NamespaceUpdateMember{
				UserID:     "000000000000000000000000",
				TenantID:   "00000000-0000-4000-0000-000000000000",
				MemberID:   "000000000000000000000001",
				MemberRole: authorizer.RoleAdministrator,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOwner,
							},
							{
								ID:   "000000000000000000000001",
								Role: authorizer.RoleAdministrator,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, 0, nil).
					Once()
				storeMock.
					On("NamespaceUpdateMember", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000001", &models.MemberChanges{Role: authorizer.RoleAdministrator}).
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.TODO()
			tc.requiredMocks(ctx)
			err := s.UpdateNamespaceMember(ctx, tc.req)
			assert.Equal(t, tc.expected, err)
		})
	}
	storeMock.AssertExpectations(t)
}

func TestRemoveNamespaceMember(t *testing.T) {
	type Expected struct {
		namespace *models.Namespace
		err       error
	}

	storeMock := new(mocks.Store)

	cases := []struct {
		description   string
		req           *requests.NamespaceRemoveMember
		requiredMocks func(context.Context)
		expected      Expected
	}{
		{
			description: "fails when the namespace was not found",
			req: &requests.NamespaceRemoveMember{
				UserID:   "000000000000000000000000",
				TenantID: "00000000-0000-4000-0000-000000000000",
				MemberID: "000000000000000000000001",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(nil, ErrNamespaceNotFound).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceNotFound("00000000-0000-4000-0000-000000000000", ErrNamespaceNotFound),
			},
		},
		{
			description: "fails when the active member was not found",
			req: &requests.NamespaceRemoveMember{
				UserID:   "000000000000000000000000",
				TenantID: "00000000-0000-4000-0000-000000000000",
				MemberID: "000000000000000000000001",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members:  []models.Member{},
					}, nil).
					Once()
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(nil, 0, ErrUserNotFound).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrUserNotFound("000000000000000000000000", ErrUserNotFound),
			},
		},
		{
			description: "fails when the active member is not on the namespace",
			req: &requests.NamespaceRemoveMember{
				UserID:   "000000000000000000000000",
				TenantID: "00000000-0000-4000-0000-000000000000",
				MemberID: "000000000000000000000001",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members:  []models.Member{},
					}, nil).
					Once()
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, 0, nil).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceMemberNotFound("000000000000000000000000", nil),
			},
		},
		{
			description: "fails when the passive member is not on the namespace",
			req: &requests.NamespaceRemoveMember{
				UserID:   "000000000000000000000000",
				TenantID: "00000000-0000-4000-0000-000000000000",
				MemberID: "000000000000000000000001",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOwner,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, 0, nil).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceMemberNotFound("000000000000000000000001", nil),
			},
		},
		{
			description: "fails when the active member's role cannot act over passive member's role",
			req: &requests.NamespaceRemoveMember{
				UserID:   "000000000000000000000000",
				TenantID: "00000000-0000-4000-0000-000000000000",
				MemberID: "000000000000000000000001",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOperator,
							},
							{
								ID:   "000000000000000000000001",
								Role: authorizer.RoleAdministrator,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, 0, nil).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrRoleInvalid(),
			},
		},
		{
			description: "fails when cannot remove the member",
			req: &requests.NamespaceRemoveMember{
				UserID:   "000000000000000000000000",
				TenantID: "00000000-0000-4000-0000-000000000000",
				MemberID: "000000000000000000000001",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOwner,
							},
							{
								ID:   "000000000000000000000001",
								Role: authorizer.RoleAdministrator,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, 0, nil).
					Once()
				storeMock.
					On("NamespaceRemoveMember", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000001").
					Return(errors.New("error")).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       errors.New("error"),
			},
		},
		{
			description: "succeeds",
			req: &requests.NamespaceRemoveMember{
				UserID:   "000000000000000000000000",
				TenantID: "00000000-0000-4000-0000-000000000000",
				MemberID: "000000000000000000000001",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOwner,
							},
							{
								ID:   "000000000000000000000001",
								Role: authorizer.RoleAdministrator,
							},
						},
					}, nil).
					Once()
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{
						ID:       "000000000000000000000000",
						UserData: models.UserData{Username: "jane_doe"},
					}, 0, nil).
					Once()
				storeMock.
					On("NamespaceRemoveMember", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000001").
					Return(nil).
					Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "namespace",
						Owner:    "000000000000000000000000",
						Members: []models.Member{
							{
								ID:   "000000000000000000000000",
								Role: authorizer.RoleOwner,
							},
						},
					}, nil).
					Once()
			},
			expected: Expected{
				namespace: &models.Namespace{
					TenantID: "00000000-0000-4000-0000-000000000000",
					Name:     "namespace",
					Owner:    "000000000000000000000000",
					Members: []models.Member{
						{
							ID:   "000000000000000000000000",
							Role: authorizer.RoleOwner,
						},
					},
				},
				err: nil,
			},
		},
	}

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.TODO()
			tc.requiredMocks(ctx)
			ns, err := s.RemoveNamespaceMember(ctx, tc.req)
			assert.Equal(t, tc.expected, Expected{ns, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestGetSessionRecord(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	type Expected struct {
		status bool
		err    error
	}

	cases := []struct {
		description   string
		requiredMocks func(namespace *models.Namespace)
		namespace     *models.Namespace
		tenantID      string
		expected      Expected
	}{
		{
			description: "fails when the namespace document is not found",
			namespace:   &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Settings: &models.NamespaceSettings{SessionRecord: true}},
			requiredMocks: func(namespace *models.Namespace) {
				mock.On("NamespaceGet", ctx, namespace.TenantID, false).Return(namespace, store.ErrNoDocuments).Once()
			},
			expected: Expected{false, NewErrNamespaceNotFound("a736a52b-5777-4f92-b0b8-e359bf484713", store.ErrNoDocuments)},
		},
		{
			description: "fails when store namespace get fails",
			namespace:   &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Settings: &models.NamespaceSettings{SessionRecord: true}},
			requiredMocks: func(namespace *models.Namespace) {
				mock.On("NamespaceGet", ctx, namespace.TenantID, false).Return(nil, errors.New("error")).Once()
			},
			tenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
			expected: Expected{false, NewErrNamespaceNotFound("a736a52b-5777-4f92-b0b8-e359bf484713", errors.New("error"))},
		},
		{
			description: "fails when store namespace get session record fails",
			namespace:   &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Settings: &models.NamespaceSettings{SessionRecord: true}},
			requiredMocks: func(namespace *models.Namespace) {
				mock.On("NamespaceGet", ctx, namespace.TenantID, false).Return(namespace, nil).Once()
				mock.On("NamespaceGetSessionRecord", ctx, namespace.TenantID).Return(false, errors.New("error")).Once()
			},
			tenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
			expected: Expected{false, errors.New("error")},
		},
		{
			description: "succeeds",
			namespace:   &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Settings: &models.NamespaceSettings{SessionRecord: true}},
			requiredMocks: func(namespace *models.Namespace) {
				mock.On("NamespaceGet", ctx, namespace.TenantID, false).Return(namespace, nil).Once()
				mock.On("NamespaceGetSessionRecord", ctx, namespace.TenantID).Return(true, nil).Once()
			},
			tenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
			expected: Expected{true, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks(tc.namespace)

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock)
			returnedUserSecurity, err := service.GetSessionRecord(ctx, tc.namespace.TenantID)
			assert.Equal(t, tc.expected, Expected{returnedUserSecurity, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestEditSessionRecord(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	cases := []struct {
		description   string
		namespace     *models.Namespace
		requiredMocks func()
		sessionRecord bool
		tenantID      string
		expected      error
	}{
		{
			description: "fails when namespace set session record fails",
			namespace: &models.Namespace{
				Name:     "group1",
				Owner:    "hash1",
				TenantID: "xxxx",
				Settings: &models.NamespaceSettings{SessionRecord: true},
				Members: []models.Member{
					{
						ID:   "hash1",
						Role: authorizer.RoleOwner,
					},
					{
						ID:   "hash2",
						Role: authorizer.RoleObserver,
					},
				},
			},
			requiredMocks: func() {
				namespace := &models.Namespace{
					Name:     "group1",
					Owner:    "hash1",
					TenantID: "xxxx",
					Settings: &models.NamespaceSettings{SessionRecord: true},
					Members: []models.Member{
						{
							ID:   "hash1",
							Role: authorizer.RoleOwner,
						},
						{
							ID:   "hash2",
							Role: authorizer.RoleObserver,
						},
					},
				}

				status := true
				mock.On("NamespaceSetSessionRecord", ctx, status, namespace.TenantID).Return(errors.New("error")).Once()
			},
			tenantID:      "xxxx",
			sessionRecord: true,
			expected:      errors.New("error"),
		},
		{
			description: "succeeds",
			namespace: &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "xxxx", Settings: &models.NamespaceSettings{SessionRecord: true}, Members: []models.Member{
				{
					ID:   "hash1",
					Role: authorizer.RoleOwner,
				},
				{
					ID:   "hash2",
					Role: authorizer.RoleObserver,
				},
			}},
			requiredMocks: func() {
				namespace := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "xxxx", Settings: &models.NamespaceSettings{SessionRecord: true}, Members: []models.Member{
					{
						ID:   "hash1",
						Role: authorizer.RoleOwner,
					},
					{
						ID:   "hash2",
						Role: authorizer.RoleObserver,
					},
				}}

				status := true
				mock.On("NamespaceSetSessionRecord", ctx, status, namespace.TenantID).Return(nil).Once()
			},
			tenantID:      "xxxx",
			sessionRecord: true,
			expected:      nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock)
			err := service.EditSessionRecordStatus(ctx, tc.sessionRecord, tc.tenantID)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}
