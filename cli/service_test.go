package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
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

func TestNamespaceCreate(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	type Expected struct {
		namespace *models.Namespace
		err       error
	}

	type Test struct {
		description   string
		args          Arguments
		requiredMocks func()
		expected      Expected
	}

	ctx := context.TODO()
	Err := errors.New("error")

	data := initData("none")

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
				mock.On("UserGetByUsername", ctx, data.User.Username).
					Return(nil, Err).Once()
			},
			expected: Expected{nil, ErrUserNotFound},
		},
		{
			description: "Fails when the namespace is duplicated",
			args:        data.Arguments,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, data.User.Username).
					Return(data.User, nil).Once()
				mock.On("NamespaceCreate", context.TODO(), data.Namespace).
					Return(nil, Err).Once()
			},
			expected: Expected{nil, ErrDuplicateNamespace},
		},
		{
			description: "Successfully creates Namespace",
			args:        data.Arguments,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, data.User.Username).
					Return(data.User, nil).Once()
				mock.On("NamespaceCreate", context.TODO(), data.Namespace).
					Return(data.Namespace, nil).Once()
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
	data := initData("usrNs")

	data.Mock.On("NamespaceAddMember", context.TODO(), data.Namespace.TenantID, data.User.ID).Return(data.Namespace, nil).Once()

	ns, err := data.Service.NamespaceAddMember(data.Arguments)
	assert.NoError(t, err)
	assert.Equal(t, data.Namespace.Owner, ns.Owner)
	assert.Equal(t, data.Namespace.Name, ns.Name)

	data.Mock.AssertExpectations(t)
}

func TestDelNamespace(t *testing.T) {
	data := initData("ns")

	data.Mock.On("NamespaceDelete", context.TODO(), data.Namespace.TenantID).Return(nil).Once()

	err := data.Service.NamespaceDelete(data.Arguments)
	assert.NoError(t, err)

	data.Mock.AssertExpectations(t)
}

func TestDelUser(t *testing.T) {
	data := initData("usr")

	data.Mock.On("UserDelete", context.TODO(), data.User.ID).Return(nil).Once()

	err := data.Service.UserDelete(data.Arguments)
	assert.NoError(t, err)

	data.Mock.AssertExpectations(t)
}

func TestDelUserNamespace(t *testing.T) {
	data := initData("usrNs")

	data.Mock.On("NamespaceRemoveMember", context.TODO(), data.Namespace.TenantID, data.User.ID).Return(data.Namespace, nil).Once()

	ns, err := data.Service.NamespaceRemoveMember(data.Arguments)
	assert.NoError(t, err)
	assert.Equal(t, data.Namespace.Owner, ns.Owner)
	assert.Equal(t, data.Namespace.Name, ns.Name)

	data.Mock.AssertExpectations(t)
}

func TestResetUserPassword(t *testing.T) {
	data := initData("usr")

	data.Arguments.Password = "testService2"

	hash := sha256.Sum256([]byte("testService2"))

	data.Mock.On("UserUpdatePassword", context.TODO(), hex.EncodeToString(hash[:]), data.User.ID).Return(nil).Once()

	err := data.Service.UserUpdate(data.Arguments)
	assert.NoError(t, err)

	data.Mock.AssertExpectations(t)
}

func initData(dataNeeded string) Data {
	var data Data

	data.Mock = &mocks.Store{}
	data.Service = NewService(store.Store(data.Mock))

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

	switch dataNeeded {
	case "none":
	case "usr":
		data.Mock.On("UserGetByUsername", context.TODO(), "testService").Return(data.User, nil).Once()
	case "ns":
		data.Mock.On("NamespaceGetByName", context.TODO(), "testService").Return(data.Namespace, nil).Once()
	case "usrNs":
		data.Mock.On("UserGetByUsername", context.TODO(), "testService").Return(data.User, nil).Once()
		data.Mock.On("NamespaceGetByName", context.TODO(), "testService").Return(data.Namespace, nil).Once()
	}

	return data
}
