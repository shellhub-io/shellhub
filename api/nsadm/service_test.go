package nsadm

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
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	namespaces := []models.Namespace{
		{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713"},
		{Name: "group2", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf48471i4"},
	}
	query := paginator.Query{Page: 1, PerPage: 10}
	mock.On("NamespaceList", ctx, query, []models.Filter(nil), false).Return(namespaces, len(namespaces), nil).Once()
	returnedNamespaces, count, err := s.ListNamespaces(ctx, query, "", false)
	assert.NoError(t, err)
	assert.Equal(t, namespaces, returnedNamespaces)
	assert.Equal(t, count, len(namespaces))
	mock.AssertExpectations(t)
}

func TestGetNamespace(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	namespace := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713"}

	mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()

	returnNamespace, err := s.GetNamespace(ctx, namespace.TenantID)
	assert.NoError(t, err)
	assert.Equal(t, namespace, returnNamespace)

	mock.AssertExpectations(t)
}

func TestCreateNamespace(t *testing.T) {
	cases := []struct {
		name       string
		isCloud    bool
		maxDevices int
	}{
		{
			name:       "CloudInstance",
			isCloud:    true,
			maxDevices: 3,
		},
		{
			name:       "EnterpriseOrCommunityInstance",
			isCloud:    false,
			maxDevices: -1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mocks.Store{}
			s := NewService(store.Store(mock))

			ctx := context.TODO()
			namespace := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "tenant"}

			envMock := &env_mocks.Backend{}
			envs.DefaultBackend = envMock
			envMock.On("Get", "SHELLHUB_CLOUD").Return(strconv.FormatBool(tc.isCloud)).Once()

			user := &models.User{Name: "user1", Username: "hash1", ID: "hash1"}
			createNamespace := &models.Namespace{
				Name:       strings.ToLower(namespace.Name),
				Owner:      user.ID,
				Members:    []interface{}{user.ID},
				Settings:   &models.NamespaceSettings{SessionRecord: true},
				TenantID:   namespace.TenantID,
				MaxDevices: tc.maxDevices,
			}

			mock.On("NamespaceGetByName", ctx, createNamespace.Name).Return(nil, nil).Once()
			mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
			mock.On("NamespaceCreate", ctx, createNamespace).Return(createNamespace, nil).Once()

			returnedNamespace, err := s.CreateNamespace(ctx, createNamespace, namespace.Owner)
			assert.NoError(t, err)
			assert.Equal(t, createNamespace, returnedNamespace)
			mock.AssertExpectations(t)
		})
	}
}

func TestEditNamespace(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	user := &models.User{Name: "user1", Username: "hash1", ID: "hash1"}

	newName := "newname"
	namespace := &models.Namespace{Name: "oldname", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713"}
	namespaceWithNewName := &models.Namespace{Name: "newname", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713"}

	mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Twice()
	mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
	mock.On("NamespaceRename", ctx, namespace.TenantID, newName).Return(namespaceWithNewName, nil).Once()
	_, err := s.EditNamespace(ctx, namespace.TenantID, newName, namespace.Owner)

	assert.NoError(t, err)

	mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespaceWithNewName, nil).Once()
	returnedNamespace, err := s.GetNamespace(ctx, namespace.TenantID)
	assert.NoError(t, err)
	assert.Equal(t, newName, returnedNamespace.Name)

	mock.AssertExpectations(t)
}

func TestDeleteNamespace(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	user := &models.User{Name: "user1", Username: "hash1", ID: "hash1"}
	namespace := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713"}

	mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
	mock.On("NamespaceDelete", ctx, namespace.TenantID).Return(nil).Once()
	mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()

	err := s.DeleteNamespace(ctx, namespace.TenantID, namespace.Owner)
	assert.NoError(t, err)

	mock.AssertExpectations(t)
}

func TestAddNamespaceUser(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))
	ctx := context.TODO()
	user := &models.User{Name: "user1", Username: "username1", ID: "hash1"}
	member := &models.User{Name: "user2", Username: "username2", ID: "hash2"}
	namespace := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Members: []interface{}{"hash1"}}
	namespace2 := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Members: []interface{}{"hash1", "hash2"}}

	mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
	mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
	mock.On("UserGetByUsername", ctx, member.Username).Return(member, nil).Once()
	mock.On("NamespaceAddMember", ctx, namespace.TenantID, member.ID).Return(namespace2, nil).Once()

	_, err := s.AddNamespaceUser(ctx, namespace.TenantID, member.Username, user.ID)
	assert.NoError(t, err)

	mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace2, nil).Once()

	returnedNamespace, err := s.GetNamespace(ctx, namespace.TenantID)
	assert.NoError(t, err)
	assert.Equal(t, namespace2, returnedNamespace)
}

func TestRemoveNamespaceUser(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))
	ctx := context.TODO()
	user := &models.User{Name: "user1", Username: "username1", ID: "hash1"}
	member := &models.User{Name: "user2", Username: "username2", ID: "hash2"}
	namespace := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Members: []interface{}{"hash1", "hash2"}}
	namespace2 := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Members: []interface{}{"hash1"}}

	mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
	mock.On("UserGetByID", ctx, user.ID, false).Return(user, 0, nil).Once()
	mock.On("UserGetByUsername", ctx, member.Username).Return(member, nil).Once()
	mock.On("NamespaceRemoveMember", ctx, namespace.TenantID, member.ID).Return(namespace2, nil).Once()

	_, err := s.RemoveNamespaceUser(ctx, namespace.TenantID, member.Username, user.ID)
	assert.NoError(t, err)

	mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace2, nil).Once()

	returnedNamespace, err := s.GetNamespace(ctx, namespace.TenantID)
	assert.NoError(t, err)
	assert.Equal(t, namespace2, returnedNamespace)
}

func TestGetSessionRecord(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	namespace := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Settings: &models.NamespaceSettings{SessionRecord: true}}

	mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
	mock.On("NamespaceGetSessionRecord", ctx, namespace.TenantID).
		Return(namespace.Settings.SessionRecord, nil).Once()

	returnedUserSecurity, err := s.GetSessionRecord(ctx, namespace.TenantID)
	assert.NoError(t, err)
	assert.Equal(t, returnedUserSecurity, namespace.Settings.SessionRecord)

	mock.AssertExpectations(t)
}

func TestEditSessionRecord(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	namespace := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Settings: &models.NamespaceSettings{SessionRecord: true}}
	namespace2 := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Settings: &models.NamespaceSettings{SessionRecord: false}}

	mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
	mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace2, nil).Once()
	mock.On("NamespaceSetSessionRecord", ctx, !namespace.Settings.SessionRecord, namespace.TenantID).
		Return(nil).Once()
	mock.On("NamespaceGetSessionRecord", ctx, namespace.TenantID).
		Return(!namespace.Settings.SessionRecord, nil).Once()

	err := s.EditSessionRecordStatus(ctx, !namespace.Settings.SessionRecord, namespace.TenantID)
	assert.NoError(t, err)

	returnedUserSecurity, err := s.GetSessionRecord(ctx, namespace.TenantID)
	assert.NoError(t, err)
	assert.Equal(t, !namespace.Settings.SessionRecord, returnedUserSecurity)

	mock.AssertExpectations(t)
}
