package services

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"

	storecache "github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/api/request"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	uuid_mocks "github.com/shellhub-io/shellhub/pkg/uuid/mocks"
	"github.com/shellhub-io/shellhub/pkg/validator"
	"github.com/stretchr/testify/assert"
)

func TestListNamespaces(t *testing.T) {
	mock := &mocks.Store{}
	services := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()

	err := errors.New("error")

	user := &models.User{UserData: models.UserData{Name: "user", Username: "hash"}, ID: "hash"}
	user1 := &models.User{UserData: models.UserData{Name: "user2", Username: "hash2"}, ID: "hash2"}

	namespaces := []models.Namespace{
		{
			Name: "group1", Owner: user.ID, TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
			Members: []models.Member{
				{
					ID:   user.ID,
					Role: guard.RoleOwner,
				},
			},
		},
		{Name: "group2", Owner: user.ID, TenantID: "a736a52b-5777-4f92-b0b8-e359bf48471i4", Members: []models.Member{
			{
				ID:   user.ID,
				Role: guard.RoleOwner,
			},
			{
				ID:   user1.ID,
				Role: guard.RoleObserver,
			},
		}},
	}

	namespacesWithUsername := []models.Namespace{
		{
			Name: "group1", Owner: user.ID, TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
			Members: []models.Member{
				{
					ID:       user.ID,
					Username: user.Username,
					Role:     guard.RoleOwner,
				},
			},
		},
		{Name: "group2", Owner: user.ID, TenantID: "a736a52b-5777-4f92-b0b8-e359bf48471i4", Members: []models.Member{
			{
				ID:       user.ID,
				Username: user.Username,
				Role:     guard.RoleOwner,
			},
			{
				ID:       user1.ID,
				Username: user1.Username,
				Role:     guard.RoleObserver,
			},
		}},
	}

	query := paginator.Query{Page: 1, PerPage: 10}

	type Expected struct {
		namespaces []models.Namespace
		count      int
		err        error
	}

	cases := []struct {
		name          string
		pagination    paginator.Query
		filter        string
		ctx           context.Context
		requiredMocks func()
		expected      Expected
	}{
		{
			name:       "fail when could not get the namespace list",
			pagination: query,
			filter:     "",
			ctx:        ctx,
			requiredMocks: func() {
				mock.On("NamespaceList", ctx, query, []models.Filter(nil), false).Return(nil, 0, err).Once()
			},
			expected: Expected{
				namespaces: nil,
				count:      0,
				err:        NewErrNamespaceList(err),
			},
		},
		{
			name:       "fail when could not get a user",
			pagination: query,
			filter:     "",
			ctx:        ctx,
			requiredMocks: func() {
				mock.On("NamespaceList", ctx, query, []models.Filter(nil), false).Return(namespaces, len(namespaces), nil).Once()
				mock.On("UserGetByID", ctx, user.ID, false).Return(nil, 0, err).Once()
			},
			expected: Expected{
				namespaces: nil,
				count:      0,
				err:        NewErrNamespaceMemberFillData(NewErrUserNotFound(user.ID, err)),
			},
		},
		{
			name:       "success to get the namespace list",
			pagination: query,
			filter:     "",
			ctx:        ctx,
			requiredMocks: func() {
				// TODO: Add mock to fillMembersData what will replace the three call to UserGetByID.
				mock.On("NamespaceList", ctx, query, []models.Filter(nil), false).Return(namespaces, len(namespaces), nil).Once()
				mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
				mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
				mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
			},
			expected: Expected{
				namespaces: namespacesWithUsername,
				count:      len(namespacesWithUsername),
				err:        nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			nss, count, err := services.ListNamespaces(tc.ctx, tc.pagination, nil, false)
			assert.Equal(t, tc.expected, Expected{nss, count, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestGetNamespace(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()

	user := &models.User{UserData: models.UserData{Name: "user1", Username: "hash1"}, ID: "hash1"}
	namespace := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Members: []models.Member{{ID: user.ID, Username: user.Username, Role: guard.RoleOwner}}}

	Err := errors.New("error")

	type Expected struct {
		namespace *models.Namespace
		err       error
	}

	cases := []struct {
		name          string
		ctx           context.Context
		tenantID      string
		requiredMocks func()
		expected      Expected
	}{
		{
			name:     "GetNamespace fails when could not get the namespace",
			ctx:      ctx,
			tenantID: namespace.TenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(nil, Err).Once()
			},
			expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceNotFound(namespace.TenantID, Err),
			},
		},
		{
			name:     "GetNamespace succeeds",
			ctx:      ctx,
			tenantID: namespace.TenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
			},
			expected: Expected{
				namespace: namespace,
				err:       nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			returnNamespace, err := s.GetNamespace(ctx, tc.tenantID)
			assert.Equal(t, tc.expected, Expected{returnNamespace, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestSetMemberData(t *testing.T) {
	mock := &mocks.Store{}
	services := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
	ctx := context.TODO()

	err := errors.New("error")

	type Expected struct {
		members []models.Member
		err     error
	}

	cases := []struct {
		name          string
		members       []models.Member
		requiredMocks func()
		expected      Expected
	}{
		{
			name: "fails when user is not found",
			members: []models.Member{
				{ID: "hash1", Role: guard.RoleObserver},
				{ID: "hash2", Role: guard.RoleObserver},
				{ID: "hash3", Role: guard.RoleObserver},
			},
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, "hash1", false).Return(nil, 0, err).Once()
			},
			expected: Expected{
				members: nil,
				err:     NewErrUserNotFound("hash1", err),
			},
		},
		{
			name: "success to fill member data",
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
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			members, err := services.fillMembersData(ctx, tc.members)
			assert.Equal(t, tc.expected, Expected{members, err})
		})
	}
}

func TestCreateNamespace(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()
	uuidMock := &uuid_mocks.Uuid{}
	uuid.DefaultBackend = uuidMock

	type Expected struct {
		ns  *models.Namespace
		err error
	}

	Err := errors.New("error")

	user := &models.User{UserData: models.UserData{Name: "user1", Username: "hash1"}, ID: "hash1"}

	namespace := request.NamespaceCreate{
		Name:     "namespace",
		TenantID: "xxxxx",
	}

	model := &models.Namespace{
		Name:  strings.ToLower("namespace"),
		Owner: user.ID,
		Members: []models.Member{
			{ID: user.ID, Role: guard.RoleOwner},
		},
		Settings: &models.NamespaceSettings{SessionRecord: true},
		TenantID: "xxxxx",
	}

	cases := []struct {
		name          string
		requiredMocks func()
		ownerID       string
		expected      Expected
		namespace     request.NamespaceCreate
	}{
		{
			name:      "CreateNamespace fails when store user get has no documents",
			ownerID:   user.ID,
			namespace: namespace,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, user.ID, false).Return(nil, 0, store.ErrNoDocuments).Once()
			},
			expected: Expected{
				nil,
				NewErrUserNotFound(user.ID, store.ErrNoDocuments),
			},
		},
		{
			name:      "CreateNamespace fails when store user get fails",
			ownerID:   user.ID,
			namespace: namespace,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, Err).Once()
			},
			expected: Expected{
				nil,
				NewErrUserNotFound(user.ID, Err),
			},
		},
		{
			name:    "CreateNamespace fails when a namespace field is invalid",
			ownerID: user.ID,
			namespace: request.NamespaceCreate{
				Name: "name.with.dot",
			},
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
			},
			expected: Expected{
				nil,
				NewErrNamespaceInvalid(validator.ErrInvalidFields),
			},
		},
		{
			name:      "CreateNamespace fails when a namespace already exists",
			ownerID:   user.ID,
			namespace: namespace,
			requiredMocks: func() {
				var isCloud bool
				mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return(strconv.FormatBool(isCloud)).Once()
				mock.On("NamespaceGetByName", ctx, namespace.Name).Return(model, nil).Once()
			},
			expected: Expected{
				nil,
				NewErrNamespaceDuplicated(nil),
			},
		},
		{
			name:      "CreateNamespace fails when store get namespace by name fails",
			ownerID:   user.ID,
			namespace: namespace,
			requiredMocks: func() {
				var isCloud bool
				mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return(strconv.FormatBool(isCloud)).Once()
				mock.On("NamespaceGetByName", ctx, namespace.Name).Return(model, Err).Once()
			},
			expected: Expected{
				nil,
				NewErrNamespaceNotFound(namespace.Name, Err),
			},
		},
		{
			name:      "CreateNamespace fails when store namespace create fails",
			ownerID:   user.ID,
			namespace: namespace,
			requiredMocks: func() {
				var isCloud bool
				notCloudNamespace := &models.Namespace{
					Name:  strings.ToLower("namespace"),
					Owner: user.ID,
					Members: []models.Member{
						{ID: user.ID, Role: guard.RoleOwner},
					},
					Settings:   &models.NamespaceSettings{SessionRecord: true},
					TenantID:   "xxxxx",
					MaxDevices: -1,
				}
				mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
				mock.On("NamespaceGetByName", ctx, namespace.Name).Return(nil, nil).Once()
				mock.On("NamespaceCreate", ctx, notCloudNamespace).Return(nil, Err).Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return(strconv.FormatBool(isCloud)).Once()
			},
			expected: Expected{
				nil, NewErrNamespaceCreateStore(Err),
			},
		},
		{
			name:    "CreateNamespace generates namespace with random tenant",
			ownerID: user.ID,
			namespace: request.NamespaceCreate{
				Name: "namespace",
			},
			requiredMocks: func() {
				var isCloud bool
				notCloudNamespace := &models.Namespace{
					Name:  strings.ToLower("namespace"),
					Owner: user.ID,
					Members: []models.Member{
						{ID: user.ID, Role: guard.RoleOwner},
					},
					Settings:   &models.NamespaceSettings{SessionRecord: true},
					TenantID:   "random_uuid",
					MaxDevices: -1,
				}
				mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
				uuidMock.On("Generate").Return("random_uuid").Once()
				mock.On("NamespaceGetByName", ctx, namespace.Name).Return(nil, nil).Once()
				mock.On("NamespaceCreate", ctx, notCloudNamespace).Return(notCloudNamespace, nil).Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return(strconv.FormatBool(isCloud)).Once()
			},
			expected: Expected{
				&models.Namespace{
					Name:  strings.ToLower("namespace"),
					Owner: user.ID,
					Members: []models.Member{
						{ID: user.ID, Role: guard.RoleOwner},
					},
					Settings:   &models.NamespaceSettings{SessionRecord: true},
					TenantID:   "random_uuid",
					MaxDevices: -1,
				}, nil,
			},
		},
		{
			name:      "CreateNamespace checks the enterprise&community instance",
			ownerID:   user.ID,
			namespace: namespace,
			requiredMocks: func() {
				var isCloud bool
				notCloudNamespace := &models.Namespace{
					Name:  strings.ToLower("namespace"),
					Owner: user.ID,
					Members: []models.Member{
						{ID: user.ID, Role: guard.RoleOwner},
					},
					Settings:   &models.NamespaceSettings{SessionRecord: true},
					TenantID:   "xxxxx",
					MaxDevices: -1,
				}
				mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
				mock.On("NamespaceGetByName", ctx, namespace.Name).Return(nil, nil).Once()
				mock.On("NamespaceCreate", ctx, notCloudNamespace).Return(nil, nil).Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return(strconv.FormatBool(isCloud)).Once()
			},
			expected: Expected{
				&models.Namespace{
					Name:  strings.ToLower(namespace.Name),
					Owner: user.ID,
					Members: []models.Member{
						{ID: user.ID, Role: guard.RoleOwner},
					},
					Settings:   &models.NamespaceSettings{SessionRecord: true},
					TenantID:   namespace.TenantID,
					MaxDevices: -1,
				}, nil,
			},
		},
		{
			name:      "CreateNamespace checks the cloud instance",
			ownerID:   user.ID,
			namespace: namespace,
			requiredMocks: func() {
				isCloud := true
				cloudNamespace := &models.Namespace{
					Name:  strings.ToLower("namespace"),
					Owner: user.ID,
					Members: []models.Member{
						{ID: user.ID, Role: guard.RoleOwner},
					},
					Settings:   &models.NamespaceSettings{SessionRecord: true},
					TenantID:   "xxxxx",
					MaxDevices: 3,
				}
				mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
				mock.On("NamespaceGetByName", ctx, namespace.Name).Return(nil, nil).Once()
				mock.On("NamespaceCreate", ctx, cloudNamespace).Return(nil, nil).Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return(strconv.FormatBool(isCloud)).Once()
			},
			expected: Expected{
				&models.Namespace{
					Name:  strings.ToLower(namespace.Name),
					Owner: user.ID,
					Members: []models.Member{
						{ID: user.ID, Role: guard.RoleOwner},
					},
					Settings:   &models.NamespaceSettings{SessionRecord: true},
					TenantID:   namespace.TenantID,
					MaxDevices: 3,
				}, nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			returnedNamespace, err := s.CreateNamespace(ctx, tc.namespace, tc.ownerID)
			assert.Equal(t, tc.expected, Expected{returnedNamespace, err})
		})
	}
	mock.AssertExpectations(t)
}

func TestEditNamespace(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()

	user := &models.User{UserData: models.UserData{Name: "user1", Username: "hash1"}, ID: "hash1"}
	Err := errors.New("error")

	namespace := &models.Namespace{
		Name:  "oldname",
		Owner: user.ID,
		Members: []models.Member{
			{ID: user.ID, Role: guard.RoleOwner},
		},
		TenantID: "xxxxx",
	}

	type Expected struct {
		namespace *models.Namespace
		err       error
	}

	cases := []struct {
		name          string
		requiredMocks func()
		tenantID      string
		namespaceName string
		ownerID       string
		expected      Expected
	}{
		{
			name: "EditNamespace fails when the name is invalid",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
			},
			tenantID:      "xxxxx",
			namespaceName: "name.with.dot",
			ownerID:       user.ID,
			expected: Expected{
				nil,
				NewErrNamespaceInvalid(validator.ErrInvalidFields),
			},
		},
		{
			name: "EditNamespace fails when the name is the same",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
			},
			tenantID:      "xxxxx",
			namespaceName: namespace.Name,
			ownerID:       user.ID,
			expected: Expected{
				nil,
				NewErrNamespaceDuplicated(nil),
			},
		},
		{
			name: "EditNamespace fails when the store namespace rename fails",
			requiredMocks: func() {
				newName := "newname"
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("NamespaceRename", ctx, namespace.TenantID, newName).Return(nil, Err).Once()
			},
			namespaceName: "newname",
			tenantID:      "xxxxx",
			ownerID:       user.ID,
			expected: Expected{
				nil,
				Err,
			},
		},
		{
			name: "EditNamespace succeeds",
			requiredMocks: func() {
				newName := "newname"
				newNamespace := &models.Namespace{
					Name:     newName,
					Owner:    user.ID,
					TenantID: "xxxxx",
				}
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("NamespaceRename", ctx, namespace.TenantID, newName).Return(newNamespace, nil).Once()
			},
			namespaceName: "newname",
			tenantID:      "xxxxx",
			ownerID:       user.ID,
			expected: Expected{
				&models.Namespace{
					Name:     "newname",
					Owner:    "hash1",
					TenantID: "xxxxx",
				},
				nil,
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			returnedNamespace, err := s.EditNamespace(ctx, tc.tenantID, tc.namespaceName)
			assert.Equal(t, tc.expected, Expected{returnedNamespace, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestDeleteNamespace(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	Err := errors.New("error")

	ctx := context.TODO()

	user1 := &models.User{UserData: models.UserData{Name: "user1", Username: "user1", Email: "user1@email.com"}, ID: "ID1"}
	namespace := &models.Namespace{Name: "oldname", Owner: "ID1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Members: []models.Member{{ID: user1.ID, Role: guard.RoleOwner}}}

	cases := []struct {
		name          string
		tenantID      string
		ownerID       string
		requiredMocks func()
		expected      error
	}{
		{
			name:     "DeleteNamespace fails when namespace does not exist",
			tenantID: namespace.TenantID,
			ownerID:  user1.ID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(nil, Err).Once()
			},
			expected: NewErrNamespaceNotFound(namespace.TenantID, Err),
		},
		{
			name:     "DeleteNamespace fails when store delete fails",
			tenantID: namespace.TenantID,
			ownerID:  user1.ID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("NamespaceDelete", ctx, namespace.TenantID).Return(Err).Once()
			},
			expected: Err,
		},
		{
			name:     "DeleteNamespace succeeds",
			tenantID: namespace.TenantID,
			ownerID:  user1.ID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("NamespaceDelete", ctx, namespace.TenantID).Return(nil).Once()
			},
			expected: nil,
		},
		{
			name:     "DeleteNamespace avoids report for disabled env",
			tenantID: namespace.TenantID,
			ownerID:  user1.ID,
			requiredMocks: func() {
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
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(ns, nil).Once()
				envMock.On("Get", "SHELLHUB_BILLING").Return(strconv.FormatBool(false)).Once()
				mock.On("NamespaceDelete", ctx, namespace.TenantID).Return(nil).Once()
			},
			expected: nil,
		},
		{
			name:     "DeleteNamespace reports delete",
			tenantID: namespace.TenantID,
			ownerID:  user1.ID,
			requiredMocks: func() {
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
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(ns, nil).Once()
				envMock.On("Get", "SHELLHUB_BILLING").Return(strconv.FormatBool(true)).Once()
				clientMock.On("ReportDelete", ns).Return(200, nil).Once()
				mock.On("NamespaceDelete", ctx, namespace.TenantID).Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			err := s.DeleteNamespace(ctx, tc.tenantID)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestAddNamespaceUser(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
	ctx := context.TODO()

	user1 := &models.User{UserData: models.UserData{Name: "user1", Username: "user1", Email: "user1@email.com"}, ID: "ID1"}
	user2 := &models.User{UserData: models.UserData{Name: "user2", Username: "user2", Email: "user2@email.com"}, ID: "ID2"}
	namespace := &models.Namespace{Name: "group1", Owner: "ID1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Members: []models.Member{{ID: user1.ID, Role: guard.RoleOwner}}}
	namespaceTwoMembers := &models.Namespace{Name: "group1", Owner: "ID1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484714", Members: []models.Member{{ID: user1.ID, Role: guard.RoleOwner}, {ID: user2.ID, Role: guard.RoleObserver}}}

	type Expected struct {
		namespace *models.Namespace
		err       error
	}

	cases := []struct {
		Name          string
		TenantID      string
		Username      string
		Role          string
		ID            string
		RequiredMocks func()
		Expected      Expected
	}{
		{
			Name:     "AddNamespaceUser fails when MemberID is not valid",
			Username: "invalid_username",
			Role:     guard.RoleObserver,
			ID:       user1.ID,
			TenantID: namespace.TenantID,
			RequiredMocks: func() {
			},
			Expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceMemberInvalid(validator.ErrInvalidFields),
			},
		},
		{
			Name:     "AddNamespaceUser fails when Role is not valid",
			Username: user2.Username,
			Role:     "invalid",
			ID:       user1.ID,
			TenantID: namespace.TenantID,
			RequiredMocks: func() {
			},
			Expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceMemberInvalid(validator.ErrInvalidFields),
			},
		},
		{
			Name:     "AddNamespaceUser fails when the namespace was not found",
			Username: user2.Username,
			Role:     guard.RoleObserver,
			ID:       user1.ID,
			TenantID: "tenantIDNotFound",
			RequiredMocks: func() {
				mock.On("NamespaceGet", ctx, "tenantIDNotFound").Return(nil, ErrNamespaceNotFound).Once()
			},
			Expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceNotFound("tenantIDNotFound", ErrNamespaceNotFound),
			},
		},
		{
			Name:     "AddNamespaceUser fails when the active member was not found",
			Username: user1.Username,
			Role:     guard.RoleObserver,
			ID:       "userIDNotFound",
			TenantID: namespace.TenantID,
			RequiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()

				mock.On("UserGetByID", ctx, "userIDNotFound", false).Return(nil, 0, ErrUserNotFound).Once()
			},
			Expected: Expected{
				namespace: nil,
				err:       NewErrUserNotFound("userIDNotFound", ErrUserNotFound),
			},
		},
		{
			Name:     "AddNamespaceUser fails when the active member is not on the namespace",
			Username: user1.Username,
			Role:     guard.RoleObserver,
			ID:       user2.ID,
			TenantID: namespace.TenantID,
			RequiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()

				mock.On("UserGetByID", ctx, user2.ID, false).Return(user2, 0, nil).Once()
			},
			Expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceMemberNotFound(user2.ID, nil),
			},
		},
		{
			Name:     "addNamespaceUser fails when passive member was not found",
			Username: "userNamespaceNotFound",
			Role:     guard.RoleObserver,
			ID:       user1.ID,
			TenantID: namespace.TenantID,
			RequiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()

				mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
				mock.On("UserGetByUsername", ctx, "userNamespaceNotFound").Return(nil, ErrBadRequest).Once()
			},
			Expected: Expected{
				namespace: nil,
				err:       NewErrUserNotFound("userNamespaceNotFound", ErrBadRequest),
			},
		},
		{
			Name:     "AddNamespaceUser fails when the passive member is on the namespace",
			Username: user2.Username,
			Role:     guard.RoleObserver,
			ID:       user1.ID,
			TenantID: namespaceTwoMembers.TenantID,
			RequiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespaceTwoMembers.TenantID).Return(namespaceTwoMembers, nil).Once()

				mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
				mock.On("UserGetByUsername", ctx, user2.Username).Return(user2, nil).Once()
			},
			Expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceMemberDuplicated(user2.ID, nil),
			},
		},
		{
			Name:     "AddNamespaceUser succeeds",
			Username: user2.Username,
			Role:     guard.RoleObserver,
			ID:       user1.ID,
			TenantID: namespace.TenantID,
			RequiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()

				mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
				mock.On("UserGetByUsername", ctx, user2.Username).Return(user2, nil).Once()

				mock.On("NamespaceAddMember", ctx, namespace.TenantID, user2.ID, guard.RoleObserver).Return(namespaceTwoMembers, nil).Once()
			},
			Expected: Expected{
				namespace: namespaceTwoMembers,
				err:       nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.RequiredMocks()
			ns, err := s.AddNamespaceUser(ctx, tc.Username, tc.Role, tc.TenantID, tc.ID)
			assert.Equal(t, tc.Expected, Expected{ns, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestRemoveNamespaceUser(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
	ctx := context.TODO()
	user := &models.User{UserData: models.UserData{Name: "user1", Username: "username1"}, ID: "hash1"}
	user2 := &models.User{UserData: models.UserData{Name: "user2", Username: "username2"}, ID: "hash2"}
	user3 := &models.User{UserData: models.UserData{Name: "user3", Username: "username3"}, ID: "hash3"}
	namespace := &models.Namespace{Name: "group1", Owner: user.ID, TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Members: []models.Member{{ID: user.ID, Role: guard.RoleOwner}}}
	namespaceTwoMembers := &models.Namespace{Name: "group2", Owner: user.ID, TenantID: "a736a52b-5777-4f92-b0b8-e359bf484714", Members: []models.Member{{ID: user.ID, Role: guard.RoleOwner}, {ID: user2.ID, Role: guard.RoleObserver}}}
	namespaceThreeMembers := &models.Namespace{Name: "group2", Owner: user.ID, TenantID: "a736a52b-5777-4f92-b0b8-e359bf484714", Members: []models.Member{{ID: user.ID, Role: guard.RoleOwner}, {ID: user2.ID, Role: guard.RoleAdministrator}, {ID: user3.ID, Role: guard.RoleAdministrator}}}
	Err := errors.New("error")

	type Expected struct {
		namespace *models.Namespace
		err       error
	}

	cases := []struct {
		Name          string
		RequiredMocks func()
		TenantID      string
		MemberID      string
		UserID        string
		Expected      Expected
	}{
		{
			Name: "RemoveNamespaceUser fails when namespace was not found",
			RequiredMocks: func() {
				mock.On("NamespaceGet", ctx, "tenantIDNotFound").Return(nil, Err).Once()
			},
			TenantID: "tenantIDNotFound",
			MemberID: user.ID,
			UserID:   namespace.Owner,
			Expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceNotFound("tenantIDNotFound", Err),
			},
		},
		{
			Name: "RemoveNamespaceUser fails when active user was not found",
			RequiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, "invalidUserID", false).Return(nil, 0, ErrUserNotFound).Once()
			},
			TenantID: namespace.TenantID,
			MemberID: user2.ID,
			UserID:   "invalidUserID",
			Expected: Expected{
				namespace: nil,
				err:       NewErrUserNotFound("invalidUserID", ErrUserNotFound),
			},
		},
		{
			Name: "RemoveNamespaceUser fails when passive user was not found",
			RequiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, namespace.Owner, false).Return(user, 0, nil).Once()
				mock.On("UserGetByID", ctx, "invalidPassiveMemberID", false).Return(nil, 0, ErrUserNotFound).Once()
			},
			TenantID: namespace.TenantID,
			MemberID: "invalidPassiveMemberID",
			UserID:   namespace.Owner,
			Expected: Expected{
				namespace: nil,
				err:       NewErrUserNotFound("invalidPassiveMemberID", ErrUserNotFound),
			},
		},
		{
			Name: "RemoveNamespaceUser fails when user is not a namespace's member",
			RequiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, namespace.Owner, false).Return(user, 0, nil).Once()
				mock.On("UserGetByID", ctx, user2.ID, false).Return(user2, 0, nil).Once()
			},
			TenantID: namespace.TenantID,
			MemberID: user2.ID,
			UserID:   user.ID,
			Expected: Expected{
				namespace: nil,
				err:       NewErrNamespaceMemberNotFound(user2.ID, nil),
			},
		},
		{
			Name: "RemoveNamespaceUser fails when user can not act over the role",
			RequiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespaceThreeMembers.TenantID).Return(namespaceThreeMembers, nil).Once()
				mock.On("UserGetByID", ctx, user2.ID, false).Return(user2, 0, nil).Once()
				mock.On("UserGetByID", ctx, user3.ID, false).Return(user3, 0, nil).Once()
			},
			TenantID: namespaceThreeMembers.TenantID,
			MemberID: user3.ID,
			UserID:   user2.ID,
			Expected: Expected{
				namespace: nil,
				err:       guard.ErrForbidden,
			},
		},
		{
			Name: "RemoveNamespaceUser when NamespaceRemoveMember store's function fails",
			RequiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespaceTwoMembers.TenantID).Return(namespaceTwoMembers, nil).Once()
				mock.On("UserGetByID", ctx, namespace.Owner, false).Return(user, 0, nil).Once()
				mock.On("UserGetByID", ctx, user2.ID, false).Return(user2, 0, nil).Once()

				mock.On("NamespaceRemoveMember", ctx, namespaceTwoMembers.TenantID, user2.ID).Return(nil, Err).Once()
			},
			TenantID: namespaceTwoMembers.TenantID,
			MemberID: user2.ID,
			UserID:   user.ID,
			Expected: Expected{
				namespace: nil,
				err:       Err,
			},
		},
		{
			Name: "RemoveNamespaceUser succeeds",
			RequiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespaceTwoMembers.TenantID).Return(namespaceTwoMembers, nil).Once()
				mock.On("UserGetByID", ctx, namespace.Owner, false).Return(user, 0, nil).Once()
				mock.On("UserGetByID", ctx, user2.ID, false).Return(user2, 0, nil).Once()

				mock.On("NamespaceRemoveMember", ctx, namespaceTwoMembers.TenantID, user2.ID).Return(namespace, nil).Once()
			},
			TenantID: namespaceTwoMembers.TenantID,
			MemberID: user2.ID,
			UserID:   user.ID,
			Expected: Expected{
				namespace: namespace,
				err:       nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.RequiredMocks()
			ns, err := s.RemoveNamespaceUser(ctx, tc.TenantID, tc.MemberID, tc.UserID)
			assert.Equal(t, tc.Expected, Expected{ns, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestEditNamespaceUser(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
	ctx := context.TODO()
	Err := errors.New("error")

	activeMember := &models.User{UserData: models.UserData{Name: "activeMemberName", Username: "activeMemberUsername"}, ID: "activeMemberID"}
	passiveMember := &models.User{UserData: models.UserData{Name: "passiveMemberName", Username: "passiveMemberUsername"}, ID: "passiveMemberID"}
	namespaceActiveOwner := &models.Namespace{Name: "group1", Owner: activeMember.ID, TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Members: []models.Member{{ID: activeMember.ID, Role: guard.RoleOwner}}}
	namespacePassiveObserver := &models.Namespace{Name: "group1", Owner: activeMember.ID, TenantID: "a736a52b-5777-4f92-b0b8-e359bf484714", Members: []models.Member{{ID: "memberID", Role: guard.RoleOwner}, {ID: passiveMember.ID, Role: guard.RoleObserver}}}
	namespaceActivePassiveSame := &models.Namespace{Name: "group1", Owner: activeMember.ID, TenantID: "a736a52b-5777-4f92-b0b8-e359bf484715", Members: []models.Member{{ID: "ownerID", Role: guard.RoleOwner}, {ID: activeMember.ID, Role: guard.RoleAdministrator}, {ID: passiveMember.ID, Role: guard.RoleAdministrator}}}
	namespaceActiveHasNoPermission := &models.Namespace{Name: "group1", Owner: activeMember.ID, TenantID: "a736a52b-5777-4f92-b0b8-e359bf484716", Members: []models.Member{{ID: "ownerID", Role: guard.RoleOwner}, {ID: activeMember.ID, Role: guard.RoleOperator}, {ID: passiveMember.ID, Role: guard.RoleObserver}}}
	namespaceActivePassive := &models.Namespace{Name: "group1", Owner: activeMember.ID, TenantID: "a736a52b-5777-4f92-b0b8-e359bf484717", Members: []models.Member{{ID: "ownerID", Role: guard.RoleOwner}, {ID: activeMember.ID, Role: guard.RoleAdministrator}, {ID: passiveMember.ID, Role: guard.RoleObserver}}}

	cases := []struct {
		Name          string
		TenantID      string
		UserID        string
		MemberID      string
		MemberNewRole string
		RequiredMocks func()
		Expected      error
	}{
		{
			Name:          "EditNamespaceUser fails when namespace was not found",
			TenantID:      "tenantIDNotFound",
			UserID:        activeMember.ID,
			MemberID:      passiveMember.ID,
			MemberNewRole: guard.RoleObserver,
			RequiredMocks: func() {
				mock.On("NamespaceGet", ctx, "tenantIDNotFound").Return(nil, Err).Once()
			},
			Expected: NewErrNamespaceNotFound("tenantIDNotFound", Err),
		},
		{
			Name:          "EditNamespaceUser fails when active member was not found",
			TenantID:      namespaceActivePassive.TenantID,
			UserID:        "invalidMemberActiveID",
			MemberID:      passiveMember.ID,
			MemberNewRole: guard.RoleObserver,
			RequiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespaceActivePassive.TenantID).Return(namespaceActivePassive, nil).Once()

				mock.On("UserGetByID", ctx, "invalidMemberActiveID", false).Return(nil, 0, Err).Once()
			},
			Expected: NewErrUserNotFound("invalidMemberActiveID", Err),
		},
		{
			Name:          "EditNamespaceUser fails when passive member was not found",
			TenantID:      namespaceActiveOwner.TenantID,
			UserID:        activeMember.ID,
			MemberID:      "invalidMemberPassiveID",
			MemberNewRole: guard.RoleObserver,
			RequiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespaceActiveOwner.TenantID).Return(namespaceActiveOwner, nil).Once()

				mock.On("UserGetByID", ctx, activeMember.ID, false).Return(activeMember, 0, nil).Once()
				mock.On("UserGetByID", ctx, "invalidMemberPassiveID", false).Return(nil, 0, Err).Once()
			},
			Expected: NewErrUserNotFound("invalidMemberPassiveID", Err),
		},
		{
			Name:          "EditNamespaceUser fails when could not find passive member inside namespace",
			TenantID:      namespaceActiveOwner.TenantID,
			UserID:        activeMember.ID,
			MemberID:      passiveMember.ID,
			MemberNewRole: guard.RoleObserver,
			RequiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespaceActiveOwner.TenantID).Return(namespaceActiveOwner, nil).Once()

				mock.On("UserGetByID", ctx, activeMember.ID, false).Return(activeMember, 0, nil).Once()
				mock.On("UserGetByID", ctx, passiveMember.ID, false).Return(passiveMember, 0, nil).Once()
			},
			Expected: NewErrNamespaceMemberNotFound(passiveMember.ID, nil),
		},
		{
			Name:          "EditNamespaceUser fails when could not find active member inside namespace",
			TenantID:      namespacePassiveObserver.TenantID,
			UserID:        activeMember.ID,
			MemberID:      passiveMember.ID,
			MemberNewRole: guard.RoleOperator,
			RequiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespacePassiveObserver.TenantID).Return(namespacePassiveObserver, nil).Once()

				mock.On("UserGetByID", ctx, activeMember.ID, false).Return(activeMember, 0, nil).Once()
				mock.On("UserGetByID", ctx, passiveMember.ID, false).Return(passiveMember, 0, nil).Once()
			},
			Expected: NewErrNamespaceMemberNotFound(activeMember.ID, nil),
		},
		{
			Name:          "EditNamespaceUser fails when active and passive roles are the same",
			TenantID:      namespaceActivePassiveSame.TenantID,
			UserID:        activeMember.ID,
			MemberID:      passiveMember.ID,
			MemberNewRole: guard.RoleOperator,
			RequiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespaceActivePassiveSame.TenantID).Return(namespaceActivePassiveSame, nil).Once()

				mock.On("UserGetByID", ctx, passiveMember.ID, false).Return(passiveMember, 0, nil).Once()
				mock.On("UserGetByID", ctx, activeMember.ID, false).Return(activeMember, 0, nil).Once()
			},
			Expected: guard.ErrForbidden,
		},
		{
			Name:          "EditNamespaceUser fails when user can not act over the role",
			TenantID:      namespaceActiveHasNoPermission.TenantID,
			UserID:        activeMember.ID,
			MemberID:      passiveMember.ID,
			MemberNewRole: guard.RoleAdministrator,
			RequiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespaceActiveHasNoPermission.TenantID).Return(namespaceActiveHasNoPermission, nil).Once()

				mock.On("UserGetByID", ctx, passiveMember.ID, false).Return(passiveMember, 0, nil).Once()
				mock.On("UserGetByID", ctx, activeMember.ID, false).Return(activeMember, 0, nil).Once()

				// mock.On("NamespaceGet", ctx, namespaceActiveHasNoPermission.TenantID).Return(namespaceActiveHasNoPermission, nil).Once()
			},
			Expected: guard.ErrForbidden,
		},
		{
			Name:          "EditNamespaceUser fails when user store function fails",
			TenantID:      namespaceActivePassive.TenantID,
			UserID:        activeMember.ID,
			MemberID:      passiveMember.ID,
			MemberNewRole: guard.RoleOperator,
			RequiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespaceActivePassive.TenantID).Return(namespaceActivePassive, nil).Once()

				mock.On("UserGetByID", ctx, passiveMember.ID, false).Return(passiveMember, 0, nil).Once()
				mock.On("UserGetByID", ctx, activeMember.ID, false).Return(activeMember, 0, nil).Once()

				mock.On("NamespaceEditMember", ctx, namespaceActivePassive.TenantID, passiveMember.ID, guard.RoleOperator).Return(Err).Once()
			},
			Expected: Err,
		},
		{
			Name:          "EditNamespaceUser Success",
			TenantID:      namespaceActivePassive.TenantID,
			UserID:        activeMember.ID,
			MemberID:      passiveMember.ID,
			MemberNewRole: guard.RoleOperator,
			RequiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespaceActivePassive.TenantID).Return(namespaceActivePassive, nil).Once()

				mock.On("UserGetByID", ctx, passiveMember.ID, false).Return(passiveMember, 0, nil).Once()
				mock.On("UserGetByID", ctx, activeMember.ID, false).Return(activeMember, 0, nil).Once()

				mock.On("NamespaceEditMember", ctx, namespaceActivePassive.TenantID, passiveMember.ID, guard.RoleOperator).Return(nil).Once()
			},
			Expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.RequiredMocks()
			err := s.EditNamespaceUser(ctx, tc.TenantID, tc.UserID, tc.MemberID, tc.MemberNewRole)
			assert.Equal(t, tc.Expected, err)
		})
	}
	mock.AssertExpectations(t)
}

func TestGetSessionRecord(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()

	Err := errors.New("error")

	type Expected struct {
		status bool
		err    error
	}

	namespace := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Settings: &models.NamespaceSettings{SessionRecord: true}}

	cases := []struct {
		name          string
		requiredMocks func()
		tenantID      string
		expected      Expected
	}{
		{
			name: "GetSessionRecord fails when the namespace document is not found",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, store.ErrNoDocuments).Once()
			},
			tenantID: namespace.TenantID,
			expected: Expected{false, NewErrNamespaceNotFound(namespace.TenantID, store.ErrNoDocuments)},
		},
		{
			name: "GetSessionRecord fails when store namespace get fails",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(nil, Err).Once()
			},
			tenantID: namespace.TenantID,
			expected: Expected{false, NewErrNamespaceNotFound(namespace.TenantID, Err)},
		},
		{
			name: "GetSessionRecord fails when store namespace get session record fails",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("NamespaceGetSessionRecord", ctx, namespace.TenantID).Return(false, Err).Once()
			},
			tenantID: namespace.TenantID,
			expected: Expected{false, Err},
		},
		{
			name: "GetSessionRecord succeeds",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("NamespaceGetSessionRecord", ctx, namespace.TenantID).Return(true, nil).Once()
			},
			tenantID: namespace.TenantID,
			expected: Expected{true, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			returnedUserSecurity, err := s.GetSessionRecord(ctx, namespace.TenantID)
			assert.Equal(t, tc.expected, Expected{returnedUserSecurity, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestEditSessionRecord(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()

	user := &models.User{UserData: models.UserData{Name: "user1", Username: "username1"}, ID: "hash1"}
	user2 := &models.User{UserData: models.UserData{Name: "user2", Username: "username2"}, ID: "hash2"}
	namespace := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "xxxx", Settings: &models.NamespaceSettings{SessionRecord: true}, Members: []models.Member{
		{
			ID:   user.ID,
			Role: guard.RoleOwner,
		},
		{
			ID:   user2.ID,
			Role: guard.RoleObserver,
		},
	}}

	Err := errors.New("error")

	cases := []struct {
		name              string
		requiredMocks     func()
		sessionRecord     bool
		ownerID, tenantID string
		expected          error
	}{
		{
			name:    "EditSessionRecord fails when namespace set session record fails",
			ownerID: namespace.Owner,
			requiredMocks: func() {
				status := true
				mock.On("NamespaceSetSessionRecord", ctx, status, namespace.TenantID).Return(Err).Once()
			},
			tenantID:      namespace.TenantID,
			sessionRecord: true,
			expected:      Err,
		},
		{
			name:    "EditSessionRecord succeeds",
			ownerID: namespace.Owner,
			requiredMocks: func() {
				status := true
				mock.On("NamespaceSetSessionRecord", ctx, status, namespace.TenantID).Return(nil).Once()
			},
			tenantID:      namespace.TenantID,
			sessionRecord: true,
			expected:      nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			err := s.EditSessionRecordStatus(ctx, tc.sessionRecord, tc.tenantID)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}
