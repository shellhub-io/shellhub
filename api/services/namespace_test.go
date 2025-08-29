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
				queryOptionsMock.
					On("Match", &query.Filters{}).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Paginate", &query.Paginator{Page: 1, PerPage: 10}).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceList", ctx, mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
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
				queryOptionsMock.
					On("Match", &query.Filters{}).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Paginate", &query.Paginator{Page: 1, PerPage: 10}).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceList", ctx, mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(
						[]models.Namespace{
							{
								Name:     "group1",
								Owner:    "66ffe0745a82ba5c4fe842ac",
								TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
								Type:     models.TypeTeam,
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
						Type:     models.TypeTeam,
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
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
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
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
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
			description: "succeeds - personal (with have changed to team temporarily)",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
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
							Type: models.TypeTeam,
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
					Type: models.TypeTeam,
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
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(nil, store.ErrNoDocuments).
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
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:            "000000000000000000000000",
						MaxNamespaces: 0,
					}, nil).
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
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:            "000000000000000000000000",
						MaxNamespaces: 1,
					}, nil).
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
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:            "000000000000000000000000",
						MaxNamespaces: 3,
					}, nil).
					Once()
				storeMock.
					On("NamespaceConflicts", ctx, &models.NamespaceConflicts{Name: "namespace"}).
					Return(nil, true, nil).
					Once()
			},
			expected: Expected{
				ns:  nil,
				err: NewErrNamespaceDuplicated(nil),
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
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:            "000000000000000000000000",
						MaxNamespaces: 3,
					}, nil).
					Once()
				storeMock.
					On("NamespaceConflicts", ctx, &models.NamespaceConflicts{Name: "namespace"}).
					Return(nil, false, nil).
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
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:            "000000000000000000000000",
						MaxNamespaces: 3,
					}, nil).
					Once()
				storeMock.
					On("NamespaceConflicts", ctx, &models.NamespaceConflicts{Name: "namespace"}).
					Return(nil, false, nil).
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
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:            "000000000000000000000000",
						MaxNamespaces: 3,
					}, nil).
					Once()
				storeMock.
					On("NamespaceConflicts", ctx, &models.NamespaceConflicts{Name: "namespace"}).
					Return(nil, false, nil).
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
						ConnectionAnnouncement: models.DefaultAnnouncementMessage,
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
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:            "000000000000000000000000",
						MaxNamespaces: 3,
					}, nil).
					Once()
				storeMock.
					On("NamespaceConflicts", ctx, &models.NamespaceConflicts{Name: "namespace"}).
					Return(nil, false, nil).
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
							MaxDevices: -1,
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
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(&models.User{
						ID:            "000000000000000000000000",
						MaxNamespaces: 3,
					}, nil).
					Once()
				storeMock.
					On("NamespaceConflicts", ctx, &models.NamespaceConflicts{Name: "namespace"}).
					Return(nil, false, nil).
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
					On("NamespaceUpdate", ctx, "xxxxx", &models.NamespaceChanges{Name: "newname"}).
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
					On("NamespaceUpdate", ctx, "xxxxx", &models.NamespaceChanges{Name: "newname"}).
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
					On("NamespaceUpdate", ctx, "xxxxx", &models.NamespaceChanges{Name: "newname"}).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "xxxxx").
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
					On("NamespaceUpdate", ctx, "xxxxx", &models.NamespaceChanges{Name: "newname"}).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "xxxxx").
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
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: NewErrNamespaceNotFound("00000000-0000-4000-0000-000000000000", errors.New("error")),
		},
		{
			description: "fails when store delete fails",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000"}, nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
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
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000"}, nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
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

func TestEditSessionRecord(t *testing.T) {
	storeMock := new(storemock.Store)

	cases := []struct {
		name          string
		sessionRecord bool
		tenantID      string
		mocks         func(context.Context)
		expected      error
	}{
		{
			name:          "fails when namespace edit fails",
			sessionRecord: true,
			tenantID:      "xxxx",
			mocks: func(ctx context.Context) {
				sessionRecord := true
				storeMock.
					On("NamespaceUpdate", ctx, "xxxx", &models.NamespaceChanges{SessionRecord: &sessionRecord}).
					Return(errors.New("error")).
					Once()
			},
			expected: errors.New("error"),
		},
		{
			name:          "fails when namespace not found",
			sessionRecord: true,
			tenantID:      "xxxx",
			mocks: func(ctx context.Context) {
				sessionRecord := true
				storeMock.
					On("NamespaceUpdate", ctx, "xxxx", &models.NamespaceChanges{SessionRecord: &sessionRecord}).
					Return(store.ErrNoDocuments).
					Once()
			},
			expected: NewErrNamespaceNotFound("xxxx", store.ErrNoDocuments),
		},
		{
			name:          "succeeds",
			sessionRecord: true,
			tenantID:      "xxxx",
			mocks: func(ctx context.Context) {
				sessionRecord := true
				storeMock.
					On("NamespaceUpdate", ctx, "xxxx", &models.NamespaceChanges{SessionRecord: &sessionRecord}).
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tc.mocks(ctx)
			err := s.EditSessionRecordStatus(ctx, tc.sessionRecord, tc.tenantID)
			assert.Equal(t, tc.expected, err)
		})
	}

	storeMock.AssertExpectations(t)
}

func TestGetSessionRecord(t *testing.T) {
	storeMock := new(storemock.Store)

	type Expected struct {
		status bool
		err    error
	}

	cases := []struct {
		name     string
		tenantID string
		mocks    func(context.Context)
		expected Expected
	}{
		{
			name:     "fails when namespace not found",
			tenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
			mocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "a736a52b-5777-4f92-b0b8-e359bf484713").Return(nil, store.ErrNoDocuments).Once()
			},
			expected: Expected{false, NewErrNamespaceNotFound("a736a52b-5777-4f92-b0b8-e359bf484713", store.ErrNoDocuments)},
		},
		{
			name:     "fails when store namespace resolve fails",
			tenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
			mocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "a736a52b-5777-4f92-b0b8-e359bf484713").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: Expected{false, NewErrNamespaceNotFound("a736a52b-5777-4f92-b0b8-e359bf484713", errors.New("error"))},
		},
		{
			name:     "succeeds",
			tenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
			mocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "a736a52b-5777-4f92-b0b8-e359bf484713").
					Return(
						&models.Namespace{
							Name:     "group1",
							Owner:    "hash1",
							TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
							Settings: &models.NamespaceSettings{SessionRecord: false},
						},
						nil,
					).
					Once()
			},
			expected: Expected{false, nil},
		},
	}

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tc.mocks(ctx)
			status, err := s.GetSessionRecord(ctx, tc.tenantID)
			assert.Equal(t, tc.expected, Expected{status, err})
		})
	}

	storeMock.AssertExpectations(t)
}
