package services

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/validator"
	"github.com/stretchr/testify/assert"
)

func TestUpdateDataUser(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), nil, nil)

	ctx := context.TODO()

	user1 := &models.User{Name: "name", Email: "user1@email.com", Username: "username1", Password: "hash1", ID: "id1"}

	user2 := &models.User{Name: "name", Email: "user2@email.com", Username: "username2", Password: "hash2", ID: "id2"}

	updateUser1 := &models.User{Name: "name", Email: "user1@email2.com", Username: user2.Username, ID: "id1"}

	Err := errors.New("conflict")

	// verifies existent username only
	mock.On("UserGetByID", ctx, updateUser1.ID, false).Return(user1, 0, nil).Once()
	mock.On("UserGetByUsername", ctx, updateUser1.Username).Return(user2, nil).Once()
	mock.On("UserGetByEmail", ctx, updateUser1.Email).Return(nil, Err).Once()
	fields, err := s.UpdateDataUser(ctx, updateUser1, updateUser1.ID)

	assert.Equal(t, err, ErrConflict)
	assert.Equal(t, fields, []validator.InvalidField{{"username", "conflict", "", ""}})

	updateUser1 = &models.User{Name: "name", Email: "user2@email.com", Username: user2.Username, ID: "id1"}

	// verifies existent email and username only
	mock.On("UserGetByID", ctx, updateUser1.ID, false).Return(user1, 0, nil).Once()
	mock.On("UserGetByUsername", ctx, updateUser1.Username).Return(user2, nil).Once()
	mock.On("UserGetByEmail", ctx, updateUser1.Email).Return(user2, nil).Once()
	fields, err = s.UpdateDataUser(ctx, updateUser1, updateUser1.ID)

	assert.Equal(t, err, ErrConflict)
	assert.Equal(t, fields, []validator.InvalidField{{"username", "conflict", "", ""}, {"email", "conflict", "", ""}})

	// shows invalid errors

	// for username
	updateUser1 = &models.User{Name: "newname", Email: "user1@email2.com", Username: "invalid_name", ID: "id1"}

	mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
	fields, err = s.UpdateDataUser(ctx, updateUser1, updateUser1.ID)
	assert.Equal(t, err, ErrBadRequest)
	assert.Equal(t, fields, []validator.InvalidField{{"username", "invalid", "alphanum", ""}})

	// for email
	updateUser1 = &models.User{Name: "newname", Email: "invalid.email", Username: "newusername", ID: "id1"}
	mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
	fields, err = s.UpdateDataUser(ctx, updateUser1, updateUser1.ID)
	assert.Equal(t, err, ErrBadRequest)
	assert.Equal(t, fields, []validator.InvalidField{{"email", "invalid", "email", ""}})

	// both email and username
	updateUser1 = &models.User{Name: "newname", Email: "invalid.email", Username: "us", ID: "id1"}
	mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
	fields, err = s.UpdateDataUser(ctx, updateUser1, updateUser1.ID)
	assert.Equal(t, err, ErrBadRequest)
	assert.Equal(t, fields, []validator.InvalidField{{"email", "invalid", "email", ""}, {"username", "invalid", "min", "3"}})

	// for empty name
	updateUser1 = &models.User{Name: "", Email: "new@email.com", Username: "newusername", ID: "id1"}
	mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
	_, err = s.UpdateDataUser(ctx, updateUser1, updateUser1.ID)

	assert.Equal(t, err, ErrBadRequest)

	// successful update
	updateUser1 = &models.User{Name: "newname", Email: "user1@email2.com", Username: "newusername", ID: "id1"}

	Err = errors.New("no documents")

	mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
	mock.On("UserGetByUsername", ctx, updateUser1.Username).Return(nil, Err).Once()
	mock.On("UserGetByEmail", ctx, updateUser1.Email).Return(nil, Err).Once()
	mock.On("UserUpdateData", ctx, updateUser1, updateUser1.ID).Return(nil).Once()
	_, err = s.UpdateDataUser(ctx, updateUser1, updateUser1.ID)
	assert.Nil(t, err)

	mock.AssertExpectations(t)
}

func TestUpdatePasswordUser(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), nil, nil)

	ctx := context.TODO()

	user1 := &models.User{Name: "name", Email: "user1@email.com", Username: "username1", Password: "hash1", ID: "id1"}

	type updatePassword struct {
		currentPassword string
		newPassword     string
		expected        error
	}

	tests := []updatePassword{
		{
			"hiadoshioasc",
			"hashnew",
			ErrUnauthorized,
		},
		{
			"pass123",
			"hashnew",
			ErrUnauthorized,
		},
		{
			"askdhkasd",
			"hashnew",
			ErrUnauthorized,
		},
		{
			"pass890",
			"hashnew",
			ErrUnauthorized,
		},
		{
			"hash1",
			"hashnew",
			nil,
		},
	}

	for _, test := range tests {
		mock.On("UserGetByID", ctx, user1.ID, false).Return(user1, 0, nil).Once()
		if test.expected == nil {
			mock.On("UserUpdatePassword", ctx, test.newPassword, user1.ID).Return(nil).Once()
		}
		err := s.UpdatePasswordUser(ctx, test.currentPassword, test.newPassword, user1.ID)
		assert.Equal(t, err, test.expected)
	}

	mock.AssertExpectations(t)
}
