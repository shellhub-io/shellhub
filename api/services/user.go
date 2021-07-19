package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/validator"
)

func (s *service) UpdateDataUser(ctx context.Context, data *models.User, id string) ([]validator.InvalidField, error) {
	var invalid []validator.InvalidField

	if _, _, err := s.store.UserGetByID(ctx, id, false); err != nil {
		return invalid, err
	}

	if invalidFields, err := validator.ValidateStruct(data); err != nil {
		return invalidFields, ErrBadRequest
	}

	var checkUsername, checkEmail bool

	if user, err := s.store.UserGetByUsername(ctx, data.Username); err == nil && user.ID != id {
		checkUsername = true
		invalid = append(invalid, validator.InvalidField{"username", "conflict", "", ""})
	}

	if user, err := s.store.UserGetByEmail(ctx, data.Email); err == nil && user.ID != id {
		checkEmail = true
		invalid = append(invalid, validator.InvalidField{"email", "conflict", "", ""})
	}

	if checkUsername || checkEmail {
		return invalid, ErrConflict
	}

	return invalid, s.store.UserUpdateData(ctx, data, id)
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
