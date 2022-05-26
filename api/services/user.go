package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/request"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/validator"
)

type UserService interface {
	UpdateDataUser(ctx context.Context, id string, userData request.UserDataUpdate) ([]string, error)
	UpdatePasswordUser(ctx context.Context, id string, currentPassword, newPassword string) error
}

// UpdateDataUser update user data.
//
// It receives a context, used to "control" the request flow, the user's ID, and a request.UserDataUpdate struct with
// fields to update in the models.User.
//
// It returns a slice of strings with the fields that contains data duplicated in the database, and an error.
func (s *service) UpdateDataUser(ctx context.Context, id string, userData request.UserDataUpdate) ([]string, error) {
	if _, _, err := s.store.UserGetByID(ctx, id, false); err != nil {
		return nil, NewErrUserNotFound(id, nil)
	}

	conflictFields := make([]string, 0)
	existentUser, _ := s.store.UserGetByUsername(ctx, userData.Username)
	if existentUser != nil && existentUser.ID != id {
		conflictFields = append(conflictFields, "username")
	}

	existentUser, _ = s.store.UserGetByEmail(ctx, userData.Email)
	if existentUser != nil && existentUser.ID != id {
		conflictFields = append(conflictFields, "email")
	}

	if len(conflictFields) > 0 {
		return conflictFields, NewErrUserDuplicated(conflictFields, nil)
	}

	user := models.User{
		UserData: models.UserData{
			Name:     userData.Name,
			Username: userData.Username,
			Email:    userData.Email,
		},
	}

	return nil, s.store.UserUpdateData(ctx, id, user)
}

func (s *service) UpdatePasswordUser(ctx context.Context, id, currentPassword, newPassword string) error {
	user, _, err := s.store.UserGetByID(ctx, id, false)
	if user == nil {
		return NewErrUserNotFound(id, err)
	}

	currentPassword = validator.HashPassword(currentPassword)

	if user.Password != currentPassword {
		return NewErrUserPasswordNotMatch(nil)
	}

	newPassword = validator.HashPassword(newPassword)

	return s.store.UserUpdatePassword(ctx, newPassword, id)
}
