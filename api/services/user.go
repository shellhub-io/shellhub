package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/validator"
)

type UserService interface {
	UpdateDataUser(ctx context.Context, user *models.User, id string) ([]string, error)
	UpdatePasswordUser(ctx context.Context, currentPassword, newPassword, id string) error
}

func (s *service) UpdateDataUser(ctx context.Context, user *models.User, id string) ([]string, error) {
	if _, _, err := s.store.UserGetByID(ctx, id, false); err != nil {
		return nil, err
	}

	if invalidFields, err := validator.ValidateStruct(user); err != nil {
		return invalidFields, ErrBadRequest
	}

	validator.FormatUser(user)

	var conflictFields []string
	existentUser, _ := s.store.UserGetByUsername(ctx, user.Username)
	if existentUser != nil && existentUser.ID != id {
		conflictFields = append(conflictFields, "username")
	}

	existentUser, _ = s.store.UserGetByEmail(ctx, user.Email)
	if existentUser != nil && existentUser.ID != id {
		conflictFields = append(conflictFields, "email")
	}

	if len(conflictFields) > 0 {
		return conflictFields, ErrConflict
	}

	return nil, s.store.UserUpdateData(ctx, user, id)
}

func (s *service) UpdatePasswordUser(ctx context.Context, currentPassword, newPassword, id string) error {
	user, _, err := s.store.UserGetByID(ctx, id, false)
	if err != nil {
		return err
	}

	if user.Password == currentPassword {
		return s.store.UserUpdatePassword(ctx, newPassword, id)
	}

	return ErrUnauthorized
}
