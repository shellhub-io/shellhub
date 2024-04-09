package services

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	uuid_mocks "github.com/shellhub-io/shellhub/pkg/uuid/mocks"
	"github.com/shellhub-io/shellhub/pkg/validator"
	"github.com/stretchr/testify/assert"
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
		paginator     query.Paginator
		filters       query.Filters
		ctx           context.Context
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fail when could not get the namespace list",
			paginator:   query.Paginator{Page: 1, PerPage: 10},
			filters:     query.Filters{},
			ctx:         ctx,
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
			description: "fail when could not get a user",
			paginator:   query.Paginator{Page: 1, PerPage: 10},
			filters:     query.Filters{},
			ctx:         ctx,
			requiredMocks: func() {
				namespaces := []models.Namespace{
					{
						Name:     "group1",
						Owner:    "hash",
						TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
						Members: []models.Member{
							{
								ID:   "hash",
								Role: guard.RoleOwner,
							},
						},
					},
					{
						Name:     "group2",
						Owner:    "hash",
						TenantID: "a736a52b-5777-4f92-b0b8-e359bf48471i4",
						Members: []models.Member{
							{
								ID:   "hash",
								Role: guard.RoleOwner,
							},
							{
								ID:   "hash2",
								Role: guard.RoleObserver,
							},
						},
					},
				}

				mock.On("NamespaceList", ctx, query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, false).Return(namespaces, len(namespaces), nil).Once()
				mock.On("UserGetByID", ctx, "hash", false).Return(nil, 0, errors.New("error")).Once()
			},
			expected: Expected{
				namespaces: nil,
				count:      0,
				err:        NewErrNamespaceMemberFillData(NewErrUserNotFound("hash", errors.New("error"))),
			},
		},
		{
			description: "success to get the namespace list",
			paginator:   query.Paginator{Page: 1, PerPage: 10},
			filters:     query.Filters{},
			ctx:         ctx,
			requiredMocks: func() {
				namespaces := []models.Namespace{
					{
						Name:     "group1",
						Owner:    "hash",
						TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
						Members: []models.Member{
							{
								ID:   "hash",
								Role: guard.RoleOwner,
							},
						},
					},
					{
						Name:     "group2",
						Owner:    "hash",
						TenantID: "a736a52b-5777-4f92-b0b8-e359bf48471i4",
						Members: []models.Member{
							{
								ID:   "hash",
								Role: guard.RoleOwner,
							},
							{
								ID:   "hash2",
								Role: guard.RoleObserver,
							},
						},
					},
				}

				user := &models.User{
					UserData: models.UserData{
						Name:     "user",
						Username: "hash",
					},
					ID: "hash",
				}

				user1 := &models.User{
					UserData: models.UserData{
						Name:     "user2",
						Username: "hash2",
					},
					ID: "hash2",
				}

				// TODO: Add mock to fillMembersData what will replace the three call to UserGetByID.
				mock.On("NamespaceList", ctx, query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, false).Return(namespaces, len(namespaces), nil).Once()
				mock.On("UserGetByID", ctx, "hash", false).Return(user, 0, nil).Once()
				mock.On("UserGetByID", ctx, "hash2", false).Return(user1, 0, nil).Once()
				mock.On("UserGetByID", ctx, "hash", false).Return(user, 0, nil).Once()
			},
			expected: Expected{
				namespaces: []models.Namespace{
					{
						Name: "group1", Owner: "hash", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
						Members: []models.Member{
							{
								ID:       "hash",
								Username: "hash",
								Role:     guard.RoleOwner,
							},
						},
					},
					{Name: "group2", Owner: "hash", TenantID: "a736a52b-5777-4f92-b0b8-e359bf48471i4", Members: []models.Member{
						{
							ID:       "hash",
							Username: "hash",
							Role:     guard.RoleOwner,
						},
						{
							ID:       "hash2",
							Username: "hash2",
							Role:     guard.RoleObserver,
						},
					}},
				},
				count: len([]models.Namespace{
					{
						Name: "group1", Owner: "hash", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
						Members: []models.Member{
							{
								ID:       "hash",
								Username: "hash",
								Role:     guard.RoleOwner,
							},
						},
					},
					{Name: "group2", Owner: "hash", TenantID: "a736a52b-5777-4f92-b0b8-e359bf48471i4", Members: []models.Member{
						{
							ID:       "hash",
							Username: "hash",
							Role:     guard.RoleOwner,
						},
						{
							ID:       "hash2",
							Username: "hash2",
							Role:     guard.RoleObserver,
						},
					}},
				}),
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			services := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
			nss, count, err := services.ListNamespaces(tc.ctx, tc.paginator, tc.filters, false)
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
		user          *models.User
		namespace     *models.Namespace
		ctx           context.Context
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when could not get the namespace",
			ctx:         ctx,
			user:        &models.User{UserData: models.UserData{Name: "user1", Username: "hash1"}, ID: "hash1"},
			namespace: &models.Namespace{
				Name:     "group1",
				Owner:    "hash1",
				TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
				Members: []models.Member{
					{
						ID:       "hash1",
						Username: "hash1",
						Role:     guard.RoleOwner,
					},
				},
			},
			requiredMocks: func() {
				namespace := &models.Namespace{
					Name:     "group1",
					Owner:    "hash1",
					TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
					Members: []models.Member{
						{
							ID:       "hash1",
							Username: "hash1",
							Role:     guard.RoleOwner,
						},
					},
				}

				mock.On("NamespaceGet", ctx, namespace.TenantID, true).Return(nil, errors.New("error")).Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceNotFound("a736a52b-5777-4f92-b0b8-e359bf484713", errors.New("error")),
			},
		},
		{
			description: "succeeds",
			ctx:         ctx,
			namespace: &models.Namespace{
				Name:     "group1",
				Owner:    "hash1",
				TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
				Members: []models.Member{
					{
						ID:       "hash1",
						Username: "hash1",
						Role:     guard.RoleOwner,
					},
				},
			},
			user: &models.User{
				UserData: models.UserData{
					Name:     "user1",
					Username: "hash1",
				},
				ID: "hash1",
			},
			requiredMocks: func() {
				namespace := &models.Namespace{
					Name:     "group1",
					Owner:    "hash1",
					TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
					Members: []models.Member{
						{
							ID:       "hash1",
							Username: "hash1",
							Role:     guard.RoleOwner,
						},
					},
				}

				user := &models.User{
					UserData: models.UserData{
						Name:     "user1",
						Username: "hash1",
					},
					ID: "hash1",
				}

				mock.On("NamespaceGet", ctx, namespace.TenantID, true).Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
			},
			expected: Expected{
				namespace: &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Members: []models.Member{{ID: "hash1", Username: "hash1", Role: guard.RoleOwner}}},
				err:       nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
			returnNamespace, err := service.GetNamespace(ctx, tc.namespace.TenantID)
			assert.Equal(t, tc.expected, Expected{returnNamespace, err})
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
			description: "fails when user is not found",
			members: []models.Member{
				{ID: "hash1", Role: guard.RoleObserver},
				{ID: "hash2", Role: guard.RoleObserver},
				{ID: "hash3", Role: guard.RoleObserver},
			},
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, "hash1", false).Return(nil, 0, errors.New("error")).Once()
			},
			expected: Expected{
				members: nil,
				err:     NewErrUserNotFound("hash1", errors.New("error")),
			},
		},
		{
			description: "success to fill member data",
			members: []models.Member{
				{ID: "hash1", Role: guard.RoleObserver},
				{ID: "hash2", Role: guard.RoleObserver},
				{ID: "hash3", Role: guard.RoleObserver},
				{ID: "hash4", Role: guard.RoleOwner},
			},
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, "hash1", false).Return(&models.User{ID: "hash1", UserData: models.UserData{Username: "username1"}}, 0, nil).Once()
				mock.On("UserGetByID", ctx, "hash2", false).Return(&models.User{ID: "hash2", UserData: models.UserData{Username: "username2"}}, 0, nil).Once()
				mock.On("UserGetByID", ctx, "hash3", false).Return(&models.User{ID: "hash3", UserData: models.UserData{Username: "username3"}}, 0, nil).Once()
				mock.On("UserGetByID", ctx, "hash4", false).Return(&models.User{ID: "hash4", UserData: models.UserData{Username: "username4"}}, 0, nil).Once()
			},
			expected: Expected{
				members: []models.Member{
					{ID: "hash1", Username: "username1", Role: guard.RoleObserver},
					{ID: "hash2", Username: "username2", Role: guard.RoleObserver},
					{ID: "hash3", Username: "username3", Role: guard.RoleObserver},
					{ID: "hash4", Username: "username4", Role: guard.RoleOwner},
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			services := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
			members, err := services.fillMembersData(ctx, tc.members)
			assert.Equal(t, tc.expected, Expected{members, err})
		})
	}
}

func TestCreateNamespace(t *testing.T) {
	mock := new(mocks.Store)

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
		ownerID       string
		namespace     requests.NamespaceCreate
		expected      Expected
	}{
		{
			description: "fails when store user get has no documents",
			ownerID:     "hash1",
			namespace: requests.NamespaceCreate{
				Name:     "namespace",
				TenantID: "xxxxx",
			},
			requiredMocks: func() {
				user := &models.User{
					UserData: models.UserData{
						Name:     "user1",
						Username: "hash1",
					},
					ID: "hash1",
				}

				mock.On("UserGetByID", ctx, user.ID, false).Return(nil, 0, store.ErrNoDocuments).Once()
			},
			expected: Expected{
				nil,
				NewErrUserNotFound("hash1", store.ErrNoDocuments),
			},
		},
		{
			description: "fails when store user get fails",
			ownerID:     "hash1",
			namespace: requests.NamespaceCreate{
				Name:     "namespace",
				TenantID: "xxxxx",
			},
			requiredMocks: func() {
				user := &models.User{
					UserData: models.UserData{
						Name:     "user1",
						Username: "hash1",
					},
					ID: "hash1",
				}

				mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, errors.New("error")).Once()
			},
			expected: Expected{
				nil,
				NewErrUserNotFound("hash1", errors.New("error")),
			},
		},
		{
			description: "fails when a namespace field is invalid",
			ownerID:     "hash1",
			namespace: requests.NamespaceCreate{
				Name: "name.with.dot",
			},
			requiredMocks: func() {
				user := &models.User{
					UserData: models.UserData{
						Name:     "user1",
						Username: "hash1",
					},
					ID: "hash1",
				}

				mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
			},
			expected: Expected{
				nil,
				NewErrNamespaceInvalid(validator.ErrStructureInvalid),
			},
		},
		{
			description: "fails when a namespace already exists",
			ownerID:     "hash1",
			namespace: requests.NamespaceCreate{
				Name:     "namespace",
				TenantID: "xxxxx",
			},
			requiredMocks: func() {
				user := &models.User{
					UserData: models.UserData{
						Name:     "user1",
						Username: "hash1",
					},
					ID: "hash1",
				}

				model := &models.Namespace{
					Name:  strings.ToLower("namespace"),
					Owner: "hash1",
					Members: []models.Member{
						{ID: "hash1", Role: guard.RoleOwner},
					},
					Settings: &models.NamespaceSettings{
						SessionRecord: true,
					},
					TenantID: "xxxxx",
				}

				var isCloud bool
				mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return(strconv.FormatBool(isCloud)).Once()
				mock.On("NamespaceGetByName", ctx, "namespace").Return(model, nil).Once()
			},
			expected: Expected{
				nil,
				NewErrNamespaceDuplicated(nil),
			},
		},
		{
			description: "fails when store get namespace by name fails",
			ownerID:     "hash1",
			namespace: requests.NamespaceCreate{
				Name:     "namespace",
				TenantID: "xxxxx",
			},
			requiredMocks: func() {
				user := &models.User{
					UserData: models.UserData{
						Name:     "user1",
						Username: "hash1",
					},
					ID: "hash1",
				}

				model := &models.Namespace{
					Name:  strings.ToLower("namespace"),
					Owner: "hash1",
					Members: []models.Member{
						{ID: "hash1", Role: guard.RoleOwner},
					},
					Settings: &models.NamespaceSettings{
						SessionRecord: true,
					},
					TenantID: "xxxxx",
				}

				var isCloud bool
				mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return(strconv.FormatBool(isCloud)).Once()
				mock.On("NamespaceGetByName", ctx, "namespace").Return(model, errors.New("error")).Once()
			},
			expected: Expected{
				nil,
				NewErrNamespaceNotFound("namespace", errors.New("error")),
			},
		},
		{
			description: "fails when store namespace create fails",
			ownerID:     "hash1",
			namespace: requests.NamespaceCreate{
				Name:     "namespace",
				TenantID: "xxxxx",
			},
			requiredMocks: func() {
				user := &models.User{
					UserData: models.UserData{
						Name:     "user1",
						Username: "hash1",
					},
					ID: "hash1",
				}

				var isCloud bool
				notCloudNamespace := &models.Namespace{
					Name:  strings.ToLower("namespace"),
					Owner: "hash1",
					Members: []models.Member{
						{ID: "hash1", Role: guard.RoleOwner},
					},
					Settings:   &models.NamespaceSettings{SessionRecord: true},
					TenantID:   "xxxxx",
					MaxDevices: -1,
				}
				mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
				mock.On("NamespaceGetByName", ctx, "namespace").Return(nil, nil).Once()
				mock.On("NamespaceCreate", ctx, notCloudNamespace).Return(nil, errors.New("error")).Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return(strconv.FormatBool(isCloud)).Once()
			},
			expected: Expected{
				nil, NewErrNamespaceCreateStore(errors.New("error")),
			},
		},
		{
			description: "generates namespace with random tenant",
			ownerID:     "hash1",
			namespace: requests.NamespaceCreate{
				Name: "namespace",
			},
			requiredMocks: func() {
				user := &models.User{
					UserData: models.UserData{
						Name:     "user1",
						Username: "hash1",
					},
					ID: "hash1",
				}

				var isCloud bool
				notCloudNamespace := &models.Namespace{
					Name:  strings.ToLower("namespace"),
					Owner: "hash1",
					Members: []models.Member{
						{ID: "hash1", Role: guard.RoleOwner},
					},
					Settings:   &models.NamespaceSettings{SessionRecord: true},
					TenantID:   "random_uuid",
					MaxDevices: -1,
				}
				mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
				uuidMock.On("Generate").Return("random_uuid").Once()
				mock.On("NamespaceGetByName", ctx, "namespace").Return(nil, nil).Once()
				mock.On("NamespaceCreate", ctx, notCloudNamespace).Return(notCloudNamespace, nil).Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return(strconv.FormatBool(isCloud)).Once()
			},
			expected: Expected{
				&models.Namespace{
					Name:  strings.ToLower("namespace"),
					Owner: "hash1",
					Members: []models.Member{
						{ID: "hash1", Role: guard.RoleOwner},
					},
					Settings:   &models.NamespaceSettings{SessionRecord: true},
					TenantID:   "random_uuid",
					MaxDevices: -1,
				}, nil,
			},
		},
		{
			description: "checks the enterprise&community instance",
			ownerID:     "hash1",
			namespace: requests.NamespaceCreate{
				Name:     "namespace",
				TenantID: "xxxxx",
			},
			requiredMocks: func() {
				user := &models.User{
					UserData: models.UserData{
						Name:     "user1",
						Username: "hash1",
					},
					ID: "hash1",
				}

				var isCloud bool
				notCloudNamespace := &models.Namespace{
					Name:  strings.ToLower("namespace"),
					Owner: "hash1",
					Members: []models.Member{
						{ID: "hash1", Role: guard.RoleOwner},
					},
					Settings:   &models.NamespaceSettings{SessionRecord: true},
					TenantID:   "xxxxx",
					MaxDevices: -1,
				}
				mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
				mock.On("NamespaceGetByName", ctx, "namespace").Return(nil, nil).Once()
				mock.On("NamespaceCreate", ctx, notCloudNamespace).Return(nil, nil).Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return(strconv.FormatBool(isCloud)).Once()
			},
			expected: Expected{
				&models.Namespace{
					Name:  strings.ToLower("namespace"),
					Owner: "hash1",
					Members: []models.Member{
						{ID: "hash1", Role: guard.RoleOwner},
					},
					Settings:   &models.NamespaceSettings{SessionRecord: true},
					TenantID:   "xxxxx",
					MaxDevices: -1,
				}, nil,
			},
		},
		{
			description: "checks the cloud instance",
			ownerID:     "hash1",
			namespace: requests.NamespaceCreate{
				Name:     "namespace",
				TenantID: "xxxxx",
			},
			requiredMocks: func() {
				user := &models.User{
					UserData: models.UserData{
						Name:     "user1",
						Username: "hash1",
					},
					ID: "hash1",
				}

				isCloud := true
				cloudNamespace := &models.Namespace{
					Name:  strings.ToLower("namespace"),
					Owner: "hash1",
					Members: []models.Member{
						{ID: "hash1", Role: guard.RoleOwner},
					},
					Settings:   &models.NamespaceSettings{SessionRecord: true},
					TenantID:   "xxxxx",
					MaxDevices: 3,
				}
				mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
				mock.On("NamespaceGetByName", ctx, "namespace").Return(nil, nil).Once()
				mock.On("NamespaceCreate", ctx, cloudNamespace).Return(nil, nil).Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return(strconv.FormatBool(isCloud)).Once()
			},
			expected: Expected{
				&models.Namespace{
					Name:  strings.ToLower("namespace"),
					Owner: "hash1",
					Members: []models.Member{
						{ID: "hash1", Role: guard.RoleOwner},
					},
					Settings:   &models.NamespaceSettings{SessionRecord: true},
					TenantID:   "xxxxx",
					MaxDevices: 3,
				}, nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
			returnedNamespace, err := service.CreateNamespace(ctx, tc.namespace, tc.ownerID)
			assert.Equal(t, tc.expected, Expected{returnedNamespace, err})
		})
	}
	mock.AssertExpectations(t)
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

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

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
			namespace:   &models.Namespace{Name: "oldname", Owner: "ID1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Members: []models.Member{{ID: "user1", Role: guard.RoleOwner}}},
			requiredMocks: func(namespace *models.Namespace) {
				mock.On("NamespaceGet", ctx, namespace.TenantID, true).Return(nil, errors.New("error")).Once()
			},
			expected: NewErrNamespaceNotFound("a736a52b-5777-4f92-b0b8-e359bf484713", errors.New("error")),
		},
		{
			description: "fails when store delete fails",
			namespace:   &models.Namespace{Name: "oldname", Owner: "ID1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Members: []models.Member{{ID: "user1", Role: guard.RoleOwner}}},
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
			namespace:   &models.Namespace{Name: "oldname", Owner: "ID1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Members: []models.Member{{ID: "user1", Role: guard.RoleOwner}}},
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
			namespace:   &models.Namespace{Name: "oldname", Owner: "ID1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Members: []models.Member{{ID: "user1", Role: guard.RoleOwner}}},
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
						{ID: user1.ID, Role: guard.RoleOwner},
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

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
			err := service.DeleteNamespace(ctx, tc.namespace.TenantID)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestAddNamespaceUser(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	type Expected struct {
		namespace *models.Namespace
		err       error
	}

	cases := []struct {
		description   string
		TenantID      string
		Username      string
		Role          string
		ID            string
		RequiredMocks func()
		Expected      Expected
	}{
		{
			description: "fails when MemberID is not valid",
			Username:    "",
			Role:        guard.RoleObserver,
			ID:          "ID1",
			TenantID:    "a736a52b-5777-4f92-b0b8-e359bf484713",
			RequiredMocks: func() {
			},
			Expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceMemberInvalid(validator.ErrStructureInvalid),
			},
		},
		{
			description: "fails when Role is not valid",
			Username:    "user2",
			Role:        "invalid",
			ID:          "ID1",
			TenantID:    "a736a52b-5777-4f92-b0b8-e359bf484713",
			RequiredMocks: func() {
			},
			Expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceMemberInvalid(validator.ErrStructureInvalid),
			},
		},
		{
			description: "fails when the namespace was not found",
			Username:    "user2",
			Role:        guard.RoleObserver,
			ID:          "ID1",
			TenantID:    "tenantIDNotFound",
			RequiredMocks: func() {
				mock.On("NamespaceGet", ctx, "tenantIDNotFound", true).Return(nil, ErrNamespaceNotFound).Once()
			},
			Expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceNotFound("tenantIDNotFound", ErrNamespaceNotFound),
			},
		},
		{
			description: "fails when the active member was not found",
			Username:    "user1",
			Role:        guard.RoleObserver,
			ID:          "userIDNotFound",
			TenantID:    "a736a52b-5777-4f92-b0b8-e359bf484713",
			RequiredMocks: func() {
				namespace := &models.Namespace{
					Name:     "group1",
					Owner:    "ID1",
					TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
					Members: []models.Member{
						{ID: "ID1", Role: guard.RoleOwner},
					},
				}

				mock.On("NamespaceGet", ctx, namespace.TenantID, true).Return(namespace, nil).Once()

				mock.On("UserGetByID", ctx, "userIDNotFound", false).Return(nil, 0, ErrUserNotFound).Once()
			},
			Expected: Expected{
				namespace: nil,
				err:       NewErrUserNotFound("userIDNotFound", ErrUserNotFound),
			},
		},
		{
			description: "fails when the active member is not on the namespace",
			Username:    "user1",
			Role:        guard.RoleObserver,
			ID:          "ID2",
			TenantID:    "a736a52b-5777-4f92-b0b8-e359bf484713",
			RequiredMocks: func() {
				namespace := &models.Namespace{
					Name:     "group1",
					Owner:    "ID1",
					TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
					Members: []models.Member{
						{ID: "ID1", Role: guard.RoleOwner},
					},
				}

				user2 := &models.User{
					UserData: models.UserData{
						Name:     "user2",
						Username: "user2",
						Email:    "user2@email.com",
					},
					ID: "ID2",
				}

				mock.On("NamespaceGet", ctx, namespace.TenantID, true).Return(namespace, nil).Once()

				mock.On("UserGetByID", ctx, user2.ID, false).Return(user2, 0, nil).Once()
			},
			Expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceMemberNotFound("ID2", nil),
			},
		},
		{
			description: "addNamespaceUser fails when passive member was not found",
			Username:    "usernamespacenotfound",
			Role:        guard.RoleObserver,
			ID:          "ID1",
			TenantID:    "a736a52b-5777-4f92-b0b8-e359bf484713",
			RequiredMocks: func() {
				namespace := &models.Namespace{
					Name:     "group1",
					Owner:    "ID1",
					TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
					Members: []models.Member{
						{ID: "ID1", Role: guard.RoleOwner},
					},
				}

				user1 := &models.User{
					UserData: models.UserData{
						Name:     "user1",
						Username: "user1",
						Email:    "user1@email.com",
					},
					ID: "ID1",
				}

				mock.On("NamespaceGet", ctx, namespace.TenantID, true).Return(namespace, nil).Once()

				mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
				mock.On("UserGetByUsername", ctx, "usernamespacenotfound").Return(nil, ErrBadRequest).Once()
			},
			Expected: Expected{
				namespace: nil,
				err:       NewErrUserNotFound("usernamespacenotfound", ErrBadRequest),
			},
		},
		{
			description: "fails when the passive member is on the namespace",
			Username:    "user2",
			Role:        guard.RoleObserver,
			ID:          "ID1",
			TenantID:    "a736a52b-5777-4f92-b0b8-e359bf484714",
			RequiredMocks: func() {
				user2 := &models.User{
					UserData: models.UserData{
						Name:     "user2",
						Username: "user2",
						Email:    "user2@email.com",
					},
					ID: "ID2",
				}

				user1 := &models.User{
					UserData: models.UserData{
						Name:     "user1",
						Username: "user1",
						Email:    "user1@email.com",
					},
					ID: "ID1",
				}

				namespaceTwoMembers := &models.Namespace{
					Name:     "group1",
					Owner:    "ID1",
					TenantID: "a736a52b-5777-4f92-b0b8-e359bf484714",
					Members: []models.Member{
						{ID: "ID1", Role: guard.RoleOwner},
						{ID: "ID2", Role: guard.RoleObserver},
					},
				}

				mock.On("NamespaceGet", ctx, namespaceTwoMembers.TenantID, true).Return(namespaceTwoMembers, nil).Once()

				mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
				mock.On("UserGetByUsername", ctx, user2.Username).Return(user2, nil).Once()
			},
			Expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceMemberDuplicated("ID2", nil),
			},
		},
		{
			description: "succeeds",
			Username:    "user2",
			Role:        guard.RoleObserver,
			ID:          "ID1",
			TenantID:    "a736a52b-5777-4f92-b0b8-e359bf484713",
			RequiredMocks: func() {
				namespace := &models.Namespace{
					Name:     "group1",
					Owner:    "ID1",
					TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
					Members: []models.Member{
						{ID: "ID1", Role: guard.RoleOwner},
					},
				}

				user1 := &models.User{
					UserData: models.UserData{
						Name:     "user1",
						Username: "user1",
						Email:    "user1@email.com",
					},
					ID: "ID1",
				}

				user2 := &models.User{
					UserData: models.UserData{
						Name:     "user2",
						Username: "user2",
						Email:    "user2@email.com",
					},
					ID: "ID2",
				}

				namespaceTwoMembers := &models.Namespace{
					Name:     "group1",
					Owner:    "ID1",
					TenantID: "a736a52b-5777-4f92-b0b8-e359bf484714",
					Members: []models.Member{
						{ID: "ID1", Role: guard.RoleOwner},
						{ID: "ID2", Role: guard.RoleObserver},
					},
				}

				mock.On("NamespaceGet", ctx, namespace.TenantID, true).Return(namespace, nil).Once()

				mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
				mock.On("UserGetByUsername", ctx, user2.Username).Return(user2, nil).Once()

				mock.On("NamespaceAddMember", ctx, namespace.TenantID, user2.ID, guard.RoleObserver).Return(namespaceTwoMembers, nil).Once()
			},
			Expected: Expected{
				namespace: &models.Namespace{Name: "group1", Owner: "ID1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484714", Members: []models.Member{{ID: "ID1", Role: guard.RoleOwner}, {ID: "ID2", Role: guard.RoleObserver}}},
				err:       nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.RequiredMocks()

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
			ns, err := service.AddNamespaceUser(ctx, tc.Username, tc.Role, tc.TenantID, tc.ID)
			assert.Equal(t, tc.Expected, Expected{ns, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestRemoveNamespaceUser(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	type Expected struct {
		namespace *models.Namespace
		err       error
	}

	cases := []struct {
		description   string
		RequiredMocks func()
		TenantID      string
		UserID        string
		MemberID      string
		Expected      Expected
	}{
		{
			description: "fails when namespace was not found",
			RequiredMocks: func() {
				mock.On("NamespaceGet", ctx, "tenantIDNotFound", true).Return(nil, errors.New("error")).Once()
			},
			TenantID: "tenantIDNotFound",
			MemberID: "hash1",
			UserID:   "hash1",
			Expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceNotFound("tenantIDNotFound", errors.New("error")),
			},
		},
		{
			description: "fails when active user was not found",
			RequiredMocks: func() {
				namespace := &models.Namespace{
					Name:     "group1",
					Owner:    "hash1",
					TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
					Members: []models.Member{
						{ID: "hash1", Role: guard.RoleOwner},
					},
				}

				mock.On("NamespaceGet", ctx, namespace.TenantID, true).Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, "invalidUserID", false).Return(nil, 0, ErrUserNotFound).Once()
			},
			TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
			MemberID: "hash2",
			UserID:   "invalidUserID",
			Expected: Expected{
				namespace: nil,
				err:       NewErrUserNotFound("invalidUserID", ErrUserNotFound),
			},
		},
		{
			description: "fails when passive user was not found",
			RequiredMocks: func() {
				namespace := &models.Namespace{
					Name:     "group1",
					Owner:    "hash1",
					TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
					Members: []models.Member{
						{ID: "hash1", Role: guard.RoleOwner},
					},
				}

				user := &models.User{
					UserData: models.UserData{
						Name:     "user1",
						Username: "username1",
					},
					ID: "hash1",
				}

				mock.On("NamespaceGet", ctx, namespace.TenantID, true).Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, namespace.Owner, false).Return(user, 0, nil).Once()
				mock.On("UserGetByID", ctx, "invalidPassiveMemberID", false).Return(nil, 0, ErrUserNotFound).Once()
			},
			TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
			MemberID: "invalidPassiveMemberID",
			UserID:   "hash1",
			Expected: Expected{
				namespace: nil,
				err:       NewErrUserNotFound("invalidPassiveMemberID", ErrUserNotFound),
			},
		},
		{
			description: "fails when user is not a namespace's member",
			RequiredMocks: func() {
				namespace := &models.Namespace{
					Name:     "group1",
					Owner:    "hash1",
					TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
					Members: []models.Member{
						{ID: "hash1", Role: guard.RoleOwner},
					},
				}

				user := &models.User{
					UserData: models.UserData{
						Name:     "user1",
						Username: "username1",
					},
					ID: "hash1",
				}

				user2 := &models.User{
					UserData: models.UserData{
						Name:     "user2",
						Username: "username2",
					},
					ID: "hash2",
				}

				mock.On("NamespaceGet", ctx, namespace.TenantID, true).Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, namespace.Owner, false).Return(user, 0, nil).Once()
				mock.On("UserGetByID", ctx, user2.ID, false).Return(user2, 0, nil).Once()
			},
			TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
			MemberID: "hash2",
			UserID:   "hash1",
			Expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceMemberNotFound("hash2", nil),
			},
		},
		{
			description: "fails when user can not act over the role",
			RequiredMocks: func() {
				user2 := &models.User{
					UserData: models.UserData{
						Name:     "user2",
						Username: "username2",
					},
					ID: "hash2",
				}

				user3 := &models.User{
					UserData: models.UserData{
						Name:     "user3",
						Username: "username3",
					},
					ID: "hash3",
				}

				namespaceThreeMembers := &models.Namespace{
					Name:     "group2",
					Owner:    "hash1",
					TenantID: "a736a52b-5777-4f92-b0b8-e359bf484714",
					Members: []models.Member{
						{
							ID: "hash1", Role: guard.RoleOwner,
						},
						{
							ID: "hash2", Role: guard.RoleAdministrator,
						},
						{
							ID: "hash3", Role: guard.RoleAdministrator,
						},
					},
				}

				mock.On("NamespaceGet", ctx, namespaceThreeMembers.TenantID, true).Return(namespaceThreeMembers, nil).Once()
				mock.On("UserGetByID", ctx, user2.ID, false).Return(user2, 0, nil).Once()
				mock.On("UserGetByID", ctx, user3.ID, false).Return(user3, 0, nil).Once()
			},
			TenantID: "a736a52b-5777-4f92-b0b8-e359bf484714",
			MemberID: "hash3",
			UserID:   "hash2",
			Expected: Expected{
				namespace: nil,
				err:       guard.ErrForbidden,
			},
		},
		{
			description: "when NamespaceRemoveMember store's function fails",
			RequiredMocks: func() {
				namespace := &models.Namespace{
					Name:     "group1",
					Owner:    "hash1",
					TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
					Members: []models.Member{
						{ID: "hash1", Role: guard.RoleOwner},
					},
				}

				user := &models.User{
					UserData: models.UserData{
						Name:     "user1",
						Username: "username1",
					},
					ID: "hash1",
				}

				user2 := &models.User{
					UserData: models.UserData{
						Name:     "user2",
						Username: "username2",
					},
					ID: "hash2",
				}

				namespaceTwoMembers := &models.Namespace{
					Name:     "group2",
					Owner:    "hash1",
					TenantID: "a736a52b-5777-4f92-b0b8-e359bf484714",
					Members: []models.Member{
						{ID: "hash1", Role: guard.RoleOwner},
						{ID: "hash2", Role: guard.RoleObserver},
					},
				}

				mock.On("NamespaceGet", ctx, namespaceTwoMembers.TenantID, true).Return(namespaceTwoMembers, nil).Once()
				mock.On("UserGetByID", ctx, namespace.Owner, false).Return(user, 0, nil).Once()
				mock.On("UserGetByID", ctx, user2.ID, false).Return(user2, 0, nil).Once()

				mock.On("NamespaceRemoveMember", ctx, namespaceTwoMembers.TenantID, user2.ID).Return(nil, errors.New("error")).Once()
			},
			TenantID: "a736a52b-5777-4f92-b0b8-e359bf484714",
			MemberID: "hash2",
			UserID:   "hash1",
			Expected: Expected{
				namespace: nil,
				err:       errors.New("error"),
			},
		},
		{
			description: "succeeds",
			RequiredMocks: func() {
				namespace := &models.Namespace{
					Name:     "group1",
					Owner:    "hash1",
					TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
					Members: []models.Member{
						{ID: "hash1", Role: guard.RoleOwner},
					},
				}

				user := &models.User{
					UserData: models.UserData{
						Name:     "user1",
						Username: "username1",
					},
					ID: "hash1",
				}

				user2 := &models.User{
					UserData: models.UserData{
						Name:     "user2",
						Username: "username2",
					},
					ID: "hash2",
				}

				namespaceTwoMembers := &models.Namespace{
					Name:     "group2",
					Owner:    "hash1",
					TenantID: "a736a52b-5777-4f92-b0b8-e359bf484714",
					Members: []models.Member{
						{ID: "hash1", Role: guard.RoleOwner},
						{ID: "hash2", Role: guard.RoleObserver},
					},
				}

				mock.On("NamespaceGet", ctx, namespaceTwoMembers.TenantID, true).Return(namespaceTwoMembers, nil).Once()
				mock.On("UserGetByID", ctx, namespace.Owner, false).Return(user, 0, nil).Once()
				mock.On("UserGetByID", ctx, user2.ID, false).Return(user2, 0, nil).Once()

				mock.On("NamespaceRemoveMember", ctx, namespaceTwoMembers.TenantID, user2.ID).Return(namespace, nil).Once()
			},
			TenantID: "a736a52b-5777-4f92-b0b8-e359bf484714",
			MemberID: "hash2",
			UserID:   "hash1",
			Expected: Expected{
				namespace: &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Members: []models.Member{{ID: "hash1", Role: guard.RoleOwner}}},
				err:       nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.RequiredMocks()

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
			ns, err := service.RemoveNamespaceUser(ctx, tc.TenantID, tc.MemberID, tc.UserID)
			assert.Equal(t, tc.Expected, Expected{ns, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestEditNamespaceUser(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	cases := []struct {
		description   string
		TenantID      string
		UserID        string
		MemberID      string
		MemberNewRole string
		RequiredMocks func()
		Expected      error
	}{
		{
			description:   "fails when namespace was not found",
			TenantID:      "tenantIDNotFound",
			UserID:        "activeMemberID",
			MemberID:      "passiveMemberID",
			MemberNewRole: guard.RoleObserver,
			RequiredMocks: func() {
				mock.On("NamespaceGet", ctx, "tenantIDNotFound", true).Return(nil, errors.New("error")).Once()
			},
			Expected: NewErrNamespaceNotFound("tenantIDNotFound", errors.New("error")),
		},
		{
			description:   "fails when active member was not found",
			TenantID:      "a736a52b-5777-4f92-b0b8-e359bf484717",
			UserID:        "invalidMemberActiveID",
			MemberID:      "passiveMemberID",
			MemberNewRole: guard.RoleObserver,
			RequiredMocks: func() {
				namespaceActivePassive := &models.Namespace{
					Name:     "group1",
					Owner:    "activeMemberID",
					TenantID: "a736a52b-5777-4f92-b0b8-e359bf484717",
					Members: []models.Member{
						{ID: "ownerID", Role: guard.RoleOwner},
						{ID: "activeMemberID", Role: guard.RoleAdministrator},
						{ID: "passiveMemberID", Role: guard.RoleObserver},
					},
				}

				mock.On("NamespaceGet", ctx, namespaceActivePassive.TenantID, true).Return(namespaceActivePassive, nil).Once()

				mock.On("UserGetByID", ctx, "invalidMemberActiveID", false).Return(nil, 0, errors.New("error")).Once()
			},
			Expected: NewErrUserNotFound("invalidMemberActiveID", errors.New("error")),
		},
		{
			description:   "fails when passive member was not found",
			TenantID:      "a736a52b-5777-4f92-b0b8-e359bf484713",
			UserID:        "activeMemberID",
			MemberID:      "invalidMemberPassiveID",
			MemberNewRole: guard.RoleObserver,
			RequiredMocks: func() {
				namespaceActiveOwner := &models.Namespace{
					Name:     "group1",
					Owner:    "activeMemberID",
					TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
					Members: []models.Member{
						{ID: "activeMemberID", Role: guard.RoleOwner},
					},
				}

				activeMember := &models.User{
					UserData: models.UserData{
						Name:     "activeMemberName",
						Username: "activeMemberUsername",
					},
					ID: "activeMemberID",
				}

				mock.On("NamespaceGet", ctx, namespaceActiveOwner.TenantID, true).Return(namespaceActiveOwner, nil).Once()

				mock.On("UserGetByID", ctx, activeMember.ID, false).Return(activeMember, 0, nil).Once()
				mock.On("UserGetByID", ctx, "invalidMemberPassiveID", false).Return(nil, 0, errors.New("error")).Once()
			},
			Expected: NewErrUserNotFound("invalidMemberPassiveID", errors.New("error")),
		},
		{
			description:   "fails when could not find passive member inside namespace",
			TenantID:      "a736a52b-5777-4f92-b0b8-e359bf484713",
			UserID:        "activeMemberID",
			MemberID:      "passiveMemberID",
			MemberNewRole: guard.RoleObserver,
			RequiredMocks: func() {
				namespaceActiveOwner := &models.Namespace{
					Name:     "group1",
					Owner:    "activeMemberID",
					TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
					Members: []models.Member{
						{ID: "activeMemberID", Role: guard.RoleOwner},
					},
				}

				activeMember := &models.User{
					UserData: models.UserData{
						Name:     "activeMemberName",
						Username: "activeMemberUsername",
					},
					ID: "activeMemberID",
				}

				passiveMember := &models.User{
					UserData: models.UserData{
						Name:     "passiveMemberName",
						Username: "passiveMemberUsername",
					},
					ID: "passiveMemberID",
				}

				mock.On("NamespaceGet", ctx, namespaceActiveOwner.TenantID, true).Return(namespaceActiveOwner, nil).Once()

				mock.On("UserGetByID", ctx, activeMember.ID, false).Return(activeMember, 0, nil).Once()
				mock.On("UserGetByID", ctx, passiveMember.ID, false).Return(passiveMember, 0, nil).Once()
			},
			Expected: NewErrNamespaceMemberNotFound("passiveMemberID", nil),
		},
		{
			description:   "fails when could not find passive member inside namespace",
			TenantID:      "a736a52b-5777-4f92-b0b8-e359bf484713",
			UserID:        "activeMemberID",
			MemberID:      "passiveMemberID",
			MemberNewRole: guard.RoleObserver,
			RequiredMocks: func() {
				namespaceActiveOwner := &models.Namespace{
					Name:     "group1",
					Owner:    "activeMemberID",
					TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
					Members: []models.Member{
						{ID: "activeMemberID", Role: guard.RoleOwner},
					},
				}

				activeMember := &models.User{
					UserData: models.UserData{
						Name:     "activeMemberName",
						Username: "activeMemberUsername",
					},
					ID: "activeMemberID",
				}

				passiveMember := &models.User{
					UserData: models.UserData{
						Name:     "passiveMemberName",
						Username: "passiveMemberUsername",
					},
					ID: "passiveMemberID",
				}

				mock.On("NamespaceGet", ctx, namespaceActiveOwner.TenantID, true).Return(namespaceActiveOwner, nil).Once()

				mock.On("UserGetByID", ctx, activeMember.ID, false).Return(activeMember, 0, nil).Once()
				mock.On("UserGetByID", ctx, passiveMember.ID, false).Return(passiveMember, 0, nil).Once()
			},
			Expected: NewErrNamespaceMemberNotFound("passiveMemberID", nil),
		},
		{
			description:   "fails when could not find active member inside namespace",
			TenantID:      "a736a52b-5777-4f92-b0b8-e359bf484714",
			UserID:        "activeMemberID",
			MemberID:      "passiveMemberID",
			MemberNewRole: guard.RoleOperator,
			RequiredMocks: func() {
				activeMember := &models.User{
					UserData: models.UserData{
						Name:     "activeMemberName",
						Username: "activeMemberUsername",
					},
					ID: "activeMemberID",
				}

				passiveMember := &models.User{
					UserData: models.UserData{
						Name:     "passiveMemberName",
						Username: "passiveMemberUsername",
					},
					ID: "passiveMemberID",
				}

				namespacePassiveObserver := &models.Namespace{
					Name:     "group1",
					Owner:    "activeMemberID",
					TenantID: "a736a52b-5777-4f92-b0b8-e359bf484714",
					Members: []models.Member{
						{ID: "memberID", Role: guard.RoleOwner},
						{ID: "passiveMemberID", Role: guard.RoleObserver},
					},
				}

				mock.On("NamespaceGet", ctx, namespacePassiveObserver.TenantID, true).Return(namespacePassiveObserver, nil).Once()

				mock.On("UserGetByID", ctx, activeMember.ID, false).Return(activeMember, 0, nil).Once()
				mock.On("UserGetByID", ctx, passiveMember.ID, false).Return(passiveMember, 0, nil).Once()
			},
			Expected: NewErrNamespaceMemberNotFound("activeMemberID", nil),
		},
		{
			description:   "fails when active and passive roles are the same",
			TenantID:      "a736a52b-5777-4f92-b0b8-e359bf484715",
			UserID:        "activeMemberID",
			MemberID:      "passiveMemberID",
			MemberNewRole: guard.RoleOperator,
			RequiredMocks: func() {
				activeMember := &models.User{
					UserData: models.UserData{
						Name:     "activeMemberName",
						Username: "activeMemberUsername",
					},
					ID: "activeMemberID",
				}

				passiveMember := &models.User{
					UserData: models.UserData{
						Name:     "passiveMemberName",
						Username: "passiveMemberUsername",
					},
					ID: "passiveMemberID",
				}

				namespaceActivePassiveSame := &models.Namespace{
					Name:     "group1",
					Owner:    "activeMemberID",
					TenantID: "a736a52b-5777-4f92-b0b8-e359bf484715",
					Members: []models.Member{
						{ID: "ownerID", Role: guard.RoleOwner},
						{ID: "activeMemberID", Role: guard.RoleAdministrator},
						{ID: "passiveMemberID", Role: guard.RoleAdministrator},
					},
				}

				mock.On("NamespaceGet", ctx, namespaceActivePassiveSame.TenantID, true).Return(namespaceActivePassiveSame, nil).Once()

				mock.On("UserGetByID", ctx, passiveMember.ID, false).Return(passiveMember, 0, nil).Once()
				mock.On("UserGetByID", ctx, activeMember.ID, false).Return(activeMember, 0, nil).Once()
			},
			Expected: guard.ErrForbidden,
		},
		{
			description:   "fails when user can not act over the role",
			TenantID:      "a736a52b-5777-4f92-b0b8-e359bf484716",
			UserID:        "activeMemberID",
			MemberID:      "passiveMemberID",
			MemberNewRole: guard.RoleAdministrator,
			RequiredMocks: func() {
				activeMember := &models.User{
					UserData: models.UserData{
						Name:     "activeMemberName",
						Username: "activeMemberUsername",
					},
					ID: "activeMemberID",
				}

				passiveMember := &models.User{
					UserData: models.UserData{
						Name:     "passiveMemberName",
						Username: "passiveMemberUsername",
					},
					ID: "passiveMemberID",
				}

				namespaceActiveHasNoPermission := &models.Namespace{
					Name:     "group1",
					Owner:    "activeMemberID",
					TenantID: "a736a52b-5777-4f92-b0b8-e359bf484716",
					Members: []models.Member{
						{ID: "ownerID", Role: guard.RoleOwner},
						{ID: "activeMemberID", Role: guard.RoleOperator},
						{ID: "passiveMemberID", Role: guard.RoleObserver},
					},
				}

				mock.On("NamespaceGet", ctx, namespaceActiveHasNoPermission.TenantID, true).Return(namespaceActiveHasNoPermission, nil).Once()

				mock.On("UserGetByID", ctx, passiveMember.ID, false).Return(passiveMember, 0, nil).Once()
				mock.On("UserGetByID", ctx, activeMember.ID, false).Return(activeMember, 0, nil).Once()
			},
			Expected: guard.ErrForbidden,
		},
		{
			description:   "fails when user store function fails",
			TenantID:      "a736a52b-5777-4f92-b0b8-e359bf484717",
			UserID:        "activeMemberID",
			MemberID:      "passiveMemberID",
			MemberNewRole: guard.RoleOperator,
			RequiredMocks: func() {
				activeMember := &models.User{
					UserData: models.UserData{
						Name:     "activeMemberName",
						Username: "activeMemberUsername",
					},
					ID: "activeMemberID",
				}

				passiveMember := &models.User{
					UserData: models.UserData{
						Name:     "passiveMemberName",
						Username: "passiveMemberUsername",
					},
					ID: "passiveMemberID",
				}

				namespaceActivePassive := &models.Namespace{
					Name:     "group1",
					Owner:    "activeMemberID",
					TenantID: "a736a52b-5777-4f92-b0b8-e359bf484717",
					Members: []models.Member{
						{ID: "ownerID", Role: guard.RoleOwner},
						{ID: "activeMemberID", Role: guard.RoleAdministrator},
						{ID: "passiveMemberID", Role: guard.RoleObserver},
					},
				}

				mock.On("NamespaceGet", ctx, namespaceActivePassive.TenantID, true).Return(namespaceActivePassive, nil).Once()

				mock.On("UserGetByID", ctx, passiveMember.ID, false).Return(passiveMember, 0, nil).Once()
				mock.On("UserGetByID", ctx, activeMember.ID, false).Return(activeMember, 0, nil).Once()

				mock.On("NamespaceEditMember", ctx, namespaceActivePassive.TenantID, passiveMember.ID, guard.RoleOperator).Return(errors.New("error")).Once()
			},
			Expected: errors.New("error"),
		},
		{
			description:   "Success",
			TenantID:      "a736a52b-5777-4f92-b0b8-e359bf484717",
			UserID:        "activeMemberID",
			MemberID:      "passiveMemberID",
			MemberNewRole: guard.RoleOperator,
			RequiredMocks: func() {
				namespaceActivePassive := &models.Namespace{
					Name:     "group1",
					Owner:    "activeMemberID",
					TenantID: "a736a52b-5777-4f92-b0b8-e359bf484717",
					Members: []models.Member{
						{ID: "ownerID", Role: guard.RoleOwner},
						{ID: "activeMemberID", Role: guard.RoleAdministrator},
						{ID: "passiveMemberID", Role: guard.RoleObserver},
					},
				}

				activeMember := &models.User{
					UserData: models.UserData{
						Name:     "activeMemberName",
						Username: "activeMemberUsername",
					},
					ID: "activeMemberID",
				}

				passiveMember := &models.User{
					UserData: models.UserData{
						Name:     "passiveMemberName",
						Username: "passiveMemberUsername",
					},
					ID: "passiveMemberID",
				}

				mock.On("NamespaceGet", ctx, namespaceActivePassive.TenantID, true).Return(namespaceActivePassive, nil).Once()

				mock.On("UserGetByID", ctx, passiveMember.ID, false).Return(passiveMember, 0, nil).Once()
				mock.On("UserGetByID", ctx, activeMember.ID, false).Return(activeMember, 0, nil).Once()

				mock.On("NamespaceEditMember", ctx, namespaceActivePassive.TenantID, passiveMember.ID, guard.RoleOperator).Return(nil).Once()
			},
			Expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.RequiredMocks()

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
			err := service.EditNamespaceUser(ctx, tc.TenantID, tc.UserID, tc.MemberID, tc.MemberNewRole)
			assert.Equal(t, tc.Expected, err)
		})
	}
	mock.AssertExpectations(t)
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

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
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
						Role: guard.RoleOwner,
					},
					{
						ID:   "hash2",
						Role: guard.RoleObserver,
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
							Role: guard.RoleOwner,
						},
						{
							ID:   "hash2",
							Role: guard.RoleObserver,
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
					Role: guard.RoleOwner,
				},
				{
					ID:   "hash2",
					Role: guard.RoleObserver,
				},
			}},
			requiredMocks: func() {
				namespace := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "xxxx", Settings: &models.NamespaceSettings{SessionRecord: true}, Members: []models.Member{
					{
						ID:   "hash1",
						Role: guard.RoleOwner,
					},
					{
						ID:   "hash2",
						Role: guard.RoleObserver,
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

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
			err := service.EditSessionRecordStatus(ctx, tc.sessionRecord, tc.tenantID)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}
