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
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/envs/envstest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	uuidmocks "github.com/shellhub-io/shellhub/pkg/uuid/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListNamespaces(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	queryOptionsMock := storemock.NewMockQueryOptions(t)
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
				IsAdmin:   true,
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
					On("NamespaceList", ctx, mock.AnythingOfType("[]store.QueryOption")).
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
				IsAdmin:   true,
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
					On("NamespaceList", ctx, mock.AnythingOfType("[]store.QueryOption")).
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
		{
			description: "success to get the namespace list with user filter",
			req: &requests.NamespaceList{
				UserID:    "66ffe0745a82ba5c4fe842ac",
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
				queryOptionsMock.
					On("WithMember", "66ffe0745a82ba5c4fe842ac").
					Return(nil).
					Once()
				storeMock.
					On("NamespaceList", ctx, mock.AnythingOfType("[]store.QueryOption")).
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
						},
						1,
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
				},
				count: 1,
				err:   nil,
			},
		},
		{
			description: "api key caller is scoped to its own tenant",
			req: &requests.NamespaceList{
				TenantID:  "a736a52b-5777-4f92-b0b8-e359bf484713",
				Paginator: query.Paginator{Page: 1, PerPage: 10},
				Filters:   query.Filters{},
			},
			ctx: ctx,
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "a736a52b-5777-4f92-b0b8-e359bf484713").
					Return(&models.Namespace{Name: "own", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713"}, nil).
					Once()
			},
			expected: Expected{
				namespaces: []models.Namespace{{Name: "own", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713"}},
				count:      1,
				err:        nil,
			},
		},
		{
			description: "api key caller without tenant id returns empty list",
			req: &requests.NamespaceList{
				Paginator: query.Paginator{Page: 1, PerPage: 10},
				Filters:   query.Filters{},
			},
			ctx:           ctx,
			requiredMocks: func() {},
			expected: Expected{
				namespaces: []models.Namespace{},
				count:      0,
				err:        nil,
			},
		},
		{
			description: "type filter name is rewritten to scope before store call",
			req: &requests.NamespaceList{
				IsAdmin:   true,
				Paginator: query.Paginator{Page: 1, PerPage: 10},
				Filters: query.Filters{
					Data: []query.Filter{
						{
							Type:   query.FilterTypeProperty,
							Params: &query.FilterProperty{Name: "type", Operator: "eq", Value: "team"},
						},
					},
				},
			},
			ctx: ctx,
			requiredMocks: func() {
				queryOptionsMock.
					On("Match", &query.Filters{
						Data: []query.Filter{
							{
								Type:   query.FilterTypeProperty,
								Params: &query.FilterProperty{Name: "scope", Operator: "eq", Value: "team"},
							},
						},
					}).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Paginate", &query.Paginator{Page: 1, PerPage: 10}).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceList", ctx, mock.AnythingOfType("[]store.QueryOption")).
					Return([]models.Namespace{}, 0, nil).
					Once()
			},
			expected: Expected{
				namespaces: []models.Namespace{},
				count:      0,
				err:        nil,
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
	storeMock := storemock.NewMockStore(t)

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
	storeMock := storemock.NewMockStore(t)
	// A namespace create also provisions its legacy install key (best-effort).
	storeMock.On("InstallKeyCreate", mock.Anything, mock.Anything).Return("", nil).Maybe()
	clockMock := clockmock.NewMockClock(t)

	prevClockBackend := clock.DefaultBackend
	t.Cleanup(func() {
		clock.DefaultBackend = prevClockBackend
	})
	clock.DefaultBackend = clockMock

	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	clockMock.On("Now").Return(now)

	ctx := context.TODO()

	uuidMock := uuidmocks.NewMockUUID(t)
	uuid.DefaultBackend = uuidMock

	type Expected struct {
		ns  *models.Namespace
		err error
	}

	cases := []struct {
		description   string
		edition       envs.Edition
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
			edition:     envs.Community,
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
					Return("", errors.New("error")).
					Once()
			},
			expected: Expected{
				ns:  nil,
				err: NewErrNamespaceCreateStore(errors.New("error")),
			},
		},
		{
			// store.ErrDuplicate from NamespaceCreate means a concurrent insert raced past the
			// NamespaceConflicts pre-check (case-sensitive name=? queries rely on lowercased
			// writes). The service must map it to ErrNamespaceDuplicated, not ErrNamespaceCreateStore.
			description: "fails when NamespaceCreate returns store.ErrDuplicate",
			edition:     envs.Community,
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
					Return("", store.ErrDuplicate).
					Once()
			},
			expected: Expected{
				ns:  nil,
				err: NewErrNamespaceDuplicated(store.ErrDuplicate),
			},
		},
		{
			description: "succeeds to create a namespace",
			edition:     envs.Community,
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
					Return("00000000-0000-4000-0000-000000000000", nil).
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
			edition:     envs.Community,
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
					Return("00000000-0000-4000-0000-000000000000", nil).
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
			edition:     envs.Cloud,
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
					Return("00000000-0000-4000-0000-000000000000", nil).
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
			edition:     envs.Cloud,
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
					Return("00000000-0000-4000-0000-000000000000", nil).
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
			if tc.edition != "" {
				envstest.SetEdition(t, tc.edition)
			}

			tc.requiredMocks()

			service := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)
			returnedNamespace, err := service.CreateNamespace(ctx, tc.req)

			assert.Equal(t, tc.expected, Expected{returnedNamespace, err})

			storeMock.AssertExpectations(t)
		})
	}
}

func TestEditNamespace(t *testing.T) {
	storeMock := storemock.NewMockStore(t)

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
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "xxxxx").
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expected: Expected{
				nil,
				NewErrNamespaceNotFound("xxxxx", store.ErrNoDocuments),
			},
		},
		{
			description:   "fails when the store namespace update fails",
			tenantID:      "xxxxx",
			namespaceName: "newname",
			requiredMocks: func() {
				namespace := &models.Namespace{
					TenantID: "xxxxx",
					Name:     "oldname",
					Settings: &models.NamespaceSettings{SessionRecord: false},
				}
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "xxxxx").
					Return(namespace, nil).
					Once()

				expectedNamespace := *namespace
				expectedNamespace.Name = "newname"
				storeMock.
					On("NamespaceUpdate", ctx, &expectedNamespace).
					Return(errors.New("error")).
					Once()
			},
			expected: Expected{
				nil,
				errors.New("error"),
			},
		},
		{
			description:   "fails with ErrNamespaceDuplicated when store returns ErrDuplicate",
			tenantID:      "xxxxx",
			namespaceName: "newname",
			requiredMocks: func() {
				namespace := &models.Namespace{
					TenantID: "xxxxx",
					Name:     "oldname",
					Settings: &models.NamespaceSettings{SessionRecord: false},
				}
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "xxxxx").
					Return(namespace, nil).
					Once()

				expectedNamespace := *namespace
				expectedNamespace.Name = "newname"
				storeMock.
					On("NamespaceUpdate", ctx, &expectedNamespace).
					Return(store.ErrDuplicate).
					Once()
			},
			expected: Expected{
				nil,
				NewErrNamespaceDuplicated(store.ErrDuplicate),
			},
		},
		{
			description:   "succeeds changing the name to lowercase",
			namespaceName: "newName",
			tenantID:      "xxxxx",
			requiredMocks: func() {
				namespace := &models.Namespace{
					TenantID: "xxxxx",
					Name:     "oldname",
					Settings: &models.NamespaceSettings{SessionRecord: false},
				}
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "xxxxx").
					Return(namespace, nil).
					Once()

				expectedNamespace := *namespace
				expectedNamespace.Name = "newname"
				storeMock.
					On("NamespaceUpdate", ctx, &expectedNamespace).
					Return(nil).
					Once()

				finalNamespace := &models.Namespace{
					TenantID: "xxxxx",
					Name:     "newname",
					Settings: &models.NamespaceSettings{SessionRecord: false},
				}
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "xxxxx").
					Return(finalNamespace, nil).
					Once()
			},
			expected: Expected{
				&models.Namespace{
					TenantID: "xxxxx",
					Name:     "newname",
					Settings: &models.NamespaceSettings{SessionRecord: false},
				},
				nil,
			},
		},
		{
			description:   "succeeds",
			namespaceName: "newname",
			tenantID:      "xxxxx",
			requiredMocks: func() {
				namespace := &models.Namespace{
					TenantID: "xxxxx",
					Name:     "oldname",
					Settings: &models.NamespaceSettings{SessionRecord: false},
				}
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "xxxxx").
					Return(namespace, nil).
					Once()

				expectedNamespace := *namespace
				expectedNamespace.Name = "newname"
				storeMock.
					On("NamespaceUpdate", ctx, &expectedNamespace).
					Return(nil).
					Once()

				finalNamespace := &models.Namespace{
					TenantID: "xxxxx",
					Name:     "newname",
					Settings: &models.NamespaceSettings{SessionRecord: false},
				}
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "xxxxx").
					Return(finalNamespace, nil).
					Once()
			},
			expected: Expected{
				&models.Namespace{
					TenantID: "xxxxx",
					Name:     "newname",
					Settings: &models.NamespaceSettings{SessionRecord: false},
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

func TestEditSessionRecord(t *testing.T) {
	storeMock := storemock.NewMockStore(t)

	cases := []struct {
		name          string
		sessionRecord bool
		tenantID      string
		mocks         func(context.Context)
		expected      error
	}{
		{
			name:          "fails when namespace not found",
			sessionRecord: true,
			tenantID:      "xxxx",
			mocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "xxxx").
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expected: NewErrNamespaceNotFound("xxxx", store.ErrNoDocuments),
		},
		{
			name:          "fails when namespace update fails",
			sessionRecord: true,
			tenantID:      "xxxx",
			mocks: func(ctx context.Context) {
				namespace := &models.Namespace{
					TenantID: "xxxx",
					Settings: &models.NamespaceSettings{SessionRecord: false},
				}
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "xxxx").
					Return(namespace, nil).
					Once()

				expectedNamespace := *namespace
				expectedNamespace.Settings.SessionRecord = true
				storeMock.
					On("NamespaceUpdate", ctx, &expectedNamespace).
					Return(errors.New("error")).
					Once()
			},
			expected: errors.New("error"),
		},
		{
			name:          "succeeds",
			sessionRecord: true,
			tenantID:      "xxxx",
			mocks: func(ctx context.Context) {
				namespace := &models.Namespace{
					TenantID: "xxxx",
					Settings: &models.NamespaceSettings{SessionRecord: false},
				}
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "xxxx").
					Return(namespace, nil).
					Once()

				expectedNamespace := *namespace
				expectedNamespace.Settings.SessionRecord = true
				storeMock.
					On("NamespaceUpdate", ctx, &expectedNamespace).
					Return(nil).
					Once()
			},
			expected: nil,
		},
		{
			name:          "succeeds when settings is nil",
			sessionRecord: true,
			tenantID:      "xxxx",
			mocks: func(ctx context.Context) {
				namespace := &models.Namespace{
					TenantID: "xxxx",
					Settings: &models.NamespaceSettings{},
				}
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "xxxx").
					Return(namespace, nil).
					Once()

				expectedNamespace := *namespace
				expectedNamespace.Settings = &models.NamespaceSettings{SessionRecord: true}
				storeMock.
					On("NamespaceUpdate", ctx, &expectedNamespace).
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

func TestEditSSHAccessMode(t *testing.T) {
	const tenantID = "00000000-0000-4000-0000-000000000000"
	const ownerID = "owner-1"

	cases := []struct {
		description string
		mode        string
		mocks       func(ctx context.Context, storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions)
		expected    error
	}{
		{
			description: "fails when namespace not found",
			mode:        models.SSHAccessModeIdentity,
			mocks: func(ctx context.Context, storeMock *storemock.MockStore, _ *storemock.MockQueryOptions) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expected: NewErrNamespaceNotFound(tenantID, store.ErrNoDocuments),
		},
		{
			description: "rejects legacy when the namespace is not grandfathered",
			mode:        models.SSHAccessModeLegacy,
			mocks: func(ctx context.Context, storeMock *storemock.MockStore, _ *storemock.MockQueryOptions) {
				namespace := &models.Namespace{
					TenantID: tenantID,
					Owner:    ownerID,
					Settings: &models.NamespaceSettings{SSHAccessMode: models.SSHAccessModeIdentity},
				}
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespace, nil).
					Once()
			},
			expected: NewErrForbidden(ErrNamespaceLegacyNotAllowed, nil),
		},
		{
			description: "allows legacy when the namespace is grandfathered",
			mode:        models.SSHAccessModeLegacy,
			mocks: func(ctx context.Context, storeMock *storemock.MockStore, _ *storemock.MockQueryOptions) {
				namespace := &models.Namespace{
					TenantID: tenantID,
					Owner:    ownerID,
					Settings: &models.NamespaceSettings{SSHAccessMode: models.SSHAccessModeIdentity, SSHLegacyAllowed: true},
				}
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespace, nil).
					Once()

				expectedNamespace := *namespace
				expectedNamespace.Settings.SSHAccessMode = models.SSHAccessModeLegacy
				storeMock.
					On("NamespaceUpdate", ctx, &expectedNamespace).
					Return(nil).
					Once()
			},
			expected: nil,
		},
		{
			description: "switching to identity seeds the owner access policy",
			mode:        models.SSHAccessModeIdentity,
			mocks: func(ctx context.Context, storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				namespace := &models.Namespace{
					TenantID: tenantID,
					Owner:    ownerID,
					Settings: &models.NamespaceSettings{SSHAccessMode: models.SSHAccessModeLegacy, SSHLegacyAllowed: true},
				}
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespace, nil).
					Once()

				expectedNamespace := *namespace
				expectedNamespace.Settings.SSHAccessMode = models.SSHAccessModeIdentity
				storeMock.
					On("NamespaceUpdate", ctx, &expectedNamespace).
					Return(nil).
					Once()

				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
				storeMock.
					On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{}, 0, nil).
					Once()
				storeMock.
					On("AccessPolicyCreate", ctx, models.NewOwnerAccessPolicy(tenantID, ownerID)).
					Return("policy-id", nil).
					Once()
			},
			expected: nil,
		},
		{
			description: "switching to identity does not seed when policies exist",
			mode:        models.SSHAccessModeIdentity,
			mocks: func(ctx context.Context, storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				namespace := &models.Namespace{
					TenantID: tenantID,
					Owner:    ownerID,
					Settings: &models.NamespaceSettings{SSHAccessMode: models.SSHAccessModeLegacy, SSHLegacyAllowed: true},
				}
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespace, nil).
					Once()

				expectedNamespace := *namespace
				expectedNamespace.Settings.SSHAccessMode = models.SSHAccessModeIdentity
				storeMock.
					On("NamespaceUpdate", ctx, &expectedNamespace).
					Return(nil).
					Once()

				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
				storeMock.
					On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{{ID: "existing"}}, 1, nil).
					Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			storeMock := new(storemock.MockStore)
			queryOptionsMock := new(storemock.MockQueryOptions)
			storeMock.On("Options").Return(queryOptionsMock).Maybe()

			tc.mocks(ctx, storeMock, queryOptionsMock)

			s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

			err := s.EditSSHAccessMode(ctx, tc.mode, tenantID)
			assert.Equal(t, tc.expected, err)

			storeMock.AssertExpectations(t)
		})
	}
}

func TestDeleteNamespace(t *testing.T) {
	storeMock := storemock.NewMockStore(t)

	ctx := context.TODO()

	cases := []struct {
		description   string
		edition       envs.Edition
		tenantID      string
		requiredMocks func()
		expected      error
	}{
		{
			description: "fails when namespace does not exist",
			edition:     envs.Community,
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
			edition:     envs.Community,
			tenantID:    "00000000-0000-4000-0000-000000000000",
			requiredMocks: func() {
				namespace := &models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000"}
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(namespace, nil).
					Once()
				storeMock.
					On("NamespaceDelete", ctx, namespace).
					Return(errors.New("error")).
					Once()
			},
			expected: errors.New("error"),
		},
		{
			description: "succeeds",
			edition:     envs.Community,
			tenantID:    "00000000-0000-4000-0000-000000000000",
			requiredMocks: func() {
				namespace := &models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000"}
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(namespace, nil).
					Once()
				storeMock.
					On("NamespaceDelete", ctx, namespace).
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	s := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			if tc.edition != "" {
				envstest.SetEdition(t, tc.edition)
			}

			tc.requiredMocks()

			err := s.DeleteNamespace(ctx, tc.tenantID)
			assert.Equal(t, tc.expected, err)
		})
	}

	storeMock.AssertExpectations(t)
}

func TestNamespaceFilterFields(t *testing.T) {
	t.Run("name field allows contains, eq and ne operators", func(t *testing.T) {
		assert.True(t, NamespaceFilterFields.Allows("name", "contains"))
		assert.True(t, NamespaceFilterFields.Allows("name", "eq"))
		assert.True(t, NamespaceFilterFields.Allows("name", "ne"))
	})

	t.Run("type field allows eq and ne operators", func(t *testing.T) {
		assert.True(t, NamespaceFilterFields.Allows("type", "eq"))
		assert.True(t, NamespaceFilterFields.Allows("type", "ne"))
	})

	t.Run("type field does not allow contains operator", func(t *testing.T) {
		assert.False(t, NamespaceFilterFields.Allows("type", "contains"))
	})

	t.Run("unknown field is rejected", func(t *testing.T) {
		assert.False(t, NamespaceFilterFields.Allows("unknown", "eq"))
	})

	t.Run("namespaceFilterColumns maps type to scope", func(t *testing.T) {
		assert.Equal(t, "scope", namespaceFilterColumns["type"])
	})

	t.Run("namespaceFilterColumns has exactly one entry", func(t *testing.T) {
		assert.Len(t, namespaceFilterColumns, 1)
	})
}
