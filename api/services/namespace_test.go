package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	storemock "github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	uuidmocks "github.com/shellhub-io/shellhub/pkg/uuid/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListNamespaces(t *testing.T) {
	storeMock := new(storemock.Store)
	queryOptionsMock := new(storemock.QueryOptions)
	storeMock.On("Options").Return(queryOptionsMock)

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
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				queryOptionsMock.On("EnrichMembersData").Return(nil).Once()
				storeMock.
					On("NamespaceList", ctx, query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, mock.AnythingOfType("store.NamespaceQueryOption"), mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(nil, 0, errors.New("error")).
					Once()
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
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				queryOptionsMock.On("EnrichMembersData").Return(nil).Once()
				storeMock.
					On("NamespaceList", ctx, query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, mock.Anything, mock.Anything).
					Return(
						[]models.Namespace{
							{
								Name:     "group1",
								Owner:    "66ffe0745a82ba5c4fe842ac",
								TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
								Type:     models.TypePersonal,
								Members: []models.Member{
									{
										ID:    "66ffe0745a82ba5c4fe842ac",
										Role:  authorizer.RoleOwner,
										Email: "john.doe@test.com",
									},
								},
							},
							{
								Name:     "group2",
								Owner:    "66ffe0745a82ba5c4fe842ac",
								TenantID: "a736a52b-5777-4f92-b0b8-e359bf48471i4",
								Type:     models.TypePersonal,
								Members: []models.Member{
									{
										ID:    "66ffe0745a82ba5c4fe842ac",
										Role:  authorizer.RoleOwner,
										Email: "john.doe@test.com",
									},
									{
										ID:    "66ffe0232da6d319c9769afb",
										Role:  authorizer.RoleObserver,
										Email: "jane.smith@test.com",
									},
								},
							},
							{
								Name:     "group3",
								Owner:    "66ffe0745a82ba5c4fe842ac",
								TenantID: "a736a52b-5777-4f92-b0b8-e359bf48471i4",
								Type:     models.TypeTeam,
								Members: []models.Member{
									{
										ID:    "66ffe0745a82ba5c4fe842ac",
										Role:  authorizer.RoleOwner,
										Email: "john.doe@test.com",
									},
									{
										ID:    "66ffe0232da6d319c9769afb",
										Role:  authorizer.RoleObserver,
										Email: "jane.smith@test.com",
									},
								},
							},
						},
						2,
						nil,
					).
					Once()
			},
			expected: Expected{
				namespaces: []models.Namespace{
					{
						Name:     "group1",
						Owner:    "66ffe0745a82ba5c4fe842ac",
						TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
						Type:     models.TypePersonal,
						Members: []models.Member{
							{
								ID:    "66ffe0745a82ba5c4fe842ac",
								Role:  authorizer.RoleOwner,
								Email: "john.doe@test.com",
							},
						},
					},
					{
						Name:     "group2",
						Owner:    "66ffe0745a82ba5c4fe842ac",
						TenantID: "a736a52b-5777-4f92-b0b8-e359bf48471i4",
						Type:     models.TypePersonal,
						Members: []models.Member{
							{
								ID:    "66ffe0745a82ba5c4fe842ac",
								Role:  authorizer.RoleOwner,
								Email: "john.doe@test.com",
							},
							{
								ID:    "66ffe0232da6d319c9769afb",
								Role:  authorizer.RoleObserver,
								Email: "jane.smith@test.com",
							},
						},
					},
					{
						Name:     "group3",
						Owner:    "66ffe0745a82ba5c4fe842ac",
						TenantID: "a736a52b-5777-4f92-b0b8-e359bf48471i4",
						Type:     models.TypeTeam,
						Members: []models.Member{
							{
								ID:    "66ffe0745a82ba5c4fe842ac",
								Role:  authorizer.RoleOwner,
								Email: "john.doe@test.com",
							},
							{
								ID:    "66ffe0232da6d319c9769afb",
								Role:  authorizer.RoleObserver,
								Email: "jane.smith@test.com",
							},
						},
					},
				},
				count: 2,
				err:   nil,
			},
		},
	}

	s := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			nss, count, err := s.ListNamespaces(tc.ctx, tc.req)
			assert.Equal(t, tc.expected, Expected{nss, count, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestGetNamespace(t *testing.T) {
	storeMock := new(storemock.Store)
	queryOptionsMock := new(storemock.QueryOptions)
	storeMock.On("Options").Return(queryOptionsMock)

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
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				queryOptionsMock.On("EnrichMembersData").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption"), mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(nil, errors.New("error")).
					Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceNotFound("00000000-0000-4000-0000-000000000000", errors.New("error")),
			},
		},
		{
			description: "succeeds - team",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				queryOptionsMock.On("EnrichMembersData").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption"), mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							Name:     "group1",
							Owner:    "66ffe21f76d5207a38a056d5",
							TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
							Type:     models.TypeTeam,
							Members: []models.Member{
								{
									ID:    "66ffe21f76d5207a38a056d5",
									Role:  authorizer.RoleOwner,
									Email: "john.doe@test.com",
								},
							},
						},
						nil,
					).
					Once()
			},
			expected: Expected{
				namespace: &models.Namespace{
					Name:     "group1",
					Owner:    "66ffe21f76d5207a38a056d5",
					TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
					Type:     models.TypeTeam,
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
		{
			description: "succeeds - personal",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				queryOptionsMock.On("EnrichMembersData").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption"), mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
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
							Type: models.TypePersonal,
						},
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
					Type: models.TypePersonal,
				},
				err: nil,
			},
		},
	}

	s := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			ns, err := s.GetNamespace(ctx, tc.tenantID)
			assert.Equal(t, tc.expected, Expected{ns, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestCreateNamespace(t *testing.T) {
	storeMock := new(storemock.Store)
	clockMock := new(clockmock.Clock)
	clock.DefaultBackend = clockMock

	now := time.Now()
	clockMock.On("Now").Return(now)

	ctx := context.TODO()

	uuidMock := &uuidmocks.Uuid{}
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
			description: "fails when user reachs the zero namespaces",
			req: &requests.NamespaceCreate{
				UserID:   "000000000000000000000000",
				Name:     "namespace",
				TenantID: "00000000-0000-4000-0000-000000000000",
			},
			requiredMocks: func() {
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{
						ID:            "000000000000000000000000",
						MaxNamespaces: 0,
					}, 0, nil).
					Once()
			},
			expected: Expected{
				ns:  nil,
				err: NewErrNamespaceCreationIsForbidden(0, nil),
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
					Return(&models.User{
						ID:            "000000000000000000000000",
						MaxNamespaces: 1,
					}, 0, nil).
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
					On("UserGetInfo", ctx, "000000000000000000000000").
					Return(
						&models.UserInfo{
							OwnedNamespaces:      []models.Namespace{{}},
							AssociatedNamespaces: []models.Namespace{},
						},
						nil,
					).
					Once()
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{
						ID:            "000000000000000000000000",
						MaxNamespaces: 3,
					}, 0, nil).
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
					On("UserGetInfo", ctx, "000000000000000000000000").
					Return(
						&models.UserInfo{
							OwnedNamespaces:      []models.Namespace{{}},
							AssociatedNamespaces: []models.Namespace{},
						},
						nil,
					).
					Once()
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{
						ID:            "000000000000000000000000",
						MaxNamespaces: 3,
					}, 0, nil).
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
					On("UserGetInfo", ctx, "000000000000000000000000").
					Return(
						&models.UserInfo{
							OwnedNamespaces:      []models.Namespace{{}},
							AssociatedNamespaces: []models.Namespace{},
						},
						nil,
					).
					Once()
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{
						ID:            "000000000000000000000000",
						MaxNamespaces: 3,
					}, 0, nil).
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
					Twice()
				// --
				storeMock.
					On(
						"NamespaceCreate",
						ctx,
						&models.Namespace{
							TenantID: "00000000-0000-4000-0000-000000000000",
							Name:     "namespace",
							Owner:    "000000000000000000000000",
							Type:     models.TypePersonal,
							Members: []models.Member{
								{
									ID:      "000000000000000000000000",
									Role:    authorizer.RoleOwner,
									Status:  models.MemberStatusAccepted,
									AddedAt: now,
								},
							},
							Settings: &models.NamespaceSettings{
								SessionRecord:          false,
								ConnectionAnnouncement: models.CommunityConnectionAnnouncement,
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
					On("UserGetInfo", ctx, "000000000000000000000000").
					Return(
						&models.UserInfo{
							OwnedNamespaces:      []models.Namespace{{}},
							AssociatedNamespaces: []models.Namespace{},
						},
						nil,
					).
					Once()
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{
						ID:            "000000000000000000000000",
						MaxNamespaces: 3,
					}, 0, nil).
					Once()
				storeMock.
					On("NamespaceGetByName", ctx, "namespace").
					Return(nil, store.ErrNoDocuments).
					Once()
				// envs.IsCommunity = true
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Twice()
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
							Type:     models.TypePersonal,
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
								ConnectionAnnouncement: models.CommunityConnectionAnnouncement,
							},
							MaxDevices: -1,
						},
					).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-4000-0000-000000000000",
							Name:     "namespace",
							Owner:    "000000000000000000000000",
							Type:     models.TypePersonal,
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
								ConnectionAnnouncement: models.CommunityConnectionAnnouncement,
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
					Type:     models.TypePersonal,
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
						ConnectionAnnouncement: models.CommunityConnectionAnnouncement,
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
					On("UserGetInfo", ctx, "000000000000000000000000").
					Return(
						&models.UserInfo{
							OwnedNamespaces:      []models.Namespace{{}},
							AssociatedNamespaces: []models.Namespace{},
						},
						nil,
					).
					Once()
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{
						ID:            "000000000000000000000000",
						MaxNamespaces: 3,
					}, 0, nil).
					Once()
				storeMock.
					On("NamespaceGetByName", ctx, "namespace").
					Return(nil, store.ErrNoDocuments).
					Once()
				// envs.IsCommunity = true
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Twice()
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
							Type:     models.TypePersonal,
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
								ConnectionAnnouncement: models.CommunityConnectionAnnouncement,
							},
							MaxDevices: -1,
						},
					).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-4000-0000-000000000000",
							Name:     "namespace",
							Owner:    "000000000000000000000000",
							Type:     models.TypePersonal,
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
								ConnectionAnnouncement: models.CommunityConnectionAnnouncement,
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
					Type:     models.TypePersonal,
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
						ConnectionAnnouncement: models.CommunityConnectionAnnouncement,
					},
					MaxDevices: -1,
				},
				err: nil,
			},
		},
		{
			description: "succeeds to create a namespace-:-env=cloud type team",
			req: &requests.NamespaceCreate{
				UserID:   "000000000000000000000000",
				Name:     "namespace",
				TenantID: "00000000-0000-4000-0000-000000000000",
				Type:     "team",
			},
			requiredMocks: func() {
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
				// envs.IsCommunity = false
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{
						ID:            "000000000000000000000000",
						MaxNamespaces: 3,
					}, 0, nil).
					Once()
				storeMock.
					On("NamespaceGetByName", ctx, "namespace").
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Twice()
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
							Type:     models.TypeTeam,
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
							Type:     models.TypeTeam,
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
								ConnectionAnnouncement: models.CommunityConnectionAnnouncement,
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
					Type:     models.TypeTeam,
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
		{
			description: "succeeds to create a namespace-:-env=cloud",
			req: &requests.NamespaceCreate{
				UserID:   "000000000000000000000000",
				Name:     "namespace",
				TenantID: "00000000-0000-4000-0000-000000000000",
				Type:     "",
			},
			requiredMocks: func() {
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
				// envs.IsCommunity = false
				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(&models.User{
						ID:            "000000000000000000000000",
						MaxNamespaces: 3,
					}, 0, nil).
					Once()
				storeMock.
					On("NamespaceGetByName", ctx, "namespace").
					Return(nil, store.ErrNoDocuments).
					Once()
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
					Twice()
				// --
				storeMock.
					On(
						"NamespaceCreate",
						ctx,
						&models.Namespace{
							TenantID: "00000000-0000-4000-0000-000000000000",
							Name:     "namespace",
							Owner:    "000000000000000000000000",
							Type:     models.TypeTeam,
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
							Type:     models.TypePersonal,
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
								ConnectionAnnouncement: models.CommunityConnectionAnnouncement,
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
					Type:     models.TypeTeam,
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
	storeMock := new(storemock.Store)
	queryOptionsMock := new(storemock.QueryOptions)
	storeMock.On("Options").Return(queryOptionsMock)

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
				storeMock.
					On("NamespaceEdit", ctx, "xxxxx", &models.NamespaceChanges{Name: "newname"}).
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
				storeMock.
					On("NamespaceEdit", ctx, "xxxxx", &models.NamespaceChanges{Name: "newname"}).
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
				storeMock.
					On("NamespaceEdit", ctx, "xxxxx", &models.NamespaceChanges{Name: "newname"}).
					Return(nil).
					Once()
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				queryOptionsMock.On("EnrichMembersData").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "xxxxx", mock.AnythingOfType("store.NamespaceQueryOption"), mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "xxxxx",
							Name:     "newname",
						}, nil,
					).
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
				storeMock.
					On("NamespaceEdit", ctx, "xxxxx", &models.NamespaceChanges{Name: "newname"}).
					Return(nil).
					Once()
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				queryOptionsMock.On("EnrichMembersData").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "xxxxx", mock.AnythingOfType("store.NamespaceQueryOption"), mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "xxxxx",
							Name:     "newname",
						}, nil,
					).
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

			service := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

			req := &requests.NamespaceEdit{
				TenantParam: requests.TenantParam{Tenant: tc.tenantID},
				Name:        tc.namespaceName,
			}
			namespace, err := service.EditNamespace(ctx, req)

			assert.Equal(t, tc.expected, Expected{namespace, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestDeleteNamespace(t *testing.T) {
	storeMock := new(storemock.Store)
	queryOptionsMock := new(storemock.QueryOptions)
	storeMock.On("Options").Return(queryOptionsMock)

	ctx := context.TODO()

	cases := []struct {
		description   string
		tenantID      string
		requiredMocks func()
		expected      error
	}{
		{
			description: "fails when namespace does not exist",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(nil, errors.New("error")).
					Once()
			},
			expected: NewErrNamespaceNotFound("00000000-0000-4000-0000-000000000000", errors.New("error")),
		},
		{
			description: "fails when store delete fails",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000"}, nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				envMock.
					On("Get", "SHELLHUB_BILLING").
					Return("false").
					Once()
				storeMock.
					On("NamespaceDelete", ctx, "00000000-0000-4000-0000-000000000000").
					Return(errors.New("error")).
					Once()
			},
			expected: errors.New("error"),
		},
		{
			description: "succeeds",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000"}, nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				envMock.
					On("Get", "SHELLHUB_BILLING").
					Return("false").
					Once()
				storeMock.
					On("NamespaceDelete", ctx, "00000000-0000-4000-0000-000000000000").
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	s := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			err := s.DeleteNamespace(ctx, tc.tenantID)
			assert.Equal(t, tc.expected, err)
		})
	}

	storeMock.AssertExpectations(t)
}

func TestGetSessionRecord(t *testing.T) {
	storeMock := new(storemock.Store)

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
				storeMock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, store.ErrNoDocuments).Once()
			},
			expected: Expected{false, NewErrNamespaceNotFound("a736a52b-5777-4f92-b0b8-e359bf484713", store.ErrNoDocuments)},
		},
		{
			description: "fails when store namespace get fails",
			namespace:   &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Settings: &models.NamespaceSettings{SessionRecord: true}},
			requiredMocks: func(namespace *models.Namespace) {
				storeMock.On("NamespaceGet", ctx, namespace.TenantID).Return(nil, errors.New("error")).Once()
			},
			tenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
			expected: Expected{false, NewErrNamespaceNotFound("a736a52b-5777-4f92-b0b8-e359bf484713", errors.New("error"))},
		},
		{
			description: "fails when store namespace get session record fails",
			namespace:   &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Settings: &models.NamespaceSettings{SessionRecord: true}},
			requiredMocks: func(namespace *models.Namespace) {
				storeMock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				storeMock.On("NamespaceGetSessionRecord", ctx, namespace.TenantID).Return(false, errors.New("error")).Once()
			},
			tenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
			expected: Expected{false, errors.New("error")},
		},
		{
			description: "succeeds",
			namespace:   &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Settings: &models.NamespaceSettings{SessionRecord: true}},
			requiredMocks: func(namespace *models.Namespace) {
				storeMock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				storeMock.On("NamespaceGetSessionRecord", ctx, namespace.TenantID).Return(true, nil).Once()
			},
			tenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
			expected: Expected{true, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks(tc.namespace)

			service := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)
			returnedUserSecurity, err := service.GetSessionRecord(ctx, tc.namespace.TenantID)
			assert.Equal(t, tc.expected, Expected{returnedUserSecurity, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestEditSessionRecord(t *testing.T) {
	storeMock := new(storemock.Store)

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
				storeMock.On("NamespaceSetSessionRecord", ctx, status, namespace.TenantID).Return(errors.New("error")).Once()
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
				storeMock.On("NamespaceSetSessionRecord", ctx, status, namespace.TenantID).Return(nil).Once()
			},
			tenantID:      "xxxx",
			sessionRecord: true,
			expected:      nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)
			err := service.EditSessionRecordStatus(ctx, tc.sessionRecord, tc.tenantID)
			assert.Equal(t, tc.expected, err)
		})
	}

	storeMock.AssertExpectations(t)
}
