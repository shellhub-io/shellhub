package user

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestUpdateDataUser(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	user := &models.User{Name: "name", Email: "oldemail@example.com", Username: "oldusername", Password: "h8h389hhu32r9", TenantID: "tenant"}

	user2 := &models.User{Name: "name", Email: "oldemail@example.com", Username: "oldusername2", Password: "bwi3248hj23k", TenantID: "tenant1"}

	user3 := &models.User{Name: "name", Email: "new@email.com", Username: "oldusername3", Password: "hasha1b2c3", TenantID: "tenant2"}

	updateUser1 := &models.User{Name: "name", Email: "", Username: "newusername", Password: "", TenantID: "tenant"}
	updateUser2 := &models.User{Name: "name", Email: "new@email.com", Username: "", Password: "", TenantID: "tenant1"}
	updateUser3 := &models.User{Name: "name", Email: "", Username: "", Password: "hasha1b2c3", TenantID: "tenant2"}

	//Change username
	mock.On("GetUserByUsername", ctx, updateUser1.Username).Return(user, nil).Once()
	mock.On("GetUserByEmail", ctx, updateUser1.Email).Return(user, nil).Once()
	mock.On("GetUserByTenant", ctx, updateUser1.TenantID).Return(user, nil).Once()
	mock.On("UpdateUser", ctx, updateUser1.Username, updateUser1.Email, updateUser1.Password, updateUser1.Password, updateUser1.TenantID).Return(nil).Once()

	err := s.UpdateDataUser(ctx, updateUser1.Username, updateUser1.Email, updateUser1.Password, updateUser1.Password, updateUser1.TenantID)

	assert.NoError(t, err)
	mock.AssertExpectations(t)

	// Change email
	mock.On("GetUserByUsername", ctx, updateUser2.Username).Return(user2, nil).Once()
	mock.On("GetUserByEmail", ctx, updateUser2.Email).Return(user2, nil).Once()
	mock.On("GetUserByTenant", ctx, updateUser2.TenantID).Return(user2, nil).Once()
	mock.On("UpdateUser", ctx, updateUser2.Username, updateUser2.Email, updateUser2.Password, updateUser2.Password, updateUser2.TenantID).Return(nil).Once()

	err = s.UpdateDataUser(ctx, updateUser2.Username, updateUser2.Email, updateUser2.Password, updateUser2.Password, updateUser2.TenantID)

	assert.NoError(t, err)
	mock.AssertExpectations(t)

	// Change password
	oldPassword := "hasha1b2c3"
	newPassword := "hashd4e5f6"

	mock.On("GetUserByUsername", ctx, updateUser3.Username).Return(user3, nil).Once()
	mock.On("GetUserByEmail", ctx, updateUser3.Email).Return(user3, nil).Once()
	mock.On("GetUserByTenant", ctx, updateUser3.TenantID).Return(user3, nil).Once()
	mock.On("UpdateUser", ctx, updateUser3.Username, updateUser3.Email, oldPassword, newPassword, updateUser3.TenantID).Return(nil).Once()

	err = s.UpdateDataUser(ctx, updateUser3.Username, updateUser3.Email, oldPassword, newPassword, updateUser3.TenantID)

	assert.NoError(t, err)
	mock.AssertExpectations(t)
}
