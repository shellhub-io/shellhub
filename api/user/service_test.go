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

	user := &models.User{Name: "name", Email: "oldemail@example.com", Username: "oldusername", Password: "h8h389hhu32r9", ID: "id"}

	user2 := &models.User{Name: "name", Email: "oldemail@example.com", Username: "oldusername2", Password: "bwi3248hj23k", ID: "id2"}

	user3 := &models.User{Name: "name", Email: "new@email.com", Username: "oldusername3", Password: "hasha1b2c3", ID: "id3"}

	updateUser1 := &models.User{Name: "name", Email: "", Username: "newusername", Password: "", ID: "id"}
	updateUser2 := &models.User{Name: "name", Email: "new@email.com", Username: "", Password: "", ID: "id2"}
	updateUser3 := &models.User{Name: "name", Email: "", Username: "", Password: "hasha1b2c3", ID: "id3"}

	//Change username
	mock.On("GetUserByUsername", ctx, updateUser1.Username).Return(user, nil).Once()
	mock.On("GetUserByEmail", ctx, updateUser1.Email).Return(user, nil).Once()
	mock.On("GetUserByID", ctx, updateUser1.ID).Return(user, nil).Once()
	mock.On("UpdateUser", ctx, updateUser1.Name, updateUser1.Username, updateUser1.Email, updateUser1.Password, updateUser1.Password, updateUser1.ID).Return(nil).Once()

	_, err := s.UpdateDataUser(ctx, updateUser1.Name, updateUser1.Username, updateUser1.Email, updateUser1.Password, updateUser1.Password, updateUser1.ID)

	assert.NoError(t, err)
	mock.AssertExpectations(t)

	// Change email
	mock.On("GetUserByUsername", ctx, updateUser2.Username).Return(user2, nil).Once()
	mock.On("GetUserByEmail", ctx, updateUser2.Email).Return(user2, nil).Once()
	mock.On("GetUserByID", ctx, updateUser2.ID).Return(user2, nil).Once()
	mock.On("UpdateUser", ctx, updateUser2.Name, updateUser2.Username, updateUser2.Email, updateUser2.Password, updateUser2.Password, updateUser2.ID).Return(nil).Once()

	_, err = s.UpdateDataUser(ctx, updateUser2.Name, updateUser2.Username, updateUser2.Email, updateUser2.Password, updateUser2.Password, updateUser2.ID)

	assert.NoError(t, err)
	mock.AssertExpectations(t)

	// Change password
	oldPassword := "hasha1b2c3"
	newPassword := "hashd4e5f6"

	mock.On("GetUserByUsername", ctx, updateUser3.Username).Return(user3, nil).Once()
	mock.On("GetUserByEmail", ctx, updateUser3.Email).Return(user3, nil).Once()
	mock.On("GetUserByID", ctx, updateUser3.ID).Return(user3, nil).Once()
	mock.On("UpdateUser", ctx, updateUser3.Name, updateUser3.Username, updateUser3.Email, oldPassword, newPassword, updateUser3.ID).Return(nil).Once()

	_, err = s.UpdateDataUser(ctx, updateUser3.Name, updateUser3.Username, updateUser3.Email, oldPassword, newPassword, updateUser3.ID)

	assert.NoError(t, err)
	mock.AssertExpectations(t)
}
