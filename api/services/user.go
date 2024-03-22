package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type UserService interface {
	UpdateDataUser(ctx context.Context, id string, userData models.UserData) ([]string, error)
	UpdatePasswordUser(ctx context.Context, id string, currentPassword, newPassword string) error
}

// UpdateDataUser update user data.
//
// It receives a context, used to "control" the request flow, the user's ID, and a requests.UserDataUpdate struct with
// fields to update in the models.User.
//
// It returns a slice of strings with the fields that contains data duplicated in the database, and an error.
func (s *service) UpdateDataUser(ctx context.Context, id string, userData models.UserData) ([]string, error) {
	if ok, err := s.validator.Struct(userData); !ok || err != nil {
		return nil, NewErrUserInvalid(nil, err)
	}

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

	return nil, s.store.UserUpdateData(ctx, id, models.User{
		UserData: models.UserData{
			Name:     userData.Name,
			Username: userData.Username,
			Email:    userData.Email,
		},
	})
}

func (s *service) UpdatePasswordUser(ctx context.Context, id, currentPassword, newPassword string) error {
	user, _, err := s.store.UserGetByID(ctx, id, false)
	if user == nil {
		return NewErrUserNotFound(id, err)
	}

	if !user.Password.Compare(currentPassword) {
		return NewErrUserPasswordNotMatch(nil)
	}

	neo, err := models.HashUserPassword(newPassword)
	if err != nil {
		return NewErrUserPasswordInvalid(err)
	}

	if err := s.store.UserUpdatePassword(ctx, neo.Hash, id); err != nil {
		return NewErrUserUpdate(user, err)
	}

	return nil
}
