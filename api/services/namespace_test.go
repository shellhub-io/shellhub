package services

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"

	utils "github.com/shellhub-io/shellhub/api/pkg/namespace"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/envs"
	env_mocks "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	uuid_mocks "github.com/shellhub-io/shellhub/pkg/uuid/mocks"
	"github.com/stretchr/testify/assert"
)

func TestIsNamespaceOwner(t *testing.T) {
	mock := &mocks.Store{}

	ctx := context.TODO()

	user := &models.User{Name: "user1", Username: "hash1", ID: "hash1"}
	user2 := &models.User{Name: "user2", Username: "hash2", ID: "hash2"}

	Err := errors.New("error")

	namespace := &models.Namespace{
		Name:     strings.ToLower("namespace"),
		Owner:    user.ID,
		Members:  []interface{}{user.ID, user2.ID},
		Settings: &models.NamespaceSettings{SessionRecord: true},
		TenantID: "xxxxx",
	}

	cases := []struct {
		name              string
		tenantID, ownerID string
		requiredMocks     func()
		expected          error
	}{
		{
			name:     "IsNamespaceOwner fails when the user document is not found",
			tenantID: namespace.TenantID,
			ownerID:  namespace.Owner,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, namespace.Owner, false).Return(nil, 0, store.ErrNoDocuments).Once()
			},
			expected: ErrUnauthorized,
		},
		{
			name:     "IsNamespaceOwner fails store user get by id fails",
			tenantID: namespace.TenantID,
			ownerID:  namespace.Owner,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, namespace.Owner, false).Return(nil, 0, Err).Once()
			},
			expected: Err,
		},
		{
			name:     "IsNamespaceOwner fails when no namespace document is found",
			tenantID: namespace.TenantID,
			ownerID:  namespace.Owner,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, namespace.Owner, false).Return(user, 0, nil).Once()
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, store.ErrNoDocuments).Once()
			},
			expected: ErrNamespaceNotFound,
		},
		{
			name:     "IsNamespaceOwner fails when store namespace get fails",
			tenantID: namespace.TenantID,
			ownerID:  namespace.Owner,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, namespace.Owner, false).Return(user, 0, nil).Once()
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, Err).Once()
			},
			expected: Err,
		},
		{
			name:     "IsNamespaceOwner fails when the user is not the owner",
			tenantID: namespace.TenantID,
			ownerID:  user2.ID,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, user2.ID, false).Return(user2, 0, nil).Once()
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
			},
			expected: ErrUnauthorized,
		},
		{
			name:     "IsNamespaceOwner succeeds",
			tenantID: namespace.TenantID,
			ownerID:  namespace.Owner,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			err := utils.IsNamespaceOwner(ctx, store.Store(mock), tc.tenantID, tc.ownerID)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestListNamespaces(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), nil, nil)

	ctx := context.TODO()
	Err := errors.New("error")

	namespaces := []models.Namespace{
		{Name: "group1", Owner: "ID1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713"},
		{Name: "group2", Owner: "ID2", TenantID: "a736a52b-5777-4f92-b0b8-e359bf48471i4"},
	}

	type Expected struct {
		namespaces []models.Namespace
		count      int
		err        error
	}

	query := paginator.Query{Page: 1, PerPage: 10}

	cases := []struct {
		name          string
		pagination    paginator.Query
		filter        string
		ctx           context.Context
		requiredMocks func()
		expected      Expected
	}{
		{
			name:       "ListNamespaces fails",
			pagination: query,
			filter:     "",
			ctx:        ctx,
			requiredMocks: func() {
				mock.On("NamespaceList", ctx, query, []models.Filter(nil), false).Return(nil, 0, Err).Once()
			},
			expected: Expected{
				namespaces: nil,
				count:      0,
				err:        Err,
			},
		},
		{
			name:       "ListNamespaces succeeds",
			pagination: query,
			filter:     "",
			ctx:        ctx,
			requiredMocks: func() {
				mock.On("NamespaceList", ctx, query, []models.Filter(nil), false).Return(namespaces, len(namespaces), nil).Once()
			},
			expected: Expected{
				namespaces: namespaces,
				count:      len(namespaces),
				err:        nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			nss, count, err := s.ListNamespaces(tc.ctx, tc.pagination, "", false)
			assert.Equal(t, tc.expected, Expected{nss, count, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestGetNamespace(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), nil, nil)

	ctx := context.TODO()

	namespace := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713"}

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
			name:     "GetNamespace fails",
			ctx:      ctx,
			tenantID: namespace.TenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(nil, Err).Once()
			},
			expected: Expected{
				namespace: nil,
				err:       Err,
			},
		},
		{
			name:     "GetNamespace succeeds",
			ctx:      ctx,
			tenantID: namespace.TenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
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

func TestListMembers(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), nil, nil)

	ctx := context.TODO()
	envMock := &env_mocks.Backend{}
	envs.DefaultBackend = envMock

	type Expected struct {
		ns  []models.Member
		err error
	}

	Err := errors.New("error")

	user := &models.User{Name: "user1", Username: "hash1", ID: "hash1"}
	user2 := &models.User{Name: "user2", Username: "hash2", ID: "hash2"}

	namespace := &models.Namespace{
		Name:     strings.ToLower("namespace"),
		Owner:    user.ID,
		Members:  []interface{}{user.ID, user2.ID},
		Settings: &models.NamespaceSettings{SessionRecord: true},
		TenantID: "xxxxx",
	}

	cases := []struct {
		name          string
		tenantID      string
		requiredMocks func()
		expected      Expected
	}{
		{
			name:     "ListMembers fails when no namespace document is found",
			tenantID: namespace.TenantID,
			requiredMocks: func() {
				tenant := namespace.TenantID
				mock.On("NamespaceGet", ctx, tenant).Return(namespace, store.ErrNoDocuments).Once()
			},
			expected: Expected{
				nil,
				ErrNamespaceNotFound,
			},
		},
		{
			name:     "ListMembers fails when no store namespace get fails",
			tenantID: namespace.TenantID,
			requiredMocks: func() {
				tenant := namespace.TenantID
				mock.On("NamespaceGet", ctx, tenant).Return(namespace, Err).Once()
			},
			expected: Expected{
				nil,
				Err,
			},
		},
		{
			name:     "ListMembers fails when no user documents is found",
			tenantID: namespace.TenantID,
			requiredMocks: func() {
				tenant := namespace.TenantID
				mock.On("NamespaceGet", ctx, tenant).Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
				mock.On("UserGetByID", ctx, user2.ID, false).Return(nil, 0, store.ErrNoDocuments).Once()
			},
			expected: Expected{
				nil,
				ErrUserNotFound,
			},
		},
		{
			name:     "ListMembers fails when store get user by id fails",
			tenantID: namespace.TenantID,
			requiredMocks: func() {
				tenant := namespace.TenantID
				mock.On("NamespaceGet", ctx, tenant).Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
				mock.On("UserGetByID", ctx, user2.ID, false).Return(user2, 0, Err).Once()
			},
			expected: Expected{
				nil,
				Err,
			},
		},
		{
			name:     "ListMembers succeeds",
			tenantID: namespace.TenantID,
			requiredMocks: func() {
				tenant := namespace.TenantID
				mock.On("NamespaceGet", ctx, tenant).Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
				mock.On("UserGetByID", ctx, user2.ID, false).Return(user2, 0, nil).Once()
			},
			expected: Expected{
				[]models.Member{{ID: user.ID, Name: user.Username}, {ID: user2.ID, Name: user2.Username}},
				nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			members, err := s.ListMembers(ctx, tc.tenantID)
			assert.Equal(t, tc.expected, Expected{members, err})
		})
	}
}

func TestCreateNamespace(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), nil, nil)

	ctx := context.TODO()
	envMock := &env_mocks.Backend{}
	envs.DefaultBackend = envMock
	uuidMock := &uuid_mocks.Uuid{}
	uuid.DefaultBackend = uuidMock

	type Expected struct {
		ns  *models.Namespace
		err error
	}

	Err := errors.New("error")

	user := &models.User{Name: "user1", Username: "hash1", ID: "hash1"}

	namespace := &models.Namespace{
		Name:     strings.ToLower("namespace"),
		Owner:    user.ID,
		Members:  []interface{}{user.ID},
		Settings: &models.NamespaceSettings{SessionRecord: true},
		TenantID: "xxxxx",
	}

	cases := []struct {
		name          string
		requiredMocks func()
		ownerID       string
		expected      Expected
		namespace     *models.Namespace
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
				ErrUnauthorized,
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
				Err,
			},
		},
		{
			name:    "CreateNamespace fails when a namespace field is invalid",
			ownerID: user.ID,
			namespace: &models.Namespace{
				Name: "name.with.dot",
			},
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
			},
			expected: Expected{
				nil,
				ErrInvalidFormat,
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
				mock.On("NamespaceGetByName", ctx, namespace.Name).Return(namespace, nil).Once()
			},
			expected: Expected{
				nil,
				ErrConflictName,
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
				mock.On("NamespaceGetByName", ctx, namespace.Name).Return(namespace, Err).Once()
			},
			expected: Expected{
				nil,
				Err,
			},
		},
		{
			name:      "CreateNamespace fails when store namespace create fails",
			ownerID:   user.ID,
			namespace: namespace,
			requiredMocks: func() {
				var isCloud bool
				notCloudNamespace := &models.Namespace{
					Name:       strings.ToLower("namespace"),
					Owner:      user.ID,
					Members:    []interface{}{user.ID},
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
				nil, Err,
			},
		},
		{
			name:    "CreateNamespace generates namespace with random tenant",
			ownerID: user.ID,
			namespace: &models.Namespace{
				Name:       strings.ToLower("namespace"),
				Owner:      user.ID,
				Members:    []interface{}{user.ID},
				Settings:   &models.NamespaceSettings{SessionRecord: true},
				TenantID:   "",
				MaxDevices: -1,
			},
			requiredMocks: func() {
				var isCloud bool
				notCloudNamespace := &models.Namespace{
					Name:       strings.ToLower("namespace"),
					Owner:      user.ID,
					Members:    []interface{}{user.ID},
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
					Name:       strings.ToLower("namespace"),
					Owner:      user.ID,
					Members:    []interface{}{user.ID},
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
					Name:       strings.ToLower("namespace"),
					Owner:      user.ID,
					Members:    []interface{}{user.ID},
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
					Name:       strings.ToLower(namespace.Name),
					Owner:      user.ID,
					Members:    []interface{}{user.ID},
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
					Name:       strings.ToLower("namespace"),
					Owner:      user.ID,
					Members:    []interface{}{user.ID},
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
					Name:       strings.ToLower(namespace.Name),
					Owner:      user.ID,
					Members:    []interface{}{user.ID},
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
	s := NewService(store.Store(mock), nil, nil)

	ctx := context.TODO()

	user := &models.User{Name: "user1", Username: "hash1", ID: "hash1"}
	user2 := &models.User{Name: "user2", Username: "hash2", ID: "hash2"}
	Err := errors.New("error")

	namespace := &models.Namespace{
		Name:     "oldname",
		Owner:    "hash1",
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
			name:     "EditNamespace fails when the user is not the owner",
			tenantID: namespace.TenantID,
			ownerID:  user2.ID,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, user2.ID, false).Return(user2, 0, nil).Once()
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
			},
			expected: Expected{
				nil,
				ErrUnauthorized,
			},
		},
		{
			name: "EditNamespace fails when the name is invalid",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Twice()
				mock.On("UserGetByID", ctx, namespace.Owner, false).Return(user, 0, nil).Once()
			},
			tenantID:      "xxxxx",
			namespaceName: "name.with.dot",
			ownerID:       user.ID,
			expected: Expected{
				nil,
				ErrInvalidFormat,
			},
		},
		{
			name: "EditNamespace fails when the name is the same",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Twice()
				mock.On("UserGetByID", ctx, namespace.Owner, false).Return(user, 0, nil).Once()
			},
			tenantID:      "xxxxx",
			namespaceName: namespace.Name,
			ownerID:       user.ID,
			expected: Expected{
				nil,
				ErrUnauthorized,
			},
		},
		{
			name: "EditNamespace fails when the store namespace rename fails",
			requiredMocks: func() {
				newName := "newname"
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Twice()
				mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
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
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Twice()
				mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
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
			returnedNamespace, err := s.EditNamespace(ctx, tc.tenantID, tc.namespaceName, tc.ownerID)
			assert.Equal(t, tc.expected, Expected{returnedNamespace, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestDeleteNamespace(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), nil, nil)

	Err := errors.New("error")

	ctx := context.TODO()

	namespace := &models.Namespace{Name: "oldname", Owner: "ID1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713"}
	user1 := &models.User{Name: "user1", ID: "ID1", Username: "user1", Email: "user1@email.com"}
	user2 := &models.User{Name: "user2", ID: "ID2", Username: "user2", Email: "user2@email.com"}

	cases := []struct {
		name          string
		tenantID      string
		ownerID       string
		requiredMocks func()
		expected      error
	}{
		{
			name:     "DeleteNamespace fails when the user is not the owner",
			tenantID: namespace.TenantID,
			ownerID:  user2.ID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, user2.ID, false).Return(user2, 0, nil).Once()
			},
			expected: ErrUnauthorized,
		},
		{
			name:     "DeleteNamespace fails when store delete fails",
			tenantID: namespace.TenantID,
			ownerID:  user1.ID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
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
				mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
				mock.On("NamespaceDelete", ctx, namespace.TenantID).Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			err := s.DeleteNamespace(ctx, tc.tenantID, tc.ownerID)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestAddNamespaceUser(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), nil, nil)
	ctx := context.TODO()

	namespace := &models.Namespace{Name: "group1", Owner: "ID1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713"}
	user1 := &models.User{Name: "user1", ID: "ID1", Username: "user1", Email: "user1@email.com"}
	user2 := &models.User{Name: "user2", ID: "ID2", Username: "user2", Email: "user2@email.com"}

	Err := errors.New("error")

	type Expected struct {
		namespace *models.Namespace
		err       error
	}

	cases := []struct {
		name          string
		tenantID      string
		username      string
		ownerID       string
		requiredMocks func()
		expected      Expected
	}{
		{
			name:     "AddNamespaceUser fails when user is not the owner",
			ownerID:  user2.ID,
			tenantID: namespace.TenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, user2.ID, false).Return(user2, 0, nil).Once()
			},
			expected: Expected{
				namespace: nil,
				err:       ErrUnauthorized,
			},
		},
		{
			name:     "AddNamespaceUser fails when user get by username has no documents",
			username: user2.Username,
			ownerID:  user1.ID,
			tenantID: namespace.TenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, namespace.Owner, false).Return(user1, 0, nil).Once()
				mock.On("UserGetByUsername", ctx, user2.Username).Return(nil, store.ErrNoDocuments).Once()
			},
			expected: Expected{
				namespace: nil,
				err:       ErrUserNotFound,
			},
		},
		{
			name:     "AddNamespaceUser fails when user get by username fails",
			username: user2.Username,
			ownerID:  user1.ID,
			tenantID: namespace.TenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, namespace.Owner, false).Return(user1, 0, nil).Once()
				mock.On("UserGetByUsername", ctx, user2.Username).Return(nil, Err).Once()
			},
			expected: Expected{
				namespace: nil,
				err:       Err,
			},
		},
		{
			name:     "AddNamespaceUser fails when store namespace add member fails",
			username: user2.Username,
			ownerID:  user1.ID,
			tenantID: namespace.TenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, namespace.Owner, false).Return(user1, 0, nil).Once()
				mock.On("UserGetByUsername", ctx, user2.Username).Return(user2, nil).Once()
				mock.On("NamespaceAddMember", ctx, namespace.TenantID, user2.ID).Return(nil, Err).Once()
			},
			expected: Expected{
				namespace: nil,
				err:       Err,
			},
		},
		{
			name:     "AddNamespaceUser succeeds",
			username: user2.Username,
			ownerID:  user1.ID,
			tenantID: namespace.TenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, namespace.Owner, false).Return(user1, 0, nil).Once()
				mock.On("UserGetByUsername", ctx, user2.Username).Return(user2, nil).Once()
				mock.On("NamespaceAddMember", ctx, namespace.TenantID, user2.ID).Return(namespace, nil).Once()
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
			ns, err := s.AddNamespaceUser(ctx, tc.tenantID, tc.username, tc.ownerID)
			assert.Equal(t, tc.expected, Expected{ns, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestRemoveNamespaceUser(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), nil, nil)
	ctx := context.TODO()
	user := &models.User{Name: "user1", Username: "username1", ID: "hash1"}
	user2 := &models.User{Name: "user2", Username: "username2", ID: "hash2"}
	namespace := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Members: []interface{}{"hash1", "hash2"}}
	Err := errors.New("error")

	type Expected struct {
		namespace *models.Namespace
		err       error
	}

	cases := []struct {
		name          string
		requiredMocks func()
		tenantID      string
		username      string
		ownerID       string
		expected      Expected
	}{
		{
			name:     "RemoveNamespaceUser fails when user is not the owner",
			ownerID:  user2.ID,
			tenantID: namespace.TenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, user2.ID, false).Return(user2, 0, nil).Once()
			},
			expected: Expected{
				namespace: nil,
				err:       ErrUnauthorized,
			},
		},
		{
			name: "RemoveNamespaceUser fails when store user get by username has no documents",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, namespace.Owner, false).Return(user, 0, nil).Once()
				mock.On("UserGetByUsername", ctx, user2.Username).Return(nil, store.ErrNoDocuments).Once()
			},
			tenantID: namespace.TenantID,
			username: user2.Username,
			ownerID:  namespace.Owner,
			expected: Expected{
				namespace: nil,
				err:       ErrUserNotFound,
			},
		},
		{
			name: "RemoveNamespaceUser fails when store user get by username fails",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, namespace.Owner, false).Return(user, 0, nil).Once()
				mock.On("UserGetByUsername", ctx, user2.Username).Return(nil, Err).Once()
			},
			tenantID: namespace.TenantID,
			username: user2.Username,
			ownerID:  namespace.Owner,
			expected: Expected{
				namespace: nil,
				err:       Err,
			},
		},
		{
			name: "RemoveNamespaceUser fails when store namespace remove member fails",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, namespace.Owner, false).Return(user, 0, nil).Once()
				mock.On("UserGetByUsername", ctx, user2.Username).Return(user2, nil).Once()
				mock.On("NamespaceRemoveMember", ctx, namespace.TenantID, user2.ID).Return(nil, Err).Once()
			},
			tenantID: namespace.TenantID,
			username: user2.Username,
			ownerID:  namespace.Owner,
			expected: Expected{
				namespace: nil,
				err:       Err,
			},
		},
		{
			name: "RemoveNamespaceUser succeeds",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, namespace.Owner, false).Return(user, 0, nil).Once()
				mock.On("UserGetByUsername", ctx, user2.Username).Return(user2, nil).Once()
				mock.On("NamespaceRemoveMember", ctx, namespace.TenantID, user2.ID).Return(namespace, nil).Once()
			},
			tenantID: namespace.TenantID,
			username: user2.Username,
			ownerID:  namespace.Owner,
			expected: Expected{
				namespace: namespace,
				err:       nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			ns, err := s.RemoveNamespaceUser(ctx, tc.tenantID, tc.username, tc.ownerID)
			assert.Equal(t, tc.expected, Expected{ns, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestGetSessionRecord(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), nil, nil)

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
			expected: Expected{false, ErrNamespaceNotFound},
		},
		{
			name: "GetSessionRecord fails when store namespace get fails",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(nil, Err).Once()
			},
			tenantID: namespace.TenantID,
			expected: Expected{false, Err},
		},
		{
			name: "GetSessionRecord fails when store namespace get fails",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(nil, Err).Once()
			},
			tenantID: namespace.TenantID,
			expected: Expected{false, Err},
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
	s := NewService(store.Store(mock), nil, nil)

	ctx := context.TODO()

	namespace := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "xxxx", Settings: &models.NamespaceSettings{SessionRecord: true}}
	user := &models.User{Name: "user1", Username: "username1", ID: "hash1"}
	user2 := &models.User{Name: "user2", Username: "username2", ID: "hash2"}

	Err := errors.New("error")

	cases := []struct {
		name              string
		requiredMocks     func()
		sessionRecord     bool
		ownerID, tenantID string
		expected          error
	}{
		{
			name:     "RemoveNamespaceUser fails when user is not the owner",
			ownerID:  user2.ID,
			tenantID: namespace.TenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, user2.ID, false).Return(user2, 0, nil).Once()
			},
			expected: ErrUnauthorized,
		},
		{
			name:    "EditSessionRecord fails when namespace set session record fails",
			ownerID: namespace.Owner,
			requiredMocks: func() {
				status := true
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, namespace.Owner, false).Return(user, 0, nil).Once()
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
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, namespace.Owner, false).Return(user, 0, nil).Once()
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
			err := s.EditSessionRecordStatus(ctx, tc.sessionRecord, tc.tenantID, tc.ownerID)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}
