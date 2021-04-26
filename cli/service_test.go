package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

type Data struct {
	Mock       *mocks.Store
	Service    Service
	User       *models.User
	Namespace  *models.Namespace
	Parameters Parameters
}

func TestAddUserNamespace(t *testing.T) {
	data := initData("usrNs")

	data.Mock.On("NamespaceAddMember", context.TODO(), data.Namespace.TenantID, data.User.ID).Return(data.Namespace, nil).Once()

	ns, err := data.Service.NamespaceAddMember(data.Parameters)
	assert.NoError(t, err)
	assert.Equal(t, data.Namespace.Owner, ns.Owner)
	assert.Equal(t, data.Namespace.Name, ns.Name)

	data.Mock.AssertExpectations(t)
}

func TestDelNamespace(t *testing.T) {
	data := initData("ns")

	data.Mock.On("NamespaceDelete", context.TODO(), data.Namespace.TenantID).Return(nil).Once()

	err := data.Service.NamespaceDelete(data.Parameters)
	assert.NoError(t, err)

	data.Mock.AssertExpectations(t)
}

func TestDelUser(t *testing.T) {
	data := initData("usr")

	data.Mock.On("UserDelete", context.TODO(), data.User.ID).Return(nil).Once()

	err := data.Service.UserDelete(data.Parameters)
	assert.NoError(t, err)

	data.Mock.AssertExpectations(t)
}

func TestDelUserNamespace(t *testing.T) {
	data := initData("usrNs")

	data.Mock.On("NamespaceRemoveMember", context.TODO(), data.Namespace.TenantID, data.User.ID).Return(data.Namespace, nil).Once()

	ns, err := data.Service.NamespaceRemoveMember(data.Parameters)
	assert.NoError(t, err)
	assert.Equal(t, data.Namespace.Owner, ns.Owner)
	assert.Equal(t, data.Namespace.Name, ns.Name)

	data.Mock.AssertExpectations(t)
}

func TestResetUserPassword(t *testing.T) {
	data := initData("usr")

	data.Parameters.Password = "testService2"

	hash := sha256.Sum256([]byte("testService2"))

	data.Mock.On("UserUpdate", context.TODO(), data.User.Username, data.User.Username, data.User.Email, data.User.Password, hex.EncodeToString(hash[:]), data.User.ID).Return(nil).Once()

	err := data.Service.UserUpdate(data.Parameters)
	assert.NoError(t, err)

	data.Mock.AssertExpectations(t)
}

func initData(dataNeeded string) Data {
	var data Data

	data.Mock = &mocks.Store{}
	data.Service = NewService(store.Store(data.Mock))

	hash := sha256.Sum256([]byte("testService"))

	data.User = &models.User{
		Name:     "testService",
		Username: "testService",
		Password: hex.EncodeToString(hash[:]),
		Email:    "test@shellhub.com",
	}

	data.Namespace = &models.Namespace{
		Name:     "testService",
		Owner:    "testService",
		TenantID: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Members:  []string{"1"},

		Settings: &models.NamespaceSettings{
			SessionRecord: true,
		},
	}

	data.Parameters = Parameters{
		Username:  "testService",
		Namespace: "testService",
	}

	switch dataNeeded {
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
