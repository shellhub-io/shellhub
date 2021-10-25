package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/authorizer"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestDelUser(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	Err := errors.New("error")
	userNotFound := &models.User{
		UserData: models.UserData{
			Name:     "userNotFound",
			Email:    "userNotFound@userNotFound.com",
			Username: "usernameNotFound",
		},
	}
	user := &models.User{
		ID: "userID",
		UserData: models.UserData{
			Name:     "user",
			Email:    "user@user.com",
			Username: "username",
		},
	}
	namespaceOwned := []*models.Namespace{
		{
			Name:     "namespace1",
			Owner:    user.ID,
			TenantID: "tenantID1",
			Members:  []models.Member{{ID: user.ID, Type: "owner"}},
			Settings: &models.NamespaceSettings{
				SessionRecord: true,
			},
			CreatedAt: clock.Now(),
		},
		{
			Name:     "namespace2",
			Owner:    user.ID,
			TenantID: "tenantID2",
			Members:  []models.Member{{ID: user.ID, Type: "owner"}},
			Settings: &models.NamespaceSettings{
				SessionRecord: true,
			},
			CreatedAt: clock.Now(),
		},
	}
	namespaceMember := []*models.Namespace{
		{
			Name:     "namespace3",
			Owner:    "ownerID",
			TenantID: "tenantID3",
			Members:  []models.Member{{ID: "ownerID", Type: authorizer.MemberTypeObserver}, {ID: user.ID, Type: authorizer.MemberTypeObserver}},
			Settings: &models.NamespaceSettings{
				SessionRecord: true,
			},
			CreatedAt: clock.Now(),
		},
		{
			Name:     "namespace1",
			Owner:    "ownerID",
			TenantID: "tenantID1",
			Members:  []models.Member{{ID: "ownerID", Type: authorizer.MemberTypeObserver}, {ID: user.ID, Type: authorizer.MemberTypeObserver}},
			Settings: &models.NamespaceSettings{
				SessionRecord: true,
			},
			CreatedAt: clock.Now(),
		},
	}

	tests := []struct {
		description   string
		username      string
		requiredMocks func()
		expected      error
	}{
		{
			description: "Fails to find the user",
			username:    userNotFound.Username,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, userNotFound.Username).Return(nil, Err).Once()
			},
			expected: ErrUserNotFound,
		},
		{
			description: "Fail to delete the user and associated data",
			username:    user.Username,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, user.Username).Return(user, nil).Once()
				mock.On("UserDetachInfo", ctx, user.ID).Return(nil, Err).Once()
			},
			expected: ErrNamespaceNotFound,
		},
		{
			description: "Successfully delete the user and associated data",
			username:    user.Username,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, user.Username).Return(user, nil).Once()
				mock.On("UserDetachInfo", ctx, user.ID).Return(map[string][]*models.Namespace{
					"owner":  namespaceOwned,
					"member": namespaceMember,
				}, nil)
				for _, v := range namespaceOwned {
					mock.On("NamespaceDelete", ctx, v.TenantID).Return(nil).Once()
				}
				for _, v := range namespaceMember {
					mock.On("NamespaceRemoveMember", ctx, v.TenantID, user.ID).Return(nil, nil).Once()
				}
				mock.On("UserDelete", ctx, user.ID).Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, ts := range tests {
		test := ts
		t.Run(test.description, func(t *testing.T) {
			test.requiredMocks()
			err := s.UserDelete(test.username)
			assert.Equal(t, test.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestResetUserPassword(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	Err := errors.New("error")

	userPasswordInvalid := models.UserPassword{
		Password: "ab",
	}
	userPassword := models.UserPassword{
		Password: "password",
	}
	user := &models.User{
		ID: "userID",
		UserData: models.UserData{
			Name:     "user",
			Email:    "user@user.com",
			Username: "username",
		},
	}
	userNotFound := &models.User{
		UserData: models.UserData{
			Name:     "userNotFound",
			Email:    "userNotFound@userNotFound.com",
			Username: "usernameNotFound",
		},
	}

	tests := []struct {
		description   string
		username      string
		password      string
		requiredMocks func()
		expected      error
	}{
		{
			description: "Fails when the field is invalid",
			username:    user.Username,
			password:    userPasswordInvalid.Password,
			requiredMocks: func() {
			},
			expected: ErrUserPasswordInvalid,
		},
		{
			description: "Fails to find the user",
			username:    userNotFound.Username,
			password:    userPassword.Password,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, userNotFound.Username).Return(nil, Err).Once()
			},
			expected: ErrUserNotFound,
		},
		{
			description: "Fail reset the user password",
			username:    user.Username,
			password:    userPassword.Password,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, user.Username).Return(user, nil).Once()
				mock.On("UserUpdatePassword", ctx, hashPassword(userPassword.Password), user.ID).Return(Err).Once()
			},
			expected: ErrFailedUpdateUser,
		},
		{
			description: "Successfully reset the user password",
			username:    user.Username,
			password:    userPassword.Password,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, user.Username).Return(user, nil).Once()
				mock.On("UserUpdatePassword", ctx, hashPassword(userPassword.Password), user.ID).Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, ts := range tests {
		test := ts
		t.Run(test.description, func(t *testing.T) {
			test.requiredMocks()
			err := s.UserUpdate(test.username, test.password)
			assert.Equal(t, test.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestNamespaceCreate(t *testing.T) {
	mockClock := &clockmock.Clock{}

	clock.DefaultBackend = mockClock

	now := time.Now()

	mockClock.On("Now").Return(now)

	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	Err := errors.New("error")

	user := &models.User{
		ID: "userID",
		UserData: models.UserData{
			Name:     "user",
			Email:    "user@user.com",
			Username: "username",
		},
	}
	userNotFound := &models.User{
		UserData: models.UserData{
			Name:     "userNotFound",
			Email:    "userNotFound@userNotFound.com",
			Username: "usernameNotFound",
		},
	}
	namespace := &models.Namespace{
		Name:     "namespace",
		Owner:    user.ID,
		TenantID: "tenantID",
		Members:  []models.Member{{ID: user.ID, Type: "owner"}},
		Settings: &models.NamespaceSettings{
			SessionRecord: true,
		},
		CreatedAt: now,
	}
	namespaceInvalid := &models.Namespace{
		Name:     "namespaceInvalid@@",
		Owner:    "",
		TenantID: "tenantIDInvliad",
	}

	type Expected struct {
		user *models.Namespace
		err  error
	}
	tests := []struct {
		description   string
		namespace     string
		username      string
		tenantID      string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "Fails when could not find a user",
			namespace:   namespace.Name,
			username:    userNotFound.Username,
			tenantID:    namespace.TenantID,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, userNotFound.Username).Return(nil, Err).Once()
			},
			expected: Expected{nil, ErrUserNotFound},
		},
		{
			description: "Fails when namespace is not valid",
			namespace:   namespaceInvalid.Name,
			username:    user.Username,
			tenantID:    namespaceInvalid.TenantID,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, user.Username).Return(user, nil).Once()
			},
			expected: Expected{nil, ErrNamespaceInvalid},
		},
		{
			description: "Fails when namespace is duplicated",
			namespace:   namespace.Name,
			username:    user.Username,
			tenantID:    namespace.TenantID,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, user.Username).Return(user, nil).Once()
				mock.On("NamespaceCreate", ctx, namespace).Return(nil, Err).Once()
			},
			expected: Expected{nil, ErrDuplicateNamespace},
		},
		{
			description: "Success to create namespace",
			namespace:   namespace.Name,
			username:    user.Username,
			tenantID:    namespace.TenantID,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, user.Username).Return(user, nil).Once()
				mock.On("NamespaceCreate", ctx, namespace).Return(namespace, nil).Once()
			},
			expected: Expected{namespace, nil},
		},
	}

	for _, ts := range tests {
		test := ts
		t.Run(test.description, func(t *testing.T) {
			test.requiredMocks()
			ns, err := s.NamespaceCreate(test.namespace, test.username, test.tenantID)
			assert.Equal(t, test.expected, Expected{ns, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestAddUserNamespace(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	Err := errors.New("error")

	user := &models.User{
		ID: "userID",
		UserData: models.UserData{
			Name:     "user",
			Email:    "user@user.com",
			Username: "username",
		},
	}
	userNotFound := &models.User{
		UserData: models.UserData{
			Name:     "userNotFound",
			Email:    "userNotFound@userNotFound.com",
			Username: "usernameNotFound",
		},
	}
	namespace := &models.Namespace{
		Name:     "namespace",
		Owner:    user.ID,
		TenantID: "tenantID",
		Members:  []models.Member{{ID: user.ID, Type: "owner"}},
		Settings: &models.NamespaceSettings{
			SessionRecord: true,
		},
		CreatedAt: clock.Now(),
	}

	namespaceNotFound := &models.Namespace{
		Name:     "namespaceNotFound",
		Owner:    user.ID,
		TenantID: "tenantIDNotFound",
		Members:  []models.Member{{ID: user.ID, Type: "owner"}},
		Settings: &models.NamespaceSettings{
			SessionRecord: true,
		},
		CreatedAt: clock.Now(),
	}

	type Expected struct {
		user *models.Namespace
		err  error
	}
	tests := []struct {
		description   string
		username      string
		namespace     string
		accessType    string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "Fails to find the user",
			username:    userNotFound.Username,
			namespace:   namespace.Name,
			accessType:  authorizer.MemberTypeObserver,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, userNotFound.Username).Return(nil, Err).Once()
			},
			expected: Expected{nil, ErrUserNotFound},
		},
		{
			description: "Fails to find the namespace",
			username:    user.Username,
			namespace:   namespaceNotFound.Name,
			accessType:  authorizer.MemberTypeObserver,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, user.Username).Return(user, nil).Once()
				mock.On("NamespaceGetByName", ctx, namespaceNotFound.Name).Return(nil, Err).Once()
			},
			expected: Expected{nil, ErrNamespaceNotFound},
		},
		{
			description: "Successfully add user to the Namespace",
			username:    user.Username,
			namespace:   namespace.Name,
			accessType:  authorizer.MemberTypeObserver,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, user.Username).Return(user, nil).Once()
				mock.On("NamespaceGetByName", ctx, namespace.Name).Return(namespace, nil).Once()
				mock.On("NamespaceAddMember", ctx, namespace.TenantID, user.ID, authorizer.MemberTypeObserver).Return(namespace, nil).Once()
			},
			expected: Expected{namespace, nil},
		},
	}

	for _, ts := range tests {
		test := ts
		t.Run(test.description, func(t *testing.T) {
			test.requiredMocks()
			ns, err := s.NamespaceAddMember(test.username, test.namespace, test.accessType)
			assert.Equal(t, test.expected, Expected{ns, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestDelUserNamespace(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	Err := errors.New("error")

	user := &models.User{
		ID: "userID",
		UserData: models.UserData{
			Name:     "user",
			Email:    "user@user.com",
			Username: "username",
		},
	}
	userNotFound := &models.User{
		UserData: models.UserData{
			Name:     "userNotFound",
			Email:    "userNotFound@userNotFound.com",
			Username: "usernameNotFound",
		},
	}
	namespace := &models.Namespace{
		Name:     "namespace",
		Owner:    user.ID,
		TenantID: "tenantID",
		Members:  []models.Member{{ID: user.ID, Type: "owner"}},
		Settings: &models.NamespaceSettings{
			SessionRecord: true,
		},
		CreatedAt: clock.Now(),
	}
	namespaceNotFound := &models.Namespace{
		Name:     "namespaceNotFound",
		Owner:    user.ID,
		TenantID: "tenantIDNotFound",
		Members:  []models.Member{{ID: user.ID, Type: "owner"}},
		Settings: &models.NamespaceSettings{
			SessionRecord: true,
		},
		CreatedAt: clock.Now(),
	}

	type Expected struct {
		user *models.Namespace
		err  error
	}
	tests := []struct {
		description   string
		username      string
		namespace     string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "Fails to find the user",
			username:    userNotFound.Username,
			namespace:   namespace.Name,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, userNotFound.Username).Return(nil, Err).Once()
			},
			expected: Expected{nil, ErrUserNotFound},
		},
		{
			description: "Fails to find the namespace",
			username:    user.Username,
			namespace:   namespaceNotFound.Name,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, user.Username).Return(user, nil).Once()
				mock.On("NamespaceGetByName", ctx, namespaceNotFound.Name).Return(nil, Err).Once()
			},
			expected: Expected{nil, ErrNamespaceNotFound},
		},
		{
			description: "Fails remove member from the namespace",
			username:    user.Username,
			namespace:   namespace.Name,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, user.Username).Return(user, nil).Once()
				mock.On("NamespaceGetByName", ctx, namespace.Name).Return(namespace, nil).Once()
				mock.On("NamespaceRemoveMember", ctx, namespace.TenantID, user.ID).Return(nil, Err).Once()
			},
			expected: Expected{nil, ErrFailedNamespaceRemoveMember},
		},
		{
			description: "Successfully remove member from the namespace",
			username:    user.Username,
			namespace:   namespace.Name,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, user.Username).Return(user, nil).Once()
				mock.On("NamespaceGetByName", ctx, namespace.Name).Return(namespace, nil).Once()
				mock.On("NamespaceRemoveMember", ctx, namespace.TenantID, user.ID).Return(namespace, nil).Once()
			},
			expected: Expected{namespace, nil},
		},
	}

	for _, ts := range tests {
		test := ts
		t.Run(test.description, func(t *testing.T) {
			test.requiredMocks()
			ns, err := s.NamespaceRemoveMember(test.username, test.namespace)
			assert.Equal(t, test.expected, Expected{ns, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestDelNamespace(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	Err := errors.New("error")
	user := &models.User{
		ID: "userID",
		UserData: models.UserData{
			Name:     "user",
			Email:    "user@user.com",
			Username: "username",
		},
	}
	namespace := &models.Namespace{
		Name:     "namespace",
		Owner:    user.ID,
		TenantID: "tenantID",
		Members:  []models.Member{{ID: user.ID, Type: "owner"}},
		Settings: &models.NamespaceSettings{
			SessionRecord: true,
		},
		CreatedAt: clock.Now(),
	}
	namespaceNotFound := &models.Namespace{
		Name:     "namespaceNotFound",
		Owner:    user.ID,
		TenantID: "tenantIDNotFound",
		Members:  []models.Member{{ID: user.ID, Type: "owner"}},
		Settings: &models.NamespaceSettings{
			SessionRecord: true,
		},
		CreatedAt: clock.Now(),
	}

	tests := []struct {
		description   string
		namespace     string
		requiredMocks func()
		expected      error
	}{
		{
			description: "Fails to find the namespace",
			namespace:   namespaceNotFound.Name,
			requiredMocks: func() {
				mock.On("NamespaceGetByName", ctx, namespaceNotFound.Name).Return(nil, Err).Once()
			},
			expected: ErrNamespaceNotFound,
		},
		{
			description: "Fails to delete the namespace",
			namespace:   namespace.Name,
			requiredMocks: func() {
				mock.On("NamespaceGetByName", ctx, namespace.Name).Return(namespace, nil).Once()
				mock.On("NamespaceDelete", ctx, namespace.TenantID).Return(Err).Once()
			},
			expected: ErrFailedDeleteNamespace,
		},
		{
			description: "Success to delete the namespace",
			namespace:   namespace.Name,
			requiredMocks: func() {
				mock.On("NamespaceGetByName", ctx, namespace.Name).Return(namespace, nil).Once()
				mock.On("NamespaceDelete", ctx, namespace.TenantID).Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, ts := range tests {
		test := ts
		t.Run(test.description, func(t *testing.T) {
			test.requiredMocks()
			err := s.NamespaceDelete(test.namespace)
			assert.Equal(t, test.expected, err)
		})
	}

	mock.AssertExpectations(t)
}
