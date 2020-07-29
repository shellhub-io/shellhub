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

	oldPassword := "oldpass123"
	newPassword := "newpass123"

	user := &models.User{Name: "name", Email: "oldemail@example.com", Username: "oldusername", Password: "currentPassword", TenantID: "tenant"}

	user2 := &models.User{Name: "name", Email: "oldemail@example.com", Username: "newusername", Password: "currentPassword", TenantID: "tenant"}

	user3 := &models.User{Name: "name", Email: "new@email.com", Username: "newusername", Password: "oldpass123", TenantID: "tenant"}

	updateUser1 := &models.User{Name: "name", Email: "", Username: "newusername", Password: "", TenantID: "tenant"}
	updateUser2 := &models.User{Name: "name", Email: "new@email.com", Username: "", Password: "", TenantID: "tenant"}
	updateUser3 := &models.User{Name: "name", Email: "", Username: "", Password: "oldpass123", TenantID: "tenant"}

	//Change username
	mock.On("GetUserByTenant", ctx, "tenant").Return(user, nil).Once()
	mock.On("UpdateUser", ctx, updateUser1.Username, updateUser1.Email, updateUser1.Password, updateUser1.Password, updateUser1.TenantID).Return(nil).Once()

	err := s.UpdateDataUser(ctx, updateUser1.Username, updateUser1.Email, updateUser1.Password, updateUser1.Password, updateUser1.TenantID)

	assert.NoError(t, err)
	mock.AssertExpectations(t)

	// changed email
	mock.On("GetUserByTenant", ctx, "tenant").Return(user2, nil).Once()
	mock.On("UpdateUser", ctx, updateUser2.Username, updateUser2.Email, updateUser2.Password, updateUser2.Password, updateUser2.TenantID).Return(nil).Once()

	err = s.UpdateDataUser(ctx, updateUser2.Username, updateUser2.Email, updateUser2.Password, updateUser2.Password, updateUser2.TenantID)

	assert.NoError(t, err)
	mock.AssertExpectations(t)

	// changed password
	mock.On("GetUserByTenant", ctx, "tenant").Return(user3, nil).Once()
	mock.On("UpdateUser", ctx, updateUser3.Username, updateUser3.Email, oldPassword, newPassword, updateUser3.TenantID).Return(nil).Once()

	err = s.UpdateDataUser(ctx, updateUser3.Username, updateUser3.Email, oldPassword, newPassword, updateUser3.TenantID)

	assert.NoError(t, err)
	mock.AssertExpectations(t)
}
