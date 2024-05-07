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
	// TODO: The route layer already validate this, remove it.
	if ok, err := s.validator.Struct(userData); !ok || err != nil {
		return nil, NewErrUserInvalid(nil, err)
	}

	if _, _, err := s.store.UserGetByID(ctx, id, false); err != nil {
		return nil, NewErrUserNotFound(id, nil)
	}

	if conflicts, has, _ := s.store.UserConflicts(ctx, &models.UserConflicts{Email: userData.Email, Username: userData.Username}); has {
		return conflicts, NewErrUserDuplicated(conflicts, nil)
	}

	// TODO: convert username and email to lower case.
	changes := &models.UserChanges{
		Name:     userData.Name,
		Username: userData.Username,
		Email:    userData.Email,
	}

	return nil, s.store.UserUpdate(ctx, id, changes)
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

	if err := s.store.UserUpdate(ctx, id, &models.UserChanges{Password: neo.Hash}); err != nil {
		return NewErrUserUpdate(user, err)
	}

	return nil
}
