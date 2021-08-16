package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

type Data struct {
	Mock      *mocks.Store
	Service   Service
	User      *models.User
	Namespace *models.Namespace
	Arguments Arguments
}

type Test struct {
	description   string
	args          Arguments
	requiredMocks func()
	expected      Expected
}

type Expected struct {
	namespace *models.Namespace
	err       error
}

func TestNamespaceCreate(t *testing.T) {
	data := initData()

	mockClock := &clockmock.Clock{}

	clock.DefaultBackend = mockClock

	now := time.Now()

	mockClock.On("Now").Return(now).Twice()

	data.Namespace.CreatedAt = now

	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	Err := errors.New("error")

	tests := []Test{
		{
			description: "Fails when the field is invalid",
			args:        Arguments{Username: "user", Namespace: "ns3ww", Email: "invalid@email", TenantID: "xxx", Password: "password"},
			requiredMocks: func() {
			},
			expected: Expected{nil, ErrCreateNewNamespace},
		},
		{
			description: "Fails to find the user",
			args:        data.Arguments,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, data.User.Username).Return(nil, Err).Once()
			},
			expected: Expected{nil, ErrUserNotFound},
		},
		{
			description: "Fails when the namespace is duplicated",
			args:        data.Arguments,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, data.User.Username).Return(data.User, nil).Once()
				mock.On("NamespaceCreate", ctx, data.Namespace).Return(nil, Err).Once()
			},
			expected: Expected{nil, ErrDuplicateNamespace},
		},
		{
			description: "Successfully creates Namespace",
			args:        data.Arguments,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, data.User.Username).Return(data.User, nil).Once()
				mock.On("NamespaceCreate", ctx, data.Namespace).Return(data.Namespace, nil).Once()
			},
			expected: Expected{data.Namespace, nil},
		},
	}

	for _, test := range tests {
		t.Log("PASS:  ", test.description)
		test.requiredMocks()
		ns, err := s.NamespaceCreate(test.args)
		assert.Equal(t, test.expected, Expected{ns, err})
	}

	mock.AssertExpectations(t)
}

func TestAddUserNamespace(t *testing.T) {
	data := initData()

	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	Err := errors.New("error")

	tests := []Test{
		{
			description: "Fails to find the user",
			args:        data.Arguments,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, data.User.Username).Return(nil, Err).Once()
			},
			expected: Expected{nil, ErrUserNotFound},
		},
		{
			description: "Fails to find the namespace",
			args:        data.Arguments,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, data.User.Username).Return(data.User, nil).Once()
				mock.On("NamespaceGetByName", ctx, data.User.Username).Return(nil, Err).Once()
			},
			expected: Expected{nil, ErrNamespaceNotFound},
		},
		{
			description: "Successfully add user to the Namespace",
			args:        data.Arguments,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, data.User.Username).Return(data.User, nil).Once()
				mock.On("NamespaceGetByName", ctx, data.User.Username).Return(data.Namespace, nil).Once()
				mock.On("NamespaceAddMember", ctx, data.Namespace.TenantID, data.User.ID).Return(data.Namespace, nil).Once()
			},
			expected: Expected{data.Namespace, nil},
		},
	}

	for _, test := range tests {
		t.Log("PASS:  ", test.description)
		test.requiredMocks()
		ns, err := s.NamespaceAddMember(test.args)
		assert.Equal(t, test.expected, Expected{ns, err})
	}

	mock.AssertExpectations(t)
}

func TestDelNamespace(t *testing.T) {
	data := initData()

	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	Err := errors.New("error")

	tests := []Test{
		{
			description: "Fails to find the user",
			args:        data.Arguments,
			requiredMocks: func() {
				mock.On("NamespaceGetByName", ctx, data.User.Username).Return(nil, Err).Once()
			},
			expected: Expected{nil, ErrNamespaceNotFound},
		},
		{
			description: "Successfully delete the Namespace",
			args:        data.Arguments,
			requiredMocks: func() {
				mock.On("NamespaceGetByName", ctx, data.User.Username).Return(data.Namespace, nil).Once()
				mock.On("NamespaceDelete", ctx, data.Namespace.TenantID).Return(nil).Once()
			},
			expected: Expected{nil, nil},
		},
	}

	for _, test := range tests {
		t.Log("PASS:  ", test.description)
		test.requiredMocks()
		err := s.NamespaceDelete(test.args)
		assert.Equal(t, test.expected, Expected{nil, err})
	}

	mock.AssertExpectations(t)
}

func TestDelUser(t *testing.T) {
	data := initData()

	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	Err := errors.New("error")

	tests := []Test{
		{
			description: "Fails to find the user",
			args:        data.Arguments,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, data.User.Username).Return(nil, Err).Once()
			},
			expected: Expected{nil, ErrUserNotFound},
		},
		{
			description: "Successfully delete the user",
			args:        data.Arguments,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, data.User.Username).Return(data.User, nil).Once()
				mock.On("UserDelete", ctx, data.User.ID).Return(nil).Once()
			},
			expected: Expected{nil, nil},
		},
	}

	for _, test := range tests {
		t.Log("PASS:  ", test.description)
		test.requiredMocks()
		err := s.UserDelete(test.args)
		assert.Equal(t, test.expected, Expected{nil, err})
	}

	mock.AssertExpectations(t)
}

func TestDelUserNamespace(t *testing.T) {
	data := initData()

	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	Err := errors.New("error")

	tests := []Test{
		{
			description: "Fails to find the user",
			args:        data.Arguments,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, data.User.Username).Return(nil, Err).Once()
			},
			expected: Expected{nil, ErrUserNotFound},
		},
		{
			description: "Fails to find the namespace",
			args:        data.Arguments,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, data.User.Username).Return(data.User, nil).Once()
				mock.On("NamespaceGetByName", ctx, data.User.Username).Return(nil, Err).Once()
			},
			expected: Expected{nil, ErrNamespaceNotFound},
		},
		{
			description: "Successfully remove member from the namespace",
			args:        data.Arguments,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, data.User.Username).Return(data.User, nil).Once()
				mock.On("NamespaceGetByName", ctx, data.User.Username).Return(data.Namespace, nil).Once()
				mock.On("NamespaceRemoveMember", ctx, data.Namespace.TenantID, data.User.ID).Return(data.Namespace, nil).Once()
			},
			expected: Expected{data.Namespace, nil},
		},
	}

	for _, test := range tests {
		t.Log("PASS:  ", test.description)
		test.requiredMocks()
		ns, err := s.NamespaceRemoveMember(test.args)
		assert.Equal(t, test.expected, Expected{ns, err})
	}

	mock.AssertExpectations(t)
}

func TestResetUserPassword(t *testing.T) {
	data := initData()
	data.Arguments.Password = "testService2"

	hash := sha256.Sum256([]byte("testService2"))

	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	Err := errors.New("error")

	tests := []Test{
		{
			description: "Fails when the field is invalid",
			args:        Arguments{Username: "user", Password: "pa"},
			requiredMocks: func() {
			},
			expected: Expected{nil, ErrChangePassword},
		},
		{
			description: "Fails to find the user",
			args:        data.Arguments,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, data.User.Username).Return(nil, Err).Once()
			},
			expected: Expected{nil, ErrUserNotFound},
		},
		{
			description: "Successfully reset the user password",
			args:        data.Arguments,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, data.User.Username).Return(data.User, nil).Once()
				mock.On("UserUpdatePassword", ctx, hex.EncodeToString(hash[:]), data.User.ID).Return(nil).Once()
			},
			expected: Expected{nil, nil},
		},
	}

	for _, test := range tests {
		t.Log("PASS:  ", test.description)
		test.requiredMocks()
		err := s.UserUpdate(test.args)
		assert.Equal(t, test.expected, Expected{nil, err})
	}

	mock.AssertExpectations(t)
}

func initData() Data {
	var data Data

	hash := sha256.Sum256([]byte("testService"))

	data.User = &models.User{
		ID:       "1",
		Name:     "testService",
		Username: "testService",
		Password: hex.EncodeToString(hash[:]),
		Email:    "test@shellhub.com",
	}

	data.Namespace = &models.Namespace{
		Name:     "testService",
		Owner:    "1",
		TenantID: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Members:  []interface{}{"1"},
		Settings: &models.NamespaceSettings{
			SessionRecord: true,
		},
	}

	data.Arguments = Arguments{
		Username:  "testService",
		Namespace: "testService",
		TenantID:  "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
	}

	return data
}
